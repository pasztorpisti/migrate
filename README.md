# migrate

Cross-platform SQL migration tool (cli).
Currently supports postgresql and mysql.

It is written in golang so it can be used easily as a go migration library but
that is strongly discouraged.
Running a migrate operation from a service at startup time is a bad practice for many reasons.
Instead run the commandline migration tool from the CI/CD pipeline before deploying the service.

(Partly for the above reasons) I reserve the right to change the public API of
the library but the config/migration file formats and the interface of the
commandline tool are considered to be stable.

## Installation

```bash
go get -u github.com/pasztorpisti/migrate/cmd/...
migrate -help
```

## Usage

### 1. Create a config file

Create a `migrate.yml` config file somewhere in your project repo:

```yaml
dev:
  # DB driver: can be postgres or mysql
  driver: postgres

  # DB driver specific connection parameters.
  data_source: user=steve host=localhost port=5432 dbname=postgres sslmode=disable

  # Currently the only supported source type is dir://.
  # Absolute directory: dir:///dir1/dir2
  # Directory relative to the parent dir of this config: dir://dir1/dir2
  migration_source: dir://migrations

prod:
  driver: postgres
  data_source: user=service host=localhost port=5432 dbname=postgres
  migration_source: dir://migrations
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

The above command creates a file like `<migrations_dir>/1512347381_initial_migration.sql`.
Open and edit the file. In my example I create two simple tables:

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

This tool stores a "bool flag" in the database for each migration to remember
which one has been applied (forward migrated). In the below example `[X]` means
that the bool flag for the given migration is set (forward migrated):

```
$ migrate status
[X] 1.sql
[ ] 2.sql
[X] 3.sql
[ ] 4.sql
[X] 5.sql
[X] 6.sql
[X] 7.sql
[ ] 8.sql
```

This migration tool has a single most important operation: `migrate goto <target_migration>`.

The `goto` operation selects a target migration and makes sure that:

1. Migrations that are newer than the target migration and have been applied
   (have `[X]`) are backward migrated in reverse/descending order.
2. The target migration along with the migrations that are older than
   the target are forward migrated in ascending order.
   (Only those that haven't yet been applied and have `[ ]`.)

Let's see the result of a `migrate goto 5` starting from the above state:

```
$ migrate goto 5
Backward-migrate 7.sql
Backward-migrate 6.sql
Forward-migrate 2.sql
Forward-migrate 4.sql

$ migrate status
[X] 1.sql
[X] 2.sql
[X] 3.sql
[X] 4.sql
[X] 5.sql
[ ] 6.sql
[ ] 7.sql
[ ] 8.sql
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
