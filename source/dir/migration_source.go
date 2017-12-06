package dir

import (
	"flag"
	"fmt"
	"github.com/pasztorpisti/migrate"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

func init() {
	migrate.RegisterMigrationSource("dir", source{})
}

type source struct{}

func (source) MigrationEntries(configLocation, migrationSource string) (migrate.MigrationEntries, error) {
	if !filepath.IsAbs(migrationSource) {
		if !filepath.IsAbs(configLocation) {
			a, err := filepath.Abs(configLocation)
			if err != nil {
				return nil, err
			}
			configLocation = a
		}
		migrationSource = filepath.Join(filepath.Dir(configLocation), migrationSource)
	}
	return newEntries(migrationSource)
}

func newEntries(migrationsDir string) (migrate.MigrationEntries, error) {
	st, err := os.Stat(migrationsDir)
	if err != nil {
		return nil, fmt.Errorf("error reading migrations directory %q: %s", migrationsDir, err)
	}
	if !st.IsDir() {
		return nil, fmt.Errorf("%q isn't a directory", migrationsDir)
	}

	items, err := loadMigrationsDir(migrationsDir)
	if err != nil {
		return nil, err
	}

	e := &entries{
		MigrationsDir: migrationsDir,
		Items:         items,
	}
	e.updateMaps()
	return e, nil
}

type entries struct {
	MigrationsDir string
	Items         []*entry
	Names         map[string]*entry
	Indexes       map[string]int
}

func (o *entries) updateMaps() {
	o.Names = make(map[string]*entry, len(o.Items)*2)
	o.Indexes = make(map[string]int, len(o.Items)*2)
	for i, item := range o.Items {
		shortName := strconv.FormatInt(item.MigrationID.Number, 10)
		o.Names[shortName] = item
		o.Names[item.MigrationID.Name] = item
		o.Indexes[shortName] = i
		o.Indexes[item.MigrationID.Name] = i
	}
}

func (o *entries) NumMigrations() int {
	return len(o.Items)
}

func (o *entries) Name(index int) string {
	return o.Items[index].MigrationID.Name
}

func (o *entries) Steps(index int) (forward, backward migrate.Step, err error) {
	e := o.Items[index]
	m, err := loadMigrationFile(e.Filepath)
	if err != nil {
		return nil, nil, err
	}
	return m.Forward, m.Backward, nil
}

func (o *entries) IndexForName(name string) (index int, ok bool) {
	index, ok = o.Indexes[name]
	return
}

func (o *entries) New(args []string) (name string, err error) {
	fs := flag.NewFlagSet("new", flag.ExitOnError)
	fs.Usage = func() {
		log.Print(newUsage)
		fs.PrintDefaults()
	}
	space := fs.String("space", "_", "Character to be used as a safe space in migration filenames.")
	noext := fs.Bool("noext", false, "Don't append the '.sql' extension.")
	fs.Parse(args)

	if fs.NArg() > 1 {
		log.Printf("Unwanted extra arguments: %q", fs.Args()[1:])
		fs.Usage()
		os.Exit(1)
	}
	description := ""
	if fs.NArg() >= 1 {
		description = fs.Arg(0)
	}

	id := time.Now().Unix()
	if len(o.Items) > 0 {
		latestID := o.Items[len(o.Items)-1].MigrationID.Number
		if id <= latestID {
			id = latestID + 1
		}
	}

	filename := strconv.FormatInt(id, 10)
	if description != "" {
		filename += " " + description
	}
	if !*noext {
		filename += ".sql"
	}
	filename = strings.Replace(filename, " ", *space, -1)
	path := filepath.Join(o.MigrationsDir, filename)

	err = ioutil.WriteFile(path, []byte(migrationTemplate), 0644)
	if err != nil {
		return "", fmt.Errorf("error writing file %q", path)
	}

	fmt.Printf("Created %s\n", path)

	o.Items = append(o.Items, &entry{
		MigrationID: migrationID{
			Name:   filename,
			Number: id,
		},
		Filepath: path,
	})
	o.updateMaps()

	return filename, nil
}

const migrationTemplate = `-- +migrate forward

-- TODO: Implement forward migration. (required)

-- +migrate backward

-- TODO: Implement backward migration. (optional)
--       As an alternative you can delete the whole '+migrate backward'
--       block because implementing backward migration is optinoal.
`

const newUsage = `Usage: migrate new [-space <space>] [-noext] [description]

Creates a new migration file in the migration_dir specified in the config file.

The new filename is CONCATENATE(current_unix_time, space, description, ".sql").
After generating the filename spaces are replaced with '_'.
You can change '_' to somethinge else with the -space option.
You can prevent appending the ".sql" extension with the -noext option.

E.g.:
The following command: migrate new "my first migration"
Results in something like: 1512307720_my_first_migration.sql

After creation you can rename the file to whatever you like before forward
migrating it. After forward migration you mustn't rename it.
The only requirement is that it has to start with a non-negative integer
that is uniqe among your migration files.
E.g.: "0", "432134", "1.migration" and "1.sql" are all valid filenames.

You don't have to pad the numbers with leading zeros to ensure correct ordering
because sorting uses the parsed integer values instead of the filenames.

Options:
`
