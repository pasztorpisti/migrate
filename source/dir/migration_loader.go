package dir

import (
	"errors"
	"fmt"
	"github.com/pasztorpisti/migrate"
	"io/ioutil"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

type entry struct {
	MigrationID migrationID
	// Filepath is the absolute path to the migration file.
	Filepath string
}

// loadMigrationsDir scans a migration directory and creates a sorted list of
// MigrationDirEntries without loading/parsing the migration files.
// It fails if there is a duplicate migration ID.
func loadMigrationsDir(dir string) ([]*entry, error) {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	ids := make(map[int64]struct{}, len(files))
	res := make([]*entry, 0, len(files))
	for _, item := range files {
		if item.IsDir() {
			continue
		}
		var e entry
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

// Migration represents one migration. It can be loaded from any source, for example
// memory, a directory that contains SQL files, or anything else.
type Migration struct {
	// Name is the name of the migration.
	// It has to have a valid integer prefix consisting of digits.
	// This integer prefix has to be unique in the list of migrations and the
	// order in which migrations are applied is based on the ordering of this
	// integer ID.
	//
	// You can use the ID type to parse names.
	//
	// Valid example names: "0", "123", "52.sql", "0anything", "67-drop-table",
	// "5.my.migration.sql", "42_my_migration", "42_my_migration.sql".
	Name string

	// Forward is the Step to execute when this migration has to be forward migrated.
	Forward migrate.Step

	// Backward is the Step to execute when this migration has to be backward migrated.
	// It can be nil if this migration can't be backward migrated.
	Backward migrate.Step
}

var migrateDirectivePattern = regexp.MustCompile(`^\s*--\s*\+migrate\s+(.*)$`)

func loadMigrationFile(filename string) (*Migration, error) {
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

	m := &Migration{
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
		step := &migrate.SQLExecStep{
			Query:         strings.Join(lines[begin+1:end], "\n"),
			NoTransaction: notransaction,
		}

		if forward {
			if m.Forward != nil {
				return nil, errors.New("multiple '+migrate forward' directives in the same migration")
			}
			m.Forward = step
		} else {
			if m.Backward != nil {
				return nil, errors.New("multiple '+migrate backward' directives in the same migration")
			}
			m.Backward = step
		}
	}

	return m, nil
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
