package migrate

//go:generate mockgen -destination mock_0_test.go -package migrate -self_package github.com/pasztorpisti/migrate github.com/pasztorpisti/migrate DB,TX,Querier,Execer,Driver,Step,Printer
//go:generate mockgen -destination mock_1_test.go -package migrate -self_package github.com/pasztorpisti/migrate io Writer
