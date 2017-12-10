package migrate

import (
	"database/sql"
	"errors"
)

// Support for (fake) nested transactions.
// Only the top level transaction performs real commit/rollback on the DB.
// If a child transaction rolls back then the parents aren't allowed to commit,
// they have to roll back as well.
// Note: the current implementation isn't goroutine safe.

type Querier interface {
	Query(query string, args ...interface{}) (*sql.Rows, error)
}

type Execer interface {
	Exec(query string, args ...interface{}) (sql.Result, error)
}

type DB interface {
	Querier
	Execer
	BeginTX() (TX, error)
}

type ClosableDB interface {
	DB
	Close() error
}

type TX interface {
	DB
	TXCloser
}

type TXCloser interface {
	Commit() error
	Rollback() error
}

// stdDB is used instead of *sql.DB in a few places to make the code easier to test.
type stdDB interface {
	Querier
	Execer
	Begin() (*sql.Tx, error)
	Close() error
}

// stdTx is used instead of *sql.Tx in a few places to make the code easier to test.
type stdTx interface {
	Querier
	Execer
	TXCloser
}

func WrapDB(db stdDB) ClosableDB {
	return dbWrapper{db}
}

type dbWrapper struct {
	stdDB
}

func (o dbWrapper) BeginTX() (TX, error) {
	tx, err := o.stdDB.Begin()
	if err != nil {
		return nil, err
	}
	return wrapTx(tx), nil
}

func wrapTx(tx stdTx) TX {
	return newParentTX(&txWrapper{tx})
}

type txWrapper struct {
	stdTx
}

func (o *txWrapper) BeginTX() (TX, error) {
	return wrapTxWrapper(o), nil
}

func wrapTxWrapper(tx TX) TX {
	return newParentTX(&recursiveTXWrapper{TX: tx})
}

type recursiveTXWrapper struct {
	TX
	finished bool
}

func (o *recursiveTXWrapper) BeginTX() (TX, error) {
	return newParentTX(&recursiveTXWrapper{TX: o}), nil
}

var errCommitFinishedTX = errors.New("can't Commit a tx that has been committed or rolled back")

func (o *recursiveTXWrapper) Commit() error {
	if o.finished {
		return errCommitFinishedTX
	}
	o.finished = true
	return nil
}

var errRollbackFinishedTX = errors.New("can't Rollback a tx that has been committed or rolled back")

func (o *recursiveTXWrapper) Rollback() error {
	if o.finished {
		return errRollbackFinishedTX
	}
	o.finished = true
	return nil
}

func newParentTX(tx TX) *parentTX {
	return &parentTX{
		TX:       tx,
		children: make(map[TX]struct{}),
	}
}

type parentTX struct {
	TX
	children    map[TX]struct{}
	hasRollback bool
}

func (o *parentTX) BeginTX() (TX, error) {
	tx, err := o.TX.BeginTX()
	if err == nil {
		tx = &childTX{
			TX:     tx,
			parent: o,
		}
		o.children[tx] = struct{}{}
	}
	return tx, err
}

func (o *parentTX) childCommit(child TX) {
	delete(o.children, child)
}

func (o *parentTX) childRollback(child TX) {
	delete(o.children, child)
	o.hasRollback = true
}

var errCommitAfterChildRollback = errors.New("can't Commit a tx that has rolled back children")
var errCommitWithUnfinishedChildren = errors.New("trying to Commit a transaction that has unfinished children")

func (o *parentTX) Commit() error {
	if o.hasRollback {
		return errCommitAfterChildRollback
	}
	if len(o.children) != 0 {
		return errCommitWithUnfinishedChildren
	}
	return o.TX.Commit()
}

var errRollbackWithUnfinishedChildren = errors.New("trying to Rollback a transaction that has unfinished children")

func (o *parentTX) Rollback() error {
	if len(o.children) != 0 {
		return errRollbackWithUnfinishedChildren
	}
	return o.TX.Rollback()
}

type childTX struct {
	TX
	parent *parentTX
}

func (o *childTX) Commit() error {
	o.parent.childCommit(o)
	return o.TX.Commit()
}

func (o *childTX) Rollback() error {
	o.parent.childRollback(o)
	return o.TX.Rollback()
}
