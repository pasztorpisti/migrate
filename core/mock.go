package core

//go:generate mockgen -destination mock_sql_result_test.go -package core -self_package github.com/pasztorpisti/migrate/core database/sql Result
//go:generate mockgen -destination mock_writer_test.go -package core -self_package github.com/pasztorpisti/migrate/core io Writer
//go:generate mockgen -source step.go -destination mock_step_test.go -package core
//go:generate mockgen -source db.go -destination mock_db_test.go -package core
//go:generate mockgen -source migration_db.go -destination mock_migration_db_test.go -package core
//go:generate mockgen -source printer.go -destination mock_printer_test.go -package core
