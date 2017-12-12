# migrate [![build-status](https://travis-ci.org/pasztorpisti/migrate.svg?branch=master)](https://travis-ci.org/pasztorpisti/migrate)

Cross-platform SQL schema migration tool (cli).

Features:

- Supported databases:
  - postgresql
  - mysql
- Migration files have plain SQL format. Some migration parameters (like the
  `notransaction` flag) can be added to migration files as special single-line
  SQL comments. E.g.: `-- +migrate notransaction`
- Keeping forward and backward migrations either in one file or separate files
  (configurable).
- Plan command that applies migrations in "dry run" mode:
  it only prints the operations without modifying the DB.
- Squashing existing migrations to a single file (or 2 files if you store
  forward and backward migrations separately). Squashing is very useful if a DB
  schema changes frequently and accumulates hundreds of migration files quickly.

Written in golang so it can be used easily as a go migration library but that is
strongly discouraged.
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

```bash
migrate config
```

Edit the newly created `migrate.yml` file by following the instructions in it.

The template config defines two database settings - `dev` and `prod` - but
you can use any number of settings with user defined names.
Use the `-db <name>` option of the `migrate` command to select one of these settings.
If you don't use the `-db <name>` option then the default is `dev`.

### 2. Initialise migrations

Initialise the database by creating the migrations table:

```bash
migrate init
```

This has to be performed only once for a given DB. Executing it again is a no-op.

### 3. Create migration file(s)

Make sure that the `<migrations_dir>` exists.
You can define it in the config file under the `migration_source.path` key.

```bash
mkdir -p <migrations_dir>
```

Create one or more migration files with the `migrate new` command and edit them.
I create only one migration:

```bash
migrate new "initial migration"
```

The above command creates a `<migrations_dir>/0001_initial_migration.sql` file.
The filename can be different if you specify a `filename_pattern` in the config.

Edit the migration file. In my example I create two simple tables.

By default the forward and backward migrations go to the same file (with annotations).
You can store the forward and backward migrations in separate files by using a
custom `filename_pattern` in the config file. Read the related instructions in
the config file (created by the `migrate config` command).

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
