package dir

import (
	"fmt"
	"regexp"
	"strconv"
)

const maxMigrationNameLength = 255

// migrationID is the parsed form of the name of the migration file.
type migrationID struct {
	// Name is the full name of the migration file without parent directories.
	// E.g.: "0001_initial.sql" or "1.sql"
	Name string

	// ZeroPaddedNumber is the integer prefix of the file with preserved zero padding.
	// E.g.: "0001" or "1"
	ZeroPaddedNumber string

	// Number is the parsed form of the integer prefix of the file.
	Number int64
}

func (o migrationID) String() string {
	return o.Name
}

var numberPrefixPattern = regexp.MustCompile(`^(\d+).*$`)

func (o *migrationID) SetName(name string) error {
	if len(name) > maxMigrationNameLength {
		return fmt.Errorf("migration name is longer than the maximum=%v: %q", maxMigrationNameLength, name)
	}
	m := numberPrefixPattern.FindStringSubmatch(name)
	if len(m) != 2 {
		return fmt.Errorf("can't parse numeric ID from name %q", name)
	}
	n, err := strconv.ParseInt(m[1], 10, 64)
	if err != nil {
		return err
	}
	o.Name = name
	o.ZeroPaddedNumber = m[1]
	o.Number = n
	return nil
}
