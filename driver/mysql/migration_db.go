package mysql

import (
	"fmt"
	"github.com/pasztorpisti/migrate"
	"strings"
	"time"
)

type migrationDB struct {
	tableName string
}

func newMigrationDB(tableName string) (migrate.MigrationDB, error) {
	// Table names can't be interpolated in SQL statements so we
	// escape them manually and format them to the query strings.
	if strings.ContainsRune(tableName, '`') {
		return nil, fmt.Errorf("table name contains the forbidden backtick character: %q", tableName)
	}
	return &migrationDB{
		tableName: "`" + tableName + "`",
	}, nil
}

func (o *migrationDB) GetForwardMigrations(q migrate.Querier) ([]*migrate.MigrationNameAndTime, error) {
	rows, err := q.Query(`SELECT name, time FROM ` + o.tableName)
	if err != nil {
		return nil, fmt.Errorf("error querying froward migrated steps: %s", err)
	}
	defer rows.Close()

	var res []*migrate.MigrationNameAndTime
	for rows.Next() {
		var item migrate.MigrationNameAndTime
		if err := rows.Scan(&item.Name, &item.Time); err != nil {
			return nil, fmt.Errorf("error scanning forward migrations: %s", err)
		}
		res = append(res, &item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row error durig the scanning of forward migrations: %s", err)
	}
	return res, nil
}

const createTableQuery = `
CREATE TABLE IF NOT EXISTS %s (
	name VARCHAR(255) NOT NULL,
	time DATETIME NOT NULL,
	PRIMARY KEY (name)
);
`

func (o *migrationDB) CreateTableIfNotExists() (migrate.Step, error) {
	return &migrate.SQLExecStep{
		Query:    fmt.Sprintf(createTableQuery, o.tableName),
		IsSystem: true,
	}, nil
}

const forwardMigrateQuery = `INSERT INTO %s (name, time) VALUES (?, ?) ON DUPLICATE KEY UPDATE time=?;`

func (o *migrationDB) ForwardMigrate(migrationName string) (migrate.Step, error) {
	now := time.Now().UTC()
	return &migrate.SQLExecStep{
		Query:    fmt.Sprintf(forwardMigrateQuery, o.tableName),
		Args:     []interface{}{migrationName, now, now},
		IsSystem: true,
	}, nil
}

func (o *migrationDB) BackwardMigrate(migrationName string) (migrate.Step, error) {
	return &migrate.SQLExecStep{
		Query:    fmt.Sprintf(`DELETE FROM %s WHERE name = ?;`, o.tableName),
		Args:     []interface{}{migrationName},
		IsSystem: true,
	}, nil
}
