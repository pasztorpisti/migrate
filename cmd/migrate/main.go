package main

import (
	"flag"
	"fmt"
	"github.com/pasztorpisti/migrate"
	"log"
	"os"
	"runtime"
	"sort"
	"strings"
)

const usage = `Usage: migrate [migrate_options] <command> [command_options] [command_args]

Migrate Options:
  -config <config_file>
            The location of the config file.
            Default: %s
  -db <db_config>
            Select a section from the config file.
            The <db_config> is user defined (dev, test, prod, etc...).
            Default: %s

Commands:
  config    Create a default config file if not exists.
  init      Create the migrations table in the DB if not exists.
  new       Create a new migration file or squash existing ones.
  status    Print info about the current state of the migrations.
  plan      Print the plan that would be executed by a goto command.
  goto      Migrate to a specific version of the DB schema.
  hack      Manipulate a single migration step. Useful for troubleshooting.
  version   Print version info.

Use 'migrate <command> -help' for more info about a command.
`

const (
	defaultConfigFile = "migrate.yml"
	defaultDB         = "dev"
)

type migrateOptions struct {
	ConfigFile string
	DB         string
}

var commands = map[string]func(opts *migrateOptions, args []string) error{
	"config":  cmdConfig,
	"init":    cmdInit,
	"new":     cmdNew,
	"goto":    cmdGoto,
	"plan":    cmdPlan,
	"status":  cmdStatus,
	"hack":    cmdHack,
	"version": cmdVersion,
}

func main() {
	log.SetFlags(0)

	var opts migrateOptions
	flag.Usage = func() {
		log.Printf(usage, defaultConfigFile, defaultDB)
	}
	flag.StringVar(&opts.ConfigFile, "config", defaultConfigFile, "")
	flag.StringVar(&opts.DB, "db", defaultDB, "")
	flag.Parse()

	if flag.NArg() == 0 {
		flag.Usage()
		os.Exit(1)
	}

	args := flag.Args()
	cmd := args[0]
	cmdFunc, ok := commands[cmd]
	if !ok {
		log.Printf("Invalid command: %s", cmd)
		flag.Usage()
		os.Exit(1)
	}

	err := cmdFunc(&opts, args[1:])
	if err != nil {
		log.Print(err)
		os.Exit(1)
	}
}

var stdoutPrinter = migrate.NewPrinter(os.Stdout)

