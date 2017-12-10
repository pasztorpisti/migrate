package main

import (
	"flag"
	"fmt"
	"github.com/pasztorpisti/migrate"
	"log"
	"os"
	"runtime"
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
  new       Create a new migration file.
  init      Create the migrations table in the DB if not exists.
  goto      Migrate to a specific version of the DB schema.
  plan      Print the plan that would be executed by a goto command.
  status    Print info about the current state of the migrations.
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
	"new":     cmdNew,
	"init":    cmdInit,
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
        This is either the name of a migration file or it's integer prefix.
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
        This is either the name of a migration file or it's integer prefix.
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
        This is either the name of a migration file or it's integer prefix.

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

var version string
var gitHash string
var buildDate string

func cmdVersion(opts *migrateOptions, args []string) error {
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
	return nil
}
