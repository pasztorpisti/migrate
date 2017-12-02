package migrate

import (
	"database/sql"
	"errors"
)

// Tx supports (fakes) embedded transactions.
// Only the top level transaction performs real commit/rollback on the DB.
// If a child transaction rolls back then the parents aren't allowed to commit,
// they have to roll back as well.
// Note: the current implementation isn't goroutine safe.
type Tx interface {
	DB
	Commit() error
	Rollback() error
}

type DB interface {
	Querier
	Execer
	BeginTx() (Tx, error)
}

type Querier interface {
	Query(query string, args ...interface{}) (*sql.Rows, error)
}

type Execer interface {
	Exec(query string, args ...interface{}) (sql.Result, error)
}

func WrapDB(db *sql.DB) DB {
	return dbWrapper{db}
}

type dbWrapper struct {
	*sql.DB
}

func (o dbWrapper) BeginTx() (Tx, error) {
	tx, err := o.DB.Begin()
	if err != nil {
		return nil, err
	}
	w := &txWrapper{tx, newTxChildren()}
	return &childTxResultChecker{w, w.txChildren}, nil
}

type txWrapper struct {
	*sql.Tx
	*txChildren
}

func (o *txWrapper) BeginTx() (Tx, error) {
	w := &recursiveTxWrapper{
		Tx:         o,
		txChildren: newTxChildren(),
		parent:     o,
	}
	return &childTxResultChecker{w, w.txChildren}, nil
}

type recursiveTxWrapper struct {
	Tx
	*txChildren
	parent   parentTx
	finished bool
}

func (o *recursiveTxWrapper) BeginTx() (Tx, error) {
	w := &recursiveTxWrapper{
		Tx:         o,
		txChildren: newTxChildren(),
		parent:     o,
	}
	return &childTxResultChecker{w, w.txChildren}, nil
}

func (o *recursiveTxWrapper) Commit() error {
	if o.finished {
		return errors.New("can't Commit a tx that has been committed or rolled back")
	}
	o.finished = true
	o.parent.childCommit(o)
	return nil
}

func (o *recursiveTxWrapper) Rollback() error {
	if o.finished {
		return errors.New("can't Rollback a tx that has been committed or rolled back")
	}
	o.finished = true
	o.parent.childRollback(o)
	return nil
}

type parentTx interface {
	Tx
	childCommit(*recursiveTxWrapper)
	childRollback(*recursiveTxWrapper)
}

func newTxChildren() *txChildren {
	return &txChildren{
		children: make(map[parentTx]struct{}),
	}
}

type txChildren struct {
	hasChildRollback bool
	children         map[parentTx]struct{}
}

func (o *txChildren) childCommit(child *recursiveTxWrapper) {
	delete(o.children, child)
}

func (o *txChildren) childRollback(child *recursiveTxWrapper) {
	delete(o.children, child)
	o.hasChildRollback = true
}

type childTxResultChecker struct {
	parentTx
	children *txChildren
}

func (o *childTxResultChecker) Commit() error {
	if o.children.hasChildRollback {
		o.Rollback()
		return errors.New("can't Commit a tx that has rolled back children")
	}
	if len(o.children.children) != 0 {
		o.Rollback()
		return errors.New("committing a transaction that has unfinished children")
	}
	return o.parentTx.Commit()
}

func (o *childTxResultChecker) Rollback() error {
	err := o.parentTx.Rollback()
	if len(o.children.children) != 0 {
		return errors.New("rolling back a transaction that has unfinished children")
	}
	return err
}
