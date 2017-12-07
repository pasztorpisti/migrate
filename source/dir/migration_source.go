package dir

import (
	"flag"
	"fmt"
	"github.com/pasztorpisti/migrate"
	"io/ioutil"
	"log"
	"net/url"
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
	parsed, err := parseMigrationSource(migrationSource)
	if err != nil {
		return nil, err
	}

	if !filepath.IsAbs(parsed.MigrationsDir) {
		if !filepath.IsAbs(configLocation) {
			a, err := filepath.Abs(configLocation)
			if err != nil {
				return nil, err
			}
			configLocation = a
		}
		parsed.MigrationsDir = filepath.Join(filepath.Dir(configLocation), parsed.MigrationsDir)
	}
	return newEntries(parsed)
}

type parsedMigrationSource struct {
	MigrationsDir       string
	UnixTimeAsNewID     bool
	MinIDLength         int
	AllowPastMigrations bool
	NoExt               bool
	Space               string
}

func parseMigrationSource(migrationSource string) (*parsedMigrationSource, error) {
	a := strings.SplitN(migrationSource, "?", 2)

	migrationsDir, err := url.PathUnescape(a[0])
	if err != nil {
		return nil, fmt.Errorf("error url-decoding migrations dir %q: %s", a[0], err)
	}

	params := ""
	if len(a) == 2 {
		params = a[1]
	}

	values, err := url.ParseQuery(params)
	if err != nil {
		return nil, fmt.Errorf("error parsing migration source params %q: %s", params, err)
	}

	p := &parsedMigrationSource{
		MigrationsDir:       migrationsDir,
		UnixTimeAsNewID:     false,
		MinIDLength:         4,
		AllowPastMigrations: false,
		NoExt:               false,
		Space:               "_",
	}

	for k, v := range values {
		if len(v) != 1 {
			return nil, fmt.Errorf("%s: expected one value, received %q", k, v)
		}
		switch k {
		case "unix_time_as_new_id":
			p.UnixTimeAsNewID, err = strconv.ParseBool(v[0])
			if err != nil {
				return nil, fmt.Errorf("unix_time_as_new_id: %v", err)
			}
		case "allow_past_migrations":
			p.AllowPastMigrations, err = strconv.ParseBool(v[0])
			if err != nil {
				return nil, fmt.Errorf("allow_past_migrations: %v", err)
			}
		case "min_id_length":
			minIDLength, err := strconv.ParseInt(v[0], 10, 0)
			if err != nil {
				return nil, fmt.Errorf("min_id_length: %v", err)
			}
			if minIDLength <= 0 {
				return nil, fmt.Errorf("min_id_length=%v has to be bigger than zero", minIDLength)
			}
			p.MinIDLength = int(minIDLength)
		case "no_ext":
			p.NoExt, err = strconv.ParseBool(v[0])
			if err != nil {
				return nil, fmt.Errorf("no_ext: %v", err)
			}
		case "space":
			p.Space = v[0]
		default:
			return nil, fmt.Errorf("invalid parameter: %q", k)
		}
	}

	return p, nil
}

func newEntries(p *parsedMigrationSource) (migrate.MigrationEntries, error) {
	st, err := os.Stat(p.MigrationsDir)
	if err != nil {
		return nil, fmt.Errorf("error reading migrations directory %q: %s", p.MigrationsDir, err)
	}
	if !st.IsDir() {
		return nil, fmt.Errorf("%q isn't a directory", p.MigrationsDir)
	}

	items, err := loadMigrationsDir(p.MigrationsDir)
	if err != nil {
		return nil, err
	}

	e := &entries{
		Params: p,
		Items:  items,
	}
	e.updateMaps()
	return e, nil
}

type entries struct {
	Params  *parsedMigrationSource
	Items   []*entry
	Names   map[string]*entry
	Indexes map[string]int
}

func (o *entries) updateMaps() {
	o.Names = make(map[string]*entry, len(o.Items)*2)
	o.Indexes = make(map[string]int, len(o.Items)*2)
	for i, item := range o.Items {
		shortName := strconv.FormatInt(item.MigrationID.Number, 10)
		o.Names[shortName] = item
		o.Names[item.MigrationID.ZeroPaddedNumber] = item
		o.Names[item.MigrationID.Name] = item
		o.Indexes[shortName] = i
		o.Indexes[item.MigrationID.ZeroPaddedNumber] = i
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

func (o *entries) AllowsPastMigrations() bool {
	return o.Params.AllowPastMigrations
}

func (o *entries) New(args []string) (name string, err error) {
	fs := flag.NewFlagSet("new", flag.ExitOnError)
	fs.Usage = func() {
		log.Print(newUsage)
		fs.PrintDefaults()
	}
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

	id := int64(1)
	if o.Params.UnixTimeAsNewID {
		id = time.Now().Unix()
	}
	if len(o.Items) > 0 {
		latestID := o.Items[len(o.Items)-1].MigrationID.Number
		if id <= latestID {
			id = latestID + 1
		}
	}

	filename := fmt.Sprintf("%.[1]*d", o.Params.MinIDLength, id)
	if description != "" {
		filename += " " + description
	}
	if !o.Params.NoExt {
		filename += ".sql"
	}
	filename = strings.Replace(filename, " ", o.Params.Space, -1)
	path := filepath.Join(o.Params.MigrationsDir, filename)

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

const newUsage = `Usage: migrate new [description]

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
