# migrate

Cross-platform SQL schema migration tool (cli).
Currently supports postgresql and mysql.

It is written in golang so it can be used easily as a go migration library but
that is strongly discouraged.
Running a migrate operation from a service at startup time is a bad practice for many reasons.
Instead run the commandline migration tool from the CI/CD pipeline before deploying the service.

(Partly for the above reasons) I reserve the right to change the public API of
the library but the config/migration file formats and the interface of the
commandline tool are considered to be stable.

## Installation

Download the latest stable binary release from
[the releases page](https://github.com/pasztorpisti/migrate/releases).

Run `migrate -help` for commandline options.

## Usage

### 1. Create a config file

Create a `migrate.yml` config file somewhere in your project repo:

```yaml
dev:
  db:
    # DB driver: can be postgres or mysql
    driver: postgres

    # DB driver specific connection parameters.
    # In case of mysql it looks like this: user@tcp(localhost:3306)/db_name
    #
    # Mysql data_source format: https://github.com/go-sql-driver/mysql#dsn-data-source-name
    # Postgres data_source format: https://godoc.org/github.com/lib/pq
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
    #       `migrate new` command to generate and format a new ID.
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
    #       The space:<space> parameter is used by the `migrate new` command
    #       to replace space characters of the description to something more
    #       filename friendly. E.g.: `migrate new "my first description"`
    #       would put "my_first_description" into the filename with space:_
    #
    #       Defaults: space:_ prefix: suffix:
    #
    # If you want to escape a special character (one of the []: characters) then
    # prefix it with a backtick (`). You can escape the backtick too.
    #
    # The filename_pattern setting is optional.
    # Default: '[id][description,prefix:_].sql'
    #filename_pattern: '[id][description,prefix:_].[direction,forward:fw,backward:bw].sql'

prod:
  db:
    driver: postgres
    data_source: 'postgres://service@localhost:5432/postgres'
    #migrations_table: migrations
  migration_source:
    path: migrations
    #filename_pattern: '[id][description,prefix:_].[direction,forward:fw,backward:bw].sql'
```

The above config defines two database settings: `dev` and `prod`.
You can use any number of settings with user defined names.
Use the `-db <name>` option of the `migrate` command to select one of these settings.
If you don't use the `-db <name>` option then the default is `dev`.

### 2. Initialise migrations

Initialise the database by creating a migrations table:

```bash
migrate init
```

This has to be performed only once for a given DB.

### 3. Create migration file(s)

Make sure that the migrations directory exists:

```bash
mkdir -p <migrations_dir>
```

Create an initial migration file. You can do this manually if you follow a few
very simple rules but for now let's use the commandline tool:

```bash
migrate new "initial migration"
```

The above command creates a `<migrations_dir>/0001_initial_migration.sql` file.
Edit it with an editor. In my example I create two simple tables:

```sql
-- +migrate forward

CREATE TABLE test1 (
  name TEXT NOT NULL,
  time TIMESTAMP NOT NULL,
  PRIMARY KEY (name)
);

CREATE TABLE test2 (
  id SERIAL,
  description TEXT NOT NULL,
  PRIMARY KEY (id)
);

-- +migrate backward

DROP TABLE test2;
DROP TABLE test1;
```

### 4. Forward migrate

After creating one or more migration files you can apply them to the DB:

```bash
migrate goto latest
```

## Design

You have a list of migrations with some kind of strict ordering.
This tool orders the migrations by their numeric ID: older migrations have
smaller numeric IDs.

A "bool flag" is stored in the database for each migration to remember
which one has been applied (forward migrated).
In the below example `[X]` means that the bool flag for the given migration is set:

```
$ migrate status
[X] 0001.sql
[ ] 0002.sql
[X] 0003.sql
[ ] 0004.sql
[X] 0005.sql
[X] 0006.sql
[X] 0007.sql
[ ] 0008.sql
```

This migration tool has a single most important operation: `migrate goto <target_migration>`.

The `goto` operation selects a target migration and makes sure that:

1. Migrations that are newer than the target migration and have been applied
   (have `[X]`) are backward migrated in reverse/descending order.
2. The target migration along with the migrations that are older than
   the target are forward migrated in ascending order.
   (Only those that haven't yet been applied and have `[ ]`.)

Let's see the result of a `migrate goto 0005.sql` starting from the above state:

```
$ migrate goto 0005.sql
backward-migrate 0007.sql ... OK
backward-migrate 0006.sql ... OK
forward-migrate 0002.sql ... OK
forward-migrate 0004.sql ... OK

$ migrate status
[X] 0001.sql
[X] 0002.sql
[X] 0003.sql
[X] 0004.sql
[X] 0005.sql
[ ] 0006.sql
[ ] 0007.sql
[ ] 0008.sql
```

## Just Another Migration Tool

I was looking for a simple SQL migration tool written in golang and found two
popular ones: [migrate](https://github.com/mattes/migrate)
and [sql-migrate](https://github.com/rubenv/sql-migrate).

I wanted something that works like the superb migration tool of the python-django
framework and from the above two libraries `sql-migrate` was the closest.

Initially I was thinking about bugfixing those tools but decided to roll my own
after discovering serious design issues in their migration logic.
You can fix a bug but bad design often requires a rewrite.
