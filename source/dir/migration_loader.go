package dir

import (
	"errors"
	"fmt"
	"github.com/pasztorpisti/migrate"
	"io/ioutil"
	"net/url"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

const defaultFilenamePattern = "[id][description,prefix:_].sql"

type parsedMigrationSourceString struct {
	MigrationsDir       string
	AllowPastMigrations bool
	FilenamePattern     *parsedFilenamePattern
}

func parseMigrationSourceString(configLocation, migrationSource string) (*parsedMigrationSourceString, error) {
	a := strings.SplitN(migrationSource, "?", 2)

	migrationsDir, err := url.PathUnescape(a[0])
	if err != nil {
		return nil, fmt.Errorf("error url-decoding migrations dir %q: %s", a[0], err)
	}
	// ensuring that migrationsDir is absolute
	if !filepath.IsAbs(migrationsDir) {
		if !filepath.IsAbs(configLocation) {
			a, err := filepath.Abs(configLocation)
			if err != nil {
				return nil, err
			}
			configLocation = a
		}
		migrationsDir = filepath.Join(filepath.Dir(configLocation), migrationsDir)
	}

	params := ""
	if len(a) == 2 {
		params = a[1]
	}

	values, err := url.ParseQuery(params)
	if err != nil {
		return nil, fmt.Errorf("error parsing migration source params %q: %s", params, err)
	}

	p := &parsedMigrationSourceString{
		MigrationsDir:       migrationsDir,
		AllowPastMigrations: false,
	}

	filenamePattern := defaultFilenamePattern

	for k, v := range values {
		if len(v) != 1 {
			return nil, fmt.Errorf("%s: expected one value, received %q", k, v)
		}
		switch k {
		case "allow_past_migrations":
			p.AllowPastMigrations, err = strconv.ParseBool(v[0])
			if err != nil {
				return nil, fmt.Errorf("allow_past_migrations: %v", err)
			}
		case "filename_pattern":
			filenamePattern = v[0]
		default:
			return nil, fmt.Errorf("invalid parameter: %q", k)
		}
	}

	p.FilenamePattern, err = parseFilenamePattern(filenamePattern)
	if err != nil {
		return nil, fmt.Errorf("filename_pattern: %v", err)
	}

	return p, nil
}

type entry struct {
	MigrationID migrationID
	Forward     *step
	Backward    *step
}

func (o *parsedMigrationSourceString) loadMigrationsDir() ([]*entry, error) {
	files, err := ioutil.ReadDir(o.MigrationsDir)
	if err != nil {
		return nil, err
	}

	entryMap := make(map[int64]*entry, len(files))
	for _, item := range files {
		if item.IsDir() {
			continue
		}

		path := filepath.Join(o.MigrationsDir, item.Name())
		fwdSteps, backSteps, err := o.loadMigrationFile(path)
		if err != nil {
			return nil, err
		}

		for _, fwdStep := range fwdSteps {
			e, ok := entryMap[fwdStep.ParsedName.ID.Number]
			if !ok {
				entryMap[fwdStep.ParsedName.ID.Number] = &entry{
					MigrationID: fwdStep.ParsedName.ID,
					Forward:     fwdStep,
				}
				continue
			}

			if e.Forward != nil {
				return nil, fmt.Errorf("duplicate forward step - %s, %s", e.Forward, fwdStep)
			}
			e.Forward = fwdStep
		}

		for _, backStep := range backSteps {
			e, ok := entryMap[backStep.ParsedName.ID.Number]
			if !ok {
				entryMap[backStep.ParsedName.ID.Number] = &entry{
					MigrationID: backStep.ParsedName.ID,
					Backward:    backStep,
				}
				continue
			}

			if e.Backward != nil {
				return nil, fmt.Errorf("duplicate backward step - %s, %s", e.Backward, backStep)
			}
			e.Backward = backStep
		}
	}

	entryList := make([]*entry, 0, len(entryMap))
	for _, e := range entryMap {
		entryList = append(entryList, e)

		if e.Forward == nil {
			return nil, fmt.Errorf("backward migration without a forward step - %s", e.Backward)
		}
		if e.Backward != nil {
			if !e.Backward.ParsedName.equals(e.Forward.ParsedName) {
				return nil, fmt.Errorf("forward and backward migrations have different descriptions - %s, %s", e.Forward, e.Backward)
			}
		}
	}

	sort.Slice(entryList, func(i, j int) bool {
		return entryList[i].MigrationID.Number < entryList[j].MigrationID.Number
	})

	return entryList, nil
}

type step struct {
	// Path contains the absolute path to the file from which this migration
	// step has been loaded. The name of the file can be different from the
	// Name of the migration if Squashed==true.
	Path     string
	Squashed bool
	// Name is the filename of the migration.
	// If Squashed==true then Name is the name of the original file before squashing.
	Name       string
	ParsedName *parsedFilename

	// MigrateDirective is the SQL comment line that contains the
	// +migrate directive for this file. Empty string if there is no directive.
	MigrateDirective string
	Step             *migrate.SQLExecStep
}

func (o *step) String() string {
	s := o.Name
	if o.Squashed {
		s += " squashed into " + filepath.Base(o.Path)
	}
	return s
}

var migrateStepDirectiveRegex = regexp.MustCompile(`^\s*--\s*\+migrate\s+(.*?)\s*$`)
var migrateSquashedDirectiveRegex = regexp.MustCompile(`^\s*--\s*\+migrate\s+squashed\s+(.*?)\s*$`)

func (o *parsedMigrationSourceString) loadMigrationFile(path string) (forward, backward []*step, err error) {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, nil, err
	}
	lines := strings.Split(string(b), "\n")

	// Find all lines that contain the '-- +migrate squash <name>' directive.
	type squashDirective struct {
		LineIdx int
		Name    string
	}
	var squashDirectives []squashDirective
	for i, line := range lines {
		a := migrateSquashedDirectiveRegex.FindStringSubmatch(line)
		if a == nil {
			continue
		}
		name := a[1]
		squashDirectives = append(squashDirectives, squashDirective{
			LineIdx: i,
			Name:    name,
		})
	}

	if len(squashDirectives) == 0 {
		// This is a file without '+migrate squash' directives.
		// This means it has to contain exactly one '+migrate forward'
		// directive and an optional '+migrate backward'.
		fwdStep, backStep, err := o.loadStepPair(path, filepath.Base(path), lines, false)
		if err != nil {
			return nil, nil, fmt.Errorf("error parsing migration from file %q: %s", path, err)
		}
		if fwdStep != nil {
			forward = append(forward, fwdStep)
		}
		if backStep != nil {
			backward = append(backward, backStep)
		}
		return forward, backward, nil
	}

	// This file might contain several migration files squashed together
	// and marked/separated with '+migrate squashed <name>' directives.

	indexes := append(squashDirectives, squashDirective{
		LineIdx: len(lines),
	})

	for i, directive := range squashDirectives {
		squashLines := lines[directive.LineIdx+1 : indexes[i+1].LineIdx]
		if len(squashLines) > 0 && strings.TrimSpace(squashLines[0]) == "" {
			squashLines = squashLines[1:]
		}
		if i+1 < len(squashDirectives) && len(squashLines) > 0 && strings.TrimSpace(squashLines[len(squashLines)-1]) == "" {
			squashLines = squashLines[:len(squashLines)-1]
		}

		fwdStep, backStep, err := o.loadStepPair(path, directive.Name, squashLines, true)
		if err != nil {
			return nil, nil, fmt.Errorf("error parsing squashed migration %q from file %q: %s", directive.Name, path, err)
		}
		if fwdStep != nil {
			forward = append(forward, fwdStep)
		}
		if backStep != nil {
			backward = append(backward, backStep)
		}
	}

	return forward, backward, nil
}