const configTemplate = `dev:
  db:
    # DB driver: can be postgres or mysql
    driver: postgres

    # DB driver specific connection parameters.
    # In case of mysql it looks like this: user@tcp(localhost:3306)/db_name
    #
    # Mysql data_source format: https://github.com/go-sql-driver/mysql#dsn-data-source-name
    # Postgres data_source format: https://godoc.org/github.com/lib/pq
    #
    # You can interpolate environment variables by using {env:ENV_VAR_NAME}
    # placeholders. Outside of the placeholders you have to escape/prefix the
    # '{' and '\' characters with a backslash. Inside the placeholders you
    # have to escape the ':', '}' and '\' characters when you want to lose
    # their special meaning. Placeholders can't be nested.
    data_source: 'postgres://steve@localhost:5432/postgres?sslmode=disable'

    # The name of the migrations table. Optional, default value: migrations
    #migrations_table: migrations

  migration_source:
    # The relative or absolute path to the directory that contains the migration files.
    # A relative path is relative to the parent dir of this config file.
    path: migrations

    # The filename pattern of the SQL migration files. It is a template string
    # with a few placeholders. Each placeholder consists of a name and
    # zero or more key:value pairs. E.g.: '[id]' or '[id,key1=val1,key2=val2]'
    #
    # You can use the following placeholders:
    #
    # [id,generate:<type>,width:<width>]
    #
    #       The numeric ID of the migration file. The parameters are used by the
    #       ` + "`" + `migrate new` + "`" + ` command to generate and format a new ID.
    #
    #       The generate:<type> parameter can be generate:sequence or
    #       generate:unix_time. Default: generate:sequence
    #
    #       The <width> parameter can be a number between 1 and 50. Default: 4
    #       It controls the zero padding of the ID. Zero padding isn't needed
    #       by the migrate tool but it looks better especially when your tools
    #       list the migration files in alphabetical order.
    #
    #       The id placeholder is required.
    #
    # [direction,forward:<forward>,backward:<backward>]
    #
    #       The direction of the migration formatted into the filename.
    #       This placeholder is optional. If you put it to the filename pattern
    #       then your forward and backward migrations will be split into separate
    #       files. Not using this placeholder means that the backward and
    #       forward part of a migration go to a single file.
    #
    #       The forward:<forward> parameter defines what string to put into
    #       the name of forward migration files. The backward:<backward> parameter
    #       does the same for backward migration files.
    #       Defaults: forward:forward backward:backward
    #
    # [description,space:<space>,prefix:<prefix>,suffix:<suffix>]
    #
    #       The description placeholder is optional. If you leave it out from
    #       the filename pattern then you won't be able to add description into
    #       the names of your migration files.
    #
    #       If you use the description placeholder without the prefix:<prefix>
    #       and suffix:<suffix> parameters then the description in your migration
    #       filenames will be required and has to be at least 1 character long.
    #
    #       If you specify at least one of the prefix or suffix parameters then
    #       the description is optional in your filenames. The prefix and
    #       suffix are glued to the description only when it is present.
    #
    #       The space:<space> parameter is used by the ` + "`" + `migrate new` + "`" + ` command
    #       to replace space characters of the description to something more
    #       filename friendly. E.g.: ` + "`" + `migrate new "my first description"` + "`" + `
    #       would put "my_first_description" into the filename with space:_
    #
    #       Defaults: space:_ prefix: suffix:
    #
    # If you want to escape a special character (one of the []: characters) then
    # prefix it with a backtick (` + "`" + `). You can escape the backtick too.
    #
    # It doesn't have to have .sql extension but adding it might be useful if
    # your editor/IDE uses the extension to determine the file type.
    # Optional. Default: '[id][description,prefix:_].sql'
    #filename_pattern: '[id][description,prefix:_].[direction,forward:fw,backward:bw].sql'

prod:
  db:
    driver: postgres
    data_source: 'postgres://{env:DB_USER}:{env:DB_PASSWORD}@{env:DB_HOST}:5432/postgres'
    #migrations_table: migrations
  migration_source:
    path: migrations
    #filename_pattern: '[id][description,prefix:_].[direction,forward:fw,backward:bw].sql'

# TODO: copy-paste the above 'prod' DB settings as many times as you wish and
# always rename the root (from 'prod' to something else, e.g.: 'staging', 'dev2').
# You can refer to one of these blocks using the -db option of the migrate
# command which uses '-db dev' as a default.
`

const configUsage = `Usage: migrate config [filename]

Creates a default config file with help/instrutions in it.
Does nothing and returns with error if the given file already exists.

The default [filename] is 'migrate.yml'. If you create the config with a
different filename then you always have run the ` + "`" + `migrate` + "`" + ` command
with the ` + "`" + `-config <your_config_filename>` + "`" + ` option.

Use '-' as the [filename] to print the config template to stdout.
`

func cmdConfig(opts *migrateOptions, args []string) error {
	fs := flag.NewFlagSet("config", flag.ExitOnError)
	fs.Usage = func() {
		log.Print(configUsage)
		fs.PrintDefaults()
	}
	fs.Parse(args)

	if fs.NArg() > 1 {
		log.Printf("Unwanted extra arguments: %q", fs.Args())
		fs.Usage()
		os.Exit(1)
	}

	fn := "migrate.yml"
	if fs.NArg() >= 1 {
		fn = fs.Arg(0)
	}

	if fn == "-" {
		fmt.Print(configTemplate)
		return nil
	}

	f, err := os.OpenFile(fn, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0644)
	if err != nil {
		return fmt.Errorf("Error creating config file %q: %s", fn, err.(*os.PathError).Err)
	}
	_, err = f.Write([]byte(configTemplate))
	if err2 := f.Close(); err == nil {
		err = err2
	}
	if err != nil {
		return fmt.Errorf("Error writing config file %q: %s", fn, err)
	}

	fmt.Println("Created migrate config file: " + fn)
	return nil
}

