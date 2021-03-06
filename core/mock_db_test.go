// Code generated by MockGen. DO NOT EDIT.
// Source: db.go

// Package core is a generated GoMock package.
package core

import (
	sql "database/sql"
	gomock "github.com/golang/mock/gomock"
	reflect "reflect"
)

// MockQuerier is a mock of Querier interface
type MockQuerier struct {
	ctrl     *gomock.Controller
	recorder *MockQuerierMockRecorder
}

// MockQuerierMockRecorder is the mock recorder for MockQuerier
type MockQuerierMockRecorder struct {
	mock *MockQuerier
}

// NewMockQuerier creates a new mock instance
func NewMockQuerier(ctrl *gomock.Controller) *MockQuerier {
	mock := &MockQuerier{ctrl: ctrl}
	mock.recorder = &MockQuerierMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockQuerier) EXPECT() *MockQuerierMockRecorder {
	return m.recorder
}

// Query mocks base method
func (m *MockQuerier) Query(query string, args ...interface{}) (*sql.Rows, error) {
	varargs := []interface{}{query}
	for _, a := range args {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "Query", varargs...)
	ret0, _ := ret[0].(*sql.Rows)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Query indicates an expected call of Query
func (mr *MockQuerierMockRecorder) Query(query interface{}, args ...interface{}) *gomock.Call {
	varargs := append([]interface{}{query}, args...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Query", reflect.TypeOf((*MockQuerier)(nil).Query), varargs...)
}

// MockExecer is a mock of Execer interface
type MockExecer struct {
	ctrl     *gomock.Controller
	recorder *MockExecerMockRecorder
}

// MockExecerMockRecorder is the mock recorder for MockExecer
type MockExecerMockRecorder struct {
	mock *MockExecer
}

// NewMockExecer creates a new mock instance
func NewMockExecer(ctrl *gomock.Controller) *MockExecer {
	mock := &MockExecer{ctrl: ctrl}
	mock.recorder = &MockExecerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockExecer) EXPECT() *MockExecerMockRecorder {
	return m.recorder
}

// Exec mocks base method
func (m *MockExecer) Exec(query string, args ...interface{}) (sql.Result, error) {
	varargs := []interface{}{query}
	for _, a := range args {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "Exec", varargs...)
	ret0, _ := ret[0].(sql.Result)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Exec indicates an expected call of Exec
func (mr *MockExecerMockRecorder) Exec(query interface{}, args ...interface{}) *gomock.Call {
	varargs := append([]interface{}{query}, args...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Exec", reflect.TypeOf((*MockExecer)(nil).Exec), varargs...)
}

// MockDB is a mock of DB interface
type MockDB struct {
	ctrl     *gomock.Controller
	recorder *MockDBMockRecorder
}

// MockDBMockRecorder is the mock recorder for MockDB
type MockDBMockRecorder struct {
	mock *MockDB
}

// NewMockDB creates a new mock instance
func NewMockDB(ctrl *gomock.Controller) *MockDB {
	mock := &MockDB{ctrl: ctrl}
	mock.recorder = &MockDBMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockDB) EXPECT() *MockDBMockRecorder {
	return m.recorder
}

// Query mocks base method
func (m *MockDB) Query(query string, args ...interface{}) (*sql.Rows, error) {
	varargs := []interface{}{query}
	for _, a := range args {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "Query", varargs...)
	ret0, _ := ret[0].(*sql.Rows)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Query indicates an expected call of Query
func (mr *MockDBMockRecorder) Query(query interface{}, args ...interface{}) *gomock.Call {
	varargs := append([]interface{}{query}, args...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Query", reflect.TypeOf((*MockDB)(nil).Query), varargs...)
}

// Exec mocks base method
func (m *MockDB) Exec(query string, args ...interface{}) (sql.Result, error) {
	varargs := []interface{}{query}
	for _, a := range args {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "Exec", varargs...)
	ret0, _ := ret[0].(sql.Result)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Exec indicates an expected call of Exec
func (mr *MockDBMockRecorder) Exec(query interface{}, args ...interface{}) *gomock.Call {
	varargs := append([]interface{}{query}, args...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Exec", reflect.TypeOf((*MockDB)(nil).Exec), varargs...)
}

// BeginTX mocks base method
func (m *MockDB) BeginTX() (TX, error) {
	ret := m.ctrl.Call(m, "BeginTX")
	ret0, _ := ret[0].(TX)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// BeginTX indicates an expected call of BeginTX
func (mr *MockDBMockRecorder) BeginTX() *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "BeginTX", reflect.TypeOf((*MockDB)(nil).BeginTX))
}

// MockClosableDB is a mock of ClosableDB interface
type MockClosableDB struct {
	ctrl     *gomock.Controller
	recorder *MockClosableDBMockRecorder
}

// MockClosableDBMockRecorder is the mock recorder for MockClosableDB
type MockClosableDBMockRecorder struct {
	mock *MockClosableDB
}

// NewMockClosableDB creates a new mock instance
func NewMockClosableDB(ctrl *gomock.Controller) *MockClosableDB {
	mock := &MockClosableDB{ctrl: ctrl}
	mock.recorder = &MockClosableDBMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockClosableDB) EXPECT() *MockClosableDBMockRecorder {
	return m.recorder
}

// Query mocks base method
func (m *MockClosableDB) Query(query string, args ...interface{}) (*sql.Rows, error) {
	varargs := []interface{}{query}
	for _, a := range args {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "Query", varargs...)
	ret0, _ := ret[0].(*sql.Rows)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Query indicates an expected call of Query
func (mr *MockClosableDBMockRecorder) Query(query interface{}, args ...interface{}) *gomock.Call {
	varargs := append([]interface{}{query}, args...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Query", reflect.TypeOf((*MockClosableDB)(nil).Query), varargs...)
}

// Exec mocks base method
func (m *MockClosableDB) Exec(query string, args ...interface{}) (sql.Result, error) {
	varargs := []interface{}{query}
	for _, a := range args {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "Exec", varargs...)
	ret0, _ := ret[0].(sql.Result)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Exec indicates an expected call of Exec
func (mr *MockClosableDBMockRecorder) Exec(query interface{}, args ...interface{}) *gomock.Call {
	varargs := append([]interface{}{query}, args...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Exec", reflect.TypeOf((*MockClosableDB)(nil).Exec), varargs...)
}

// BeginTX mocks base method
func (m *MockClosableDB) BeginTX() (TX, error) {
	ret := m.ctrl.Call(m, "BeginTX")
	ret0, _ := ret[0].(TX)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// BeginTX indicates an expected call of BeginTX
func (mr *MockClosableDBMockRecorder) BeginTX() *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "BeginTX", reflect.TypeOf((*MockClosableDB)(nil).BeginTX))
}

// Close mocks base method
func (m *MockClosableDB) Close() error {
	ret := m.ctrl.Call(m, "Close")
	ret0, _ := ret[0].(error)
	return ret0
}

// Close indicates an expected call of Close
func (mr *MockClosableDBMockRecorder) Close() *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Close", reflect.TypeOf((*MockClosableDB)(nil).Close))
}

// MockTX is a mock of TX interface
type MockTX struct {
	ctrl     *gomock.Controller
	recorder *MockTXMockRecorder
}

// MockTXMockRecorder is the mock recorder for MockTX
type MockTXMockRecorder struct {
	mock *MockTX
}

// NewMockTX creates a new mock instance
func NewMockTX(ctrl *gomock.Controller) *MockTX {
	mock := &MockTX{ctrl: ctrl}
	mock.recorder = &MockTXMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockTX) EXPECT() *MockTXMockRecorder {
	return m.recorder
}

// Query mocks base method
func (m *MockTX) Query(query string, args ...interface{}) (*sql.Rows, error) {
	varargs := []interface{}{query}
	for _, a := range args {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "Query", varargs...)
	ret0, _ := ret[0].(*sql.Rows)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Query indicates an expected call of Query
func (mr *MockTXMockRecorder) Query(query interface{}, args ...interface{}) *gomock.Call {
	varargs := append([]interface{}{query}, args...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Query", reflect.TypeOf((*MockTX)(nil).Query), varargs...)
}

// Exec mocks base method
func (m *MockTX) Exec(query string, args ...interface{}) (sql.Result, error) {
	varargs := []interface{}{query}
	for _, a := range args {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "Exec", varargs...)
	ret0, _ := ret[0].(sql.Result)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Exec indicates an expected call of Exec
func (mr *MockTXMockRecorder) Exec(query interface{}, args ...interface{}) *gomock.Call {
	varargs := append([]interface{}{query}, args...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Exec", reflect.TypeOf((*MockTX)(nil).Exec), varargs...)
}

// BeginTX mocks base method
func (m *MockTX) BeginTX() (TX, error) {
	ret := m.ctrl.Call(m, "BeginTX")
	ret0, _ := ret[0].(TX)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// BeginTX indicates an expected call of BeginTX
func (mr *MockTXMockRecorder) BeginTX() *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "BeginTX", reflect.TypeOf((*MockTX)(nil).BeginTX))
}

// Commit mocks base method
func (m *MockTX) Commit() error {
	ret := m.ctrl.Call(m, "Commit")
	ret0, _ := ret[0].(error)
	return ret0
}

// Commit indicates an expected call of Commit
func (mr *MockTXMockRecorder) Commit() *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Commit", reflect.TypeOf((*MockTX)(nil).Commit))
}

// Rollback mocks base method
func (m *MockTX) Rollback() error {
	ret := m.ctrl.Call(m, "Rollback")
	ret0, _ := ret[0].(error)
	return ret0
}

// Rollback indicates an expected call of Rollback
func (mr *MockTXMockRecorder) Rollback() *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Rollback", reflect.TypeOf((*MockTX)(nil).Rollback))
}

// MockTXCloser is a mock of TXCloser interface
type MockTXCloser struct {
	ctrl     *gomock.Controller
	recorder *MockTXCloserMockRecorder
}

// MockTXCloserMockRecorder is the mock recorder for MockTXCloser
type MockTXCloserMockRecorder struct {
	mock *MockTXCloser
}

// NewMockTXCloser creates a new mock instance
func NewMockTXCloser(ctrl *gomock.Controller) *MockTXCloser {
	mock := &MockTXCloser{ctrl: ctrl}
	mock.recorder = &MockTXCloserMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockTXCloser) EXPECT() *MockTXCloserMockRecorder {
	return m.recorder
}

// Commit mocks base method
func (m *MockTXCloser) Commit() error {
	ret := m.ctrl.Call(m, "Commit")
	ret0, _ := ret[0].(error)
	return ret0
}

// Commit indicates an expected call of Commit
func (mr *MockTXCloserMockRecorder) Commit() *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Commit", reflect.TypeOf((*MockTXCloser)(nil).Commit))
}

// Rollback mocks base method
func (m *MockTXCloser) Rollback() error {
	ret := m.ctrl.Call(m, "Rollback")
	ret0, _ := ret[0].(error)
	return ret0
}

// Rollback indicates an expected call of Rollback
func (mr *MockTXCloserMockRecorder) Rollback() *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Rollback", reflect.TypeOf((*MockTXCloser)(nil).Rollback))
}

// MockstdDB is a mock of stdDB interface
type MockstdDB struct {
	ctrl     *gomock.Controller
	recorder *MockstdDBMockRecorder
}

// MockstdDBMockRecorder is the mock recorder for MockstdDB
type MockstdDBMockRecorder struct {
	mock *MockstdDB
}

// NewMockstdDB creates a new mock instance
func NewMockstdDB(ctrl *gomock.Controller) *MockstdDB {
	mock := &MockstdDB{ctrl: ctrl}
	mock.recorder = &MockstdDBMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockstdDB) EXPECT() *MockstdDBMockRecorder {
	return m.recorder
}

// Query mocks base method
func (m *MockstdDB) Query(query string, args ...interface{}) (*sql.Rows, error) {
	varargs := []interface{}{query}
	for _, a := range args {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "Query", varargs...)
	ret0, _ := ret[0].(*sql.Rows)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Query indicates an expected call of Query
func (mr *MockstdDBMockRecorder) Query(query interface{}, args ...interface{}) *gomock.Call {
	varargs := append([]interface{}{query}, args...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Query", reflect.TypeOf((*MockstdDB)(nil).Query), varargs...)
}

// Exec mocks base method
func (m *MockstdDB) Exec(query string, args ...interface{}) (sql.Result, error) {
	varargs := []interface{}{query}
	for _, a := range args {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "Exec", varargs...)
	ret0, _ := ret[0].(sql.Result)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Exec indicates an expected call of Exec
func (mr *MockstdDBMockRecorder) Exec(query interface{}, args ...interface{}) *gomock.Call {
	varargs := append([]interface{}{query}, args...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Exec", reflect.TypeOf((*MockstdDB)(nil).Exec), varargs...)
}

// Begin mocks base method
func (m *MockstdDB) Begin() (*sql.Tx, error) {
	ret := m.ctrl.Call(m, "Begin")
	ret0, _ := ret[0].(*sql.Tx)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Begin indicates an expected call of Begin
func (mr *MockstdDBMockRecorder) Begin() *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Begin", reflect.TypeOf((*MockstdDB)(nil).Begin))
}

// Close mocks base method
func (m *MockstdDB) Close() error {
	ret := m.ctrl.Call(m, "Close")
	ret0, _ := ret[0].(error)
	return ret0
}

// Close indicates an expected call of Close
func (mr *MockstdDBMockRecorder) Close() *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Close", reflect.TypeOf((*MockstdDB)(nil).Close))
}

// MockstdTx is a mock of stdTx interface
type MockstdTx struct {
	ctrl     *gomock.Controller
	recorder *MockstdTxMockRecorder
}

// MockstdTxMockRecorder is the mock recorder for MockstdTx
type MockstdTxMockRecorder struct {
	mock *MockstdTx
}

// NewMockstdTx creates a new mock instance
func NewMockstdTx(ctrl *gomock.Controller) *MockstdTx {
	mock := &MockstdTx{ctrl: ctrl}
	mock.recorder = &MockstdTxMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockstdTx) EXPECT() *MockstdTxMockRecorder {
	return m.recorder
}

// Query mocks base method
func (m *MockstdTx) Query(query string, args ...interface{}) (*sql.Rows, error) {
	varargs := []interface{}{query}
	for _, a := range args {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "Query", varargs...)
	ret0, _ := ret[0].(*sql.Rows)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Query indicates an expected call of Query
func (mr *MockstdTxMockRecorder) Query(query interface{}, args ...interface{}) *gomock.Call {
	varargs := append([]interface{}{query}, args...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Query", reflect.TypeOf((*MockstdTx)(nil).Query), varargs...)
}

// Exec mocks base method
func (m *MockstdTx) Exec(query string, args ...interface{}) (sql.Result, error) {
	varargs := []interface{}{query}
	for _, a := range args {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "Exec", varargs...)
	ret0, _ := ret[0].(sql.Result)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Exec indicates an expected call of Exec
func (mr *MockstdTxMockRecorder) Exec(query interface{}, args ...interface{}) *gomock.Call {
	varargs := append([]interface{}{query}, args...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Exec", reflect.TypeOf((*MockstdTx)(nil).Exec), varargs...)
}

// Commit mocks base method
func (m *MockstdTx) Commit() error {
	ret := m.ctrl.Call(m, "Commit")
	ret0, _ := ret[0].(error)
	return ret0
}

// Commit indicates an expected call of Commit
func (mr *MockstdTxMockRecorder) Commit() *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Commit", reflect.TypeOf((*MockstdTx)(nil).Commit))
}

// Rollback mocks base method
func (m *MockstdTx) Rollback() error {
	ret := m.ctrl.Call(m, "Rollback")
	ret0, _ := ret[0].(error)
	return ret0
}

// Rollback indicates an expected call of Rollback
func (mr *MockstdTxMockRecorder) Rollback() *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Rollback", reflect.TypeOf((*MockstdTx)(nil).Rollback))
}