// loadStepPair loads either a forward or a backward step, or both.
// Name is either the name of the file that contains the given lines or the
// squashed name.
//
// If the given migration is a squashed one then lines contains only those
// lines that belong to the given squashed entry.
//
// When the filename pattern contains {direction} the given lines
// don't have to contain a "+migrate" directive to specify a direction.
// If they still have a "+migrate" directive then the direction in that has
// to match the direction in the filename.
func (o *parsedMigrationSourceString) loadStepPair(path, name string, lines []string, squashed bool) (forward, backward *step, err error) {
	parsedFilename, err := o.FilenamePattern.ParseFilename(filepath.Base(path))
	if err != nil {
		return nil, nil, err
	}

	// Find all lines that contain the '-- +migrate' directive.
	type directive struct {
		LineIdx int
		Params  string
	}
	var directives []directive
	for i, line := range lines {
		m := migrateStepDirectiveRegex.FindStringSubmatch(line)
		if len(m) != 2 {
			continue
		}
		directives = append(directives, directive{
			LineIdx: i,
			Params:  m[1],
		})
	}

	newStep := func(migrateDirective string, s *migrate.SQLExecStep) (*step, error) {
		parsedName, err := o.FilenamePattern.ParseFilename(name)
		if err != nil {
			return nil, err
		}
		return &step{
			Path:             path,
			Squashed:         squashed,
			Name:             name,
			ParsedName:       parsedName,
			MigrateDirective: migrateDirective,
			Step:             s,
		}, nil
	}

	if len(directives) == 0 {
		if o.FilenamePattern.HasDirection {
			// The filename contains the migration direction so
			// the "+migrate <forward|backward>" directive is optional.
			step, err := newStep("", &migrate.SQLExecStep{
				Query: strings.Join(lines, "\n"),
			})
			if err != nil {
				return nil, nil, err
			}
			if parsedFilename.Forward {
				return step, nil, nil
			}
			return nil, step, nil
		}
		return nil, nil, errors.New("couldn't find any +migrate directives")
	}
	if len(directives) > 2 {
		return nil, nil, errors.New("too many (more than 2) +migrate directives")
	}
	indexes := append(directives, directive{
		LineIdx: len(lines),
	})

	// Processing the found directives.
	for i, d := range directives {
		fwd, notransaction, err := parseDirectiveParams(d.Params)
		if err != nil {
			return nil, nil, fmt.Errorf("error parsing +migrate directive params: %s", err)
		}

		if o.FilenamePattern.HasDirection && parsedFilename.Forward != fwd {
			return nil, nil, fmt.Errorf("directive %q conflicts with migration direction in the containing filename", "+migrate "+d.Params)
		}

		begin := indexes[i].LineIdx
		end := indexes[i+1].LineIdx
		step, err := newStep(lines[begin], &migrate.SQLExecStep{
			Query:         strings.Join(lines[begin+1:end], "\n"),
			NoTransaction: notransaction,
		})
		if err != nil {
			return nil, nil, err
		}

		if fwd {
			if forward != nil {
				return nil, nil, errors.New(`duplicate "+migrate forward" directive`)
			}
			forward = step
		} else {
			if backward != nil {
				return nil, nil, errors.New(`duplicate "+migrate backward" directive`)
			}
			backward = step
		}
	}

	return forward, backward, nil
}

// TODO: create a direction enum
// TODO: make the direction parameter optional for +migrate directives
// that are in files that define the direction in the filename.
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