func cmdNew(opts *migrateOptions, args []string) error {
	return migrate.CmdNew(&migrate.CmdNewInput{
		ConfigFile: opts.ConfigFile,
		DB:         opts.DB,
		Args:       args,
	})
}

const initUsage = `Usage: migrate init

Creates the migration table if it hasn't yet been created.
Issuing an init command on an already initialised DB is a harmless no-op.
`

func cmdInit(opts *migrateOptions, args []string) error {
	fs := flag.NewFlagSet("init", flag.ExitOnError)
	fs.Usage = func() {
		log.Print(initUsage)
		fs.PrintDefaults()
	}
	fs.Parse(args)

	if fs.NArg() != 0 {
		log.Printf("Unwanted extra arguments: %q", fs.Args())
		fs.Usage()
		os.Exit(1)
	}

	return migrate.CmdInit(&migrate.CmdInitInput{
		Output:     stdoutPrinter,
		ConfigFile: opts.ConfigFile,
		DB:         opts.DB,
	})
}

const gotoUsage = `Usage: migrate goto [-quiet] <migration_id>

Backward migrate everything that is newer than <migration_id> and
forward migrate <migration_id> along with everything that is older.

Options:
`

const gotoUsageArgs = `
Args:
  <migration_id>
        This is either the name of a forward migration file or its
        numeric id (with or without zero prefix).
        It can also be one of the following special values:

        initial      Backward migrates everything.
        latest       Forward migrates everything.
`

func cmdGoto(opts *migrateOptions, args []string) error {
	fs := flag.NewFlagSet("goto", flag.ExitOnError)
	fs.Usage = func() {
		log.Print(gotoUsage)
		fs.PrintDefaults()
		log.Print(gotoUsageArgs)
	}
	quiet := fs.Bool("quiet", false, "Don't log migration steps.")
	fs.Parse(args)

	if fs.NArg() != 1 {
		if fs.NArg() > 1 {
			log.Printf("Unwanted extra arguments: %q", fs.Args()[1:])
		}
		fs.Usage()
		os.Exit(1)
	}
	migrationID := fs.Arg(0)

	return migrate.CmdGoto(&migrate.CmdGotoInput{
		Output:      stdoutPrinter,
		ConfigFile:  opts.ConfigFile,
		DB:          opts.DB,
		MigrationID: migrationID,
		Quiet:       *quiet,
	})
}

const planUsage = `Usage: migrate plan [-sql] [-sys] <migration_id>

Print a plan without modifying the database.

Options:
`

const planUsageArgs = `
Args:
  <migration_id>
        This is either the name of a forward migration file or its
        numeric id (with or without zero prefix).
        It can also be one of the following special values:

        initial      Backward migrates everything.
        latest       Forward migrates everything.

`

func cmdPlan(opts *migrateOptions, args []string) error {
	fs := flag.NewFlagSet("plan", flag.ExitOnError)
	fs.Usage = func() {
		log.Print(planUsage)
		fs.PrintDefaults()
		log.Print(planUsageArgs)
	}
	sql := fs.Bool("sql", false, "Log the migration SQL statements (those that modify user tables).")
	sys := fs.Bool("sys", false, "Log all SQL statements including those that modify the migrations table. Implies -sql.")
	fs.Parse(args)

	if fs.NArg() != 1 {
		if fs.NArg() > 1 {
			log.Printf("Unwanted extra arguments: %q", fs.Args()[1:])
		}
		fs.Usage()
		os.Exit(1)
	}
	migrationID := fs.Arg(0)

	return migrate.CmdPlan(&migrate.CmdPlanInput{
		Output:         stdoutPrinter,
		ConfigFile:     opts.ConfigFile,
		DB:             opts.DB,
		MigrationID:    migrationID,
		PrintSQL:       *sql,
		PrintSystemSQL: *sys,
	})
}

const statusUsage = `Usage: migrate status

Print the status of the migrations.
`

