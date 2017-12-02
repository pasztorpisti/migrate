package migrate

import (
	"fmt"
	"regexp"
	"strconv"
)

const MaxMigrationNameLength = 255

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
	Forward Step

	// Backward is the Step to execute when this migration has to be backward migrated.
	// It can be nil if this migration can't be backward migrated.
	Backward Step
}

// migrationID is the parsed form of the Migration.Name field.
// migrationID.Name has the same value as Migration.Name and migrationID.Number
// is the integer parsed from the integer prefix of Name.
//
// E.g.: if Migration.Name == "42_create_my_table.sql" then
// migrationID.Name == "42_create_my_table.sql" and migrationID.Number == 42
//
// Note that the digits in the name can be terminated by any other character
// not only '_'. It can also be terminated by the end of the string.
type MigrationID struct {
	Name   string
	Number int64
}

func (o MigrationID) String() string {
	return o.Name
}

var numberPrefixPattern = regexp.MustCompile(`^(\d+).*$`)

func (o *MigrationID) SetName(name string) error {
	if len(name) > MaxMigrationNameLength {
		return fmt.Errorf("migration name is longer than the maximum=%v: %q", MaxMigrationNameLength, name)
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
