package dir

import (
	"errors"
	"flag"
	"fmt"
	"github.com/pasztorpisti/migrate"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func newEntries(src *source) (migrate.MigrationEntries, error) {
	st, err := os.Stat(src.MigrationsDir)
	if err != nil {
		return nil, fmt.Errorf("error reading migrations directory %q: %s", src.MigrationsDir, err)
	}
	if !st.IsDir() {
		return nil, fmt.Errorf("%q isn't a directory", src.MigrationsDir)
	}

	items, err := src.loadMigrationsDir()
	if err != nil {
		return nil, err
	}

	e := &entries{
		Source: src,
		Items:  items,
	}
	e.updateMaps()
	return e, nil
}

type entries struct {
	Source  *source
	Items   []*entry
	Names   map[string]*entry
	Indexes map[string]int
}

func (o *entries) updateMaps() {
	o.Names = make(map[string]*entry, len(o.Items)*2)
	o.Indexes = make(map[string]int, len(o.Items)*2)
	for i, item := range o.Items {
		for _, name := range item.MigrationID.Names {
			o.Names[name] = item
			o.Indexes[name] = i
		}
		o.Names[item.Forward.Name] = item
		o.Indexes[item.Forward.Name] = i
	}
}

func (o *entries) NumMigrations() int {
	return len(o.Items)
}

func (o *entries) Name(index int) string {
	return o.Items[index].Forward.Name
}

func (o *entries) Steps(index int) (forward, backward migrate.Step, err error) {
	e := o.Items[index]
	return e.Forward.Step, e.Backward.Step, nil
}

func (o *entries) IndexForName(name string) (index int, ok bool) {
	index, ok = o.Indexes[name]
	return
}

func (o *entries) AllowsPastMigrations() bool {
	return o.Source.AllowPastMigrations
}

const newUsageFmtStr = `Usage: migrate new [-squashed] %s

Creates a new migration file in the migration_dir specified in the config file.

The new filename is CONCATENATE(generated_numeric_id, space, description, ".sql").
After generating the filename spaces are replaced with '_'.

E.g.:
The following command: migrate new "my first migration"
Results in something like: 0001_my_first_migration.sql

After creation you can rename the file to whatever you like before forward
migrating it. After forward migration you mustn't rename it.
The only requirement is that it has to start with a non-negative integer
that is uniqe among your migration files. It can be left padded with zeros.
E.g.: "0", "00012", "432134", "1_migration" and "1.sql" are all valid filenames.

You don't have to pad the numbers with leading zeros to ensure correct migration
ordering because sorting uses the parsed integer values instead of the filenames.
However, zero padding is useful because it helps to keep your filenames sorted
when you list them (ls) in alphabetical order.

Options:
`

func (o *entries) New(args []string) (name string, err error) {
	fp := o.Source.FilenamePattern

	fs := flag.NewFlagSet("new", flag.ExitOnError)
	fs.Usage = func() {
		descriptionParam := ""
		if fp.HasDescription {
			if fp.OptionalDescription {
				descriptionParam = "[description]"
			} else {
				descriptionParam = "<description>"
			}
		}
		log.Printf(newUsageFmtStr, descriptionParam)
		fs.PrintDefaults()
	}
	squashed := fs.Bool("squashed", false, "Create a squashed migration from the existing ones.")
	fs.Parse(args)

	description := ""
	if fp.HasDescription {
		if fs.NArg() > 1 {
			log.Printf("Unwanted extra arguments: %q", fs.Args()[1:])
			fs.Usage()
			os.Exit(1)
		}
		hasDescription := fs.NArg() >= 1
		if !hasDescription && !fp.OptionalDescription {
			fs.Usage()
			os.Exit(1)
		}
		description = fs.Arg(0)
	} else {
		if fs.NArg() > 0 {
			log.Printf("Unwanted extra arguments: %q", fs.Args())
			fs.Usage()
			os.Exit(1)
		}
	}

	if !fp.OptionalDescription && description == "" {
		return "", errors.New("you have to provide a non-empty description")
	}

	if *squashed {
		return o.createSquashedMigrationFile(description)
	}

	return o.createEmptyMigrationFile(description)
}

const singleFileMigrationTemplate = `-- +migrate forward

-- TODO: Append the "notransaction" flag (without quotes) to the
--       above "+migrate forward" directive if you want to execute
--       your SQL outside of a transaction.

-- TODO: Implement forward migration. (required)

-- +migrate backward

-- TODO: Append the "notransaction" flag (without quotes) to the
--       above "+migrate backward" directive if you want to execute
--       your SQL outside of a transaction.

-- TODO: Implement backward migration. (optional)
--       As an alternative you can delete the whole '+migrate backward'
--       block because implementing backward migration is optinoal.
`

const multiFileMigrationTemplate = `-- TODO: Add the "-- +migrate notransaction" single line comment (without quotes)
--       to the header of this file if you want to execute your SQL outside of a
--       transaction.

-- TODO: add SQL statements
`

func (o *entries) createEmptyMigrationFile(description string) (name string, err error) {
	fp := o.Source.FilenamePattern

	id := int64(1)
	if !fp.IDSequence {
		id = time.Now().Unix()
	}
	if len(o.Items) > 0 {
		latestID := o.Items[len(o.Items)-1].MigrationID.Number
		if id <= latestID {
			id = latestID + 1
		}
	}

	writeFile := func(path, contents string) error {
		err = ioutil.WriteFile(path, []byte(contents), 0644)
		if err != nil {
			return fmt.Errorf("error writing file %q", path)
		}

		fmt.Printf("Created %s\n", path)
		return nil
	}

	// We don't update entries.Items because after this operation the
	// migrate tool exits anyway.

	if fp.HasDirection {
		// forward
		fwdFilename := fp.FormatFilename(id, description, true)
		fwdPath := filepath.Join(o.Source.MigrationsDir, fwdFilename)
		err := writeFile(fwdPath, multiFileMigrationTemplate)
		if err != nil {
			return "", err
		}

		// backward
		backFilename := fp.FormatFilename(id, description, false)
		backPath := filepath.Join(o.Source.MigrationsDir, backFilename)
		err = writeFile(backPath, multiFileMigrationTemplate)
		if err != nil {
			return "", err
		}

		return fwdFilename, nil
	} else {
		filename := fp.FormatFilename(id, description, false)
		path := filepath.Join(o.Source.MigrationsDir, filename)
		err := writeFile(path, singleFileMigrationTemplate)
		if err != nil {
			return "", err
		}
		return filename, nil
	}
}

func (o *entries) createSquashedMigrationFile(description string) (name string, err error) {
	if len(o.Items) == 0 {
		return "", errors.New("there is nothing to squash")
	}

	var fwdLines []string
	for i, e := range o.Items {
		if i != 0 {
			fwdLines = append(fwdLines, "")
		}
		fwdLines = append(fwdLines, "-- +migrate squashed "+e.Forward.Name, "")
		if e.Forward.MigrateDirective != "" {
			fwdLines = append(fwdLines, e.Forward.MigrateDirective)
		}
		fwdLines = append(fwdLines, e.Forward.Step.Query)
	}

	var backLines []string
	for i := len(o.Items) - 1; i >= 0; i-- {
		if i != len(o.Items)-1 {
			backLines = append(backLines, "")
		}
		e := o.Items[i]
		backLines = append(backLines, "-- +migrate squashed "+e.Backward.Name, "")
		if e.Backward.MigrateDirective != "" {
			backLines = append(backLines, e.Backward.MigrateDirective)
		}
		backLines = append(backLines, e.Backward.Step.Query)
	}

	id := o.Items[len(o.Items)-1].MigrationID.Number

	deleteOrigFiles := func() error {
		// needed to filter duplicate paths
		pathMap := make(map[string]struct{}, len(o.Items)*2)
		// contains unique path names sorted by ID
		paths := make([]string, 0, len(o.Items)*2)
		for _, e := range o.Items {
			if e.Forward != nil {
				if _, ok := pathMap[e.Forward.Path]; !ok {
					pathMap[e.Forward.Path] = struct{}{}
					paths = append(paths, e.Forward.Path)
				}
			}
			if e.Backward != nil {
				if _, ok := pathMap[e.Backward.Path]; !ok {
					pathMap[e.Backward.Path] = struct{}{}
					paths = append(paths, e.Backward.Path)
				}
			}
		}

		for _, path := range paths {
			if err := os.Remove(path); err != nil {
				return fmt.Errorf("error deleting %q: %s", path, err)
			}
		}
		return nil
	}

	if o.Source.FilenamePattern.HasDirection {
		// forward and backward migrations go to separate files
		fwdFilename := o.Source.FilenamePattern.FormatFilename(id, description, true)
		backFilename := o.Source.FilenamePattern.FormatFilename(id, description, false)

		fwdPath := filepath.Join(o.Source.MigrationsDir, fwdFilename)
		backPath := filepath.Join(o.Source.MigrationsDir, backFilename)

		fwdSquashedPath := fwdPath + ".squash.tmp"
		backSquashedPath := backPath + ".squash.tmp"

		err = ioutil.WriteFile(fwdSquashedPath, []byte(strings.Join(fwdLines, "\n")), 0644)
		if err != nil {
			return "", fmt.Errorf("error writing %q: %s", fwdSquashedPath, err)
		}

		if len(backLines) > 0 {
			err = ioutil.WriteFile(backSquashedPath, []byte(strings.Join(backLines, "\n")), 0644)
			if err != nil {
				return "", fmt.Errorf("error writing %q: %s", backSquashedPath, err)
			}
		}

		if err := deleteOrigFiles(); err != nil {
			return "", err
		}

		// renaming the temporary squashed sql files to their final name
		if err := os.Rename(fwdSquashedPath, fwdPath); err != nil {
			return "", fmt.Errorf("error renaming %q to %q: %s", fwdSquashedPath, fwdPath, err)
		}
		if len(backLines) > 0 {
			if err := os.Rename(backSquashedPath, backPath); err != nil {
				return "", fmt.Errorf("error renaming %q to %q: %s", backSquashedPath, backPath, err)
			}
		}

		return fwdFilename, nil
	}

	// forward and backward migrations go to the same file

	filename := o.Source.FilenamePattern.FormatFilename(id, description, true)
	path := filepath.Join(o.Source.MigrationsDir, filename)
	squashedPath := path + ".squash.tmp"

	contents := strings.Join(append(append(fwdLines, ""), backLines...), "\n")

	err = ioutil.WriteFile(squashedPath, []byte(contents), 0644)
	if err != nil {
		return "", fmt.Errorf("error writing %q: %s", squashedPath, err)
	}

	if err := deleteOrigFiles(); err != nil {
		return "", err
	}

	// renaming the temporary squashed sql file to its final name
	if err := os.Rename(squashedPath, path); err != nil {
		return "", fmt.Errorf("error renaming %q to %q: %s", squashedPath, path, err)
	}

	return filename, nil
}
