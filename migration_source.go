package migrate

type MigrationSource interface {
	MigrationEntries(configLocation, migrationSource string) (MigrationEntries, error)
}

type MigrationEntries interface {
	NumMigrations() int
	Name(index int) string
	Steps(index int) (forward, backward Step, err error)
	IndexForName(name string) (index int, ok bool)
	AllowsPastMigrations() bool
	New(args []string) (name string, err error)
}

var GetMigrationSource = sourceRegistry.GetMigrationSource
var RegisterMigrationSource = sourceRegistry.RegisterMigrationSource

var sourceRegistry = make(sourceMap)

type sourceMap map[string]MigrationSource

func (o sourceMap) GetMigrationSource(name string) (s MigrationSource, ok bool) {
	s, ok = o[name]
	return
}

func (o sourceMap) RegisterMigrationSource(name string, s MigrationSource) {
	_, ok := o[name]
	if ok {
		panic("duplicate migration source name: " + name)
	}
	if s == nil {
		panic("migration source is nil")
	}
	o[name] = s
}
