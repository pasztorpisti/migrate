package migrate

import (
	"errors"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

var migrateDirectivePattern = regexp.MustCompile(`^\s*--\s*\+migrate\s+(.*)$`)

func LoadMigrationFile(filename string) (*Migration, error) {
	b, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	lines := strings.Split(string(b), "\n")

	// Find all lines that contain the '-- +migrate' directive.
	type directive struct {
		LineIdx int
		Params  string
	}
	var directives []directive
	for i, line := range lines {
		m := migrateDirectivePattern.FindStringSubmatch(line)
		if len(m) != 2 {
			continue
		}
		directives = append(directives, directive{
			LineIdx: i,
			Params:  m[1],
		})
	}
	if len(directives) == 0 {
		return nil, errors.New("couldn't find any +migrate directives")
	}
	if len(directives) > 2 {
		return nil, errors.New("too many +migrate directives")
	}
	indexes := append(directives, directive{
		LineIdx: len(lines),
	})

	migration := &Migration{
		Name: filepath.Base(filename),
	}

	// Processing the found directives.
	for i, d := range directives {
		forward, notransaction, err := parseDirectiveParams(d.Params)
		if err != nil {
			return nil, fmt.Errorf("error parsing +migrate directive params: %s", err)
		}
		begin := indexes[i].LineIdx
		end := indexes[i+1].LineIdx
		step := &SQLExecStep{
			Query:         strings.Join(lines[begin+1:end], "\n"),
			NoTransaction: notransaction,
		}

		if forward {
			if migration.Forward != nil {
				return nil, errors.New("multiple '+migrate forward' directives in the same migration")
			}
			migration.Forward = step
		} else {
			if migration.Backward != nil {
				return nil, errors.New("multiple '+migrate backward' directives in the same migration")
			}
			migration.Backward = step
		}
	}

	return migration, nil
}

func parseDirectiveParams(params string) (forward, notransaction bool, err error) {
	forward, backward, notransaction := false, false, false
	for _, f := range strings.Split(params, " \t") {
		switch f {
		case "backward":
			if backward {
				return false, false, errors.New("duplicate backward flag")
			}
			if forward {
				return false, false, errors.New("backward and forward are exlusive")
			}
			backward = true
		case "forward":
			if forward {
				return false, false, errors.New("duplicate forward flag")
			}
			if backward {
				return false, false, errors.New("backward and forward are exlusive")
			}
			forward = true
		case "notransaction":
			if notransaction {
				return false, false, errors.New("duplicate notransaction flag")
			}
			notransaction = true
		default:
			return false, false, fmt.Errorf("invalid parameter: %q", f)
		}
	}

	if !forward && !backward {
		return false, false, errors.New("either forward or backward has to be specified")
	}
	return forward, notransaction, nil
}

type MigrationDirEntry struct {
	MigrationID MigrationID
	// Filepath is the absolute path to the migration file.
	Filepath string
}

// LoadMigrationsDir scans a migration directory and creates a sorted list of
// MigrationDirEntries without loading/parsing the migration files.
// It fails if there is a duplicate migration ID.
func LoadMigrationsDir(dir string) ([]*MigrationDirEntry, error) {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	ids := make(map[int64]struct{}, len(files))
	res := make([]*MigrationDirEntry, 0, len(files))
	for _, item := range files {
		if item.IsDir() {
			continue
		}
		var e MigrationDirEntry
		if err := e.MigrationID.SetName(item.Name()); err != nil {
			return nil, err
		}

		_, ok := ids[e.MigrationID.Number]
		if ok {
			return nil, fmt.Errorf("duplicate migration ID (%v)", e.MigrationID.Number)
		}
		ids[e.MigrationID.Number] = struct{}{}

		e.Filepath = filepath.Join(dir, item.Name())
		res = append(res, &e)
	}

	sort.Slice(res, func(i, j int) bool {
		return res[i].MigrationID.Number < res[j].MigrationID.Number
	})

	return res, nil
}

func LoadMigrationFiles(entries []*MigrationDirEntry) ([]*Migration, error) {
	res := make([]*Migration, len(entries))
	for i, e := range entries {
		m, err := LoadMigrationFile(e.Filepath)
		if err != nil {
			return nil, fmt.Errorf("error loading migration file %q: %s", e.Filepath, err)
		}
		res[i] = m
	}
	return res, nil
}
