package migrate

type MigrationSourceFactory interface {
	NewMigrationSource(baseDir string, params map[string]string) (MigrationSource, error)
}

type MigrationSource interface {
	MigrationEntries() (MigrationEntries, error)
}

type MigrationEntries interface {
	NumMigrations() int
	Name(index int) string
	Steps(index int) (forward, backward Step, err error)
	IndexForName(name string) (index int, ok bool)
	New(args []string) (name string, err error)
}

var GetMigrationSourceFactory = sourceRegistry.GetMigrationSourceFactory
var RegisterMigrationSourceFactory = sourceRegistry.RegisterMigrationSourceFactory

var sourceRegistry = make(sourceFactoryMap)

type sourceFactoryMap map[string]MigrationSourceFactory

func (o sourceFactoryMap) GetMigrationSourceFactory(name string) (f MigrationSourceFactory, ok bool) {
	f, ok = o[name]
	return
}

func (o sourceFactoryMap) RegisterMigrationSourceFactory(name string, f MigrationSourceFactory) {
	_, ok := o[name]
	if ok {
		panic("duplicate migration source name: " + name)
	}
	if f == nil {
		panic("migration source is nil")
	}
	o[name] = f
}