func cmdStatus(opts *migrateOptions, args []string) error {
	fs := flag.NewFlagSet("status", flag.ExitOnError)
	fs.Usage = func() {
		log.Print(statusUsage)
	}
	fs.Parse(args)

	if fs.NArg() != 0 {
		log.Printf("Unwanted extra arguments: %q", fs.Args())
		fs.Usage()
		os.Exit(1)
	}

	return migrate.CmdStatus(&migrate.CmdStatusInput{
		Output:     stdoutPrinter,
		ConfigFile: opts.ConfigFile,
		DB:         opts.DB,
	})
}

const hackUsage = `Usage: migrate hack [-force] [-useronly|-sysonly] <forward|backward> <migration_id>

Forward- or backward-migrate a single step specified by <migration_id>.
Useful for troubleshooting.

It executes two sets of SQL statements:

1. SQL that forward or backward migrates user tables.
2. SQL that updates the migrations table (system data) used by this tool.

You can use the -useronly or -sysonly options to execute only one of these.

Options:
`

const hackUsageArgs = `
Args:
  <forward|backward>
        Select the direction in which the given <migration_id> will be migrated.

  <migration_id>
        This is either the name of a forward migration file or its
        numeric id (with or without zero prefix).

`

func cmdHack(opts *migrateOptions, args []string) error {
	fs := flag.NewFlagSet("hack", flag.ExitOnError)
	fs.Usage = func() {
		log.Print(hackUsage)
		fs.PrintDefaults()
		log.Print(hackUsageArgs)
	}
	force := fs.Bool("force", false, "Skip checking the migration system data for the current state of <migration_id>.")
	useronly := fs.Bool("useronly", false, "Skip the execution of SQL that modifies the migrations table.")
	sysonly := fs.Bool("sysonly", false, "Skip the execution of SQL that modifies the user tables.")
	fs.Parse(args)

	if fs.NArg() != 2 {
		if fs.NArg() > 2 {
			log.Printf("Unwanted extra arguments: %q", fs.Args()[2:])
		}
		fs.Usage()
		os.Exit(1)
	}
	direction := fs.Arg(0)
	migrationID := fs.Arg(1)

	var forward bool
	switch direction {
	case "forward":
		forward = true
	case "backward":
		forward = false
	default:
		log.Printf("Direction has to be forward or backward, got %q", direction)
		fs.Usage()
		os.Exit(1)
	}

	if *useronly && *sysonly {
		log.Print("The -useronly and -sysonly options are exclusive.")
		fs.Usage()
		os.Exit(1)
	}

	return migrate.CmdHack(&migrate.CmdHackInput{
		Output:      stdoutPrinter,
		ConfigFile:  opts.ConfigFile,
		DB:          opts.DB,
		Forward:     forward,
		MigrationID: migrationID,
		Force:       *force,
		UserOnly:    *useronly,
		SystemOnly:  *sysonly,
	})
}

const versionUsage = `Usage: migrate version

Shows the version and build information.
`

var version string
var gitHash string
var buildDate string

func cmdVersion(opts *migrateOptions, args []string) error {
	fs := flag.NewFlagSet("version", flag.ExitOnError)
	fs.Usage = func() {
		log.Print(versionUsage)
	}
	fs.Parse(args)

	if fs.NArg() != 0 {
		log.Printf("Unwanted extra arguments: %q", fs.Args())
		fs.Usage()
		os.Exit(1)
	}

	if version == "" {
		version = "dev"
	}
	fmt.Printf("version     : %s\n", version)
	if buildDate != "" {
		fmt.Printf("build date  : %s\n", buildDate)
	}
	if gitHash != "" {
		fmt.Printf("git hash    : %s\n", gitHash)
	}
	fmt.Printf("go version  : %s\n", runtime.Version())
	fmt.Printf("go compiler : %s\n", runtime.Compiler)
	fmt.Printf("platform    : %s/%s\n", runtime.GOOS, runtime.GOARCH)

	drivers := migrate.SupportedDrivers()
	sort.Strings(drivers)
	fmt.Printf("db drivers  : %s\n", strings.Join(drivers, ", "))
	return nil
}
