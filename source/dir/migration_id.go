package dir

import (
	"fmt"
	"regexp"
	"strconv"
)

const maxMigrationNameLength = 255

// migrationID is the parsed form of the Migration.Name field.
// migrationID.Name has the same value as Migration.Name and migrationID.Number
// is the integer parsed from the integer prefix of Name.
//
// E.g.: if Migration.Name == "42_create_my_table.sql" then
// migrationID.Name == "42_create_my_table.sql" and migrationID.Number == 42
//
// Note that the digits in the name can be terminated by any other character
// not only '_'. It can also be terminated by the end of the string.
type migrationID struct {
	Name   string
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
	o.Number = n
	return nil
}
