package core

import (
	"database/sql"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestDBWrapper(t *testing.T) {
	t.Run("BeginTX", func(t *testing.T) {
		t.Run("Success", func(t *testing.T) {
			ctrl := gomock.NewController(t)
			db := NewMockstdDB(ctrl)
			w := WrapDB(db)

			fakeTx := &sql.Tx{}
			db.EXPECT().Begin().Return(fakeTx, nil)

			tx, err := w.BeginTX()

			assert.NoError(t, err)
			assert.NotNil(t, tx)
			assert.NotEqual(t, w, tx)
			assert.NotEqual(t, fakeTx, tx)
			ctrl.Finish()
		})
		t.Run("Error", func(t *testing.T) {
			ctrl := gomock.NewController(t)
			db := NewMockstdDB(ctrl)
			w := dbWrapper{db}

			db.EXPECT().Begin().Return(nil, assert.AnError)

			_, err := w.BeginTX()

			assert.Equal(t, assert.AnError, err)
			ctrl.Finish()
		})
	})

	t.Run("Close", func(t *testing.T) {
		t.Run("Success", func(t *testing.T) {
			ctrl := gomock.NewController(t)
			db := NewMockstdDB(ctrl)
			w := WrapDB(db)

			db.EXPECT().Close()

			err := w.Close()

			assert.NoError(t, err)
			ctrl.Finish()
		})
		t.Run("Error", func(t *testing.T) {
			ctrl := gomock.NewController(t)
			db := NewMockstdDB(ctrl)
			w := WrapDB(db)

			db.EXPECT().Close().Return(assert.AnError)

			err := w.Close()

			assert.Equal(t, assert.AnError, err)
			ctrl.Finish()
		})
	})

	t.Run("Query", func(t *testing.T) {
		t.Run("Success", func(t *testing.T) {
			ctrl := gomock.NewController(t)
			db := NewMockstdDB(ctrl)
			w := WrapDB(db)

			rows := &sql.Rows{}
			db.EXPECT().Query("test query", "arg1", 2).Return(rows, nil)

			res, err := w.Query("test query", "arg1", 2)

			assert.NoError(t, err)
			assert.Equal(t, rows, res)
			ctrl.Finish()
		})
		t.Run("Error", func(t *testing.T) {
			ctrl := gomock.NewController(t)
			db := NewMockstdDB(ctrl)
			w := WrapDB(db)

			db.EXPECT().Query("test query", "arg1", 2).Return(nil, assert.AnError)

			_, err := w.Query("test query", "arg1", 2)

			assert.Equal(t, assert.AnError, err)
			ctrl.Finish()
		})
	})

	t.Run("Exec", func(t *testing.T) {
		t.Run("Success", func(t *testing.T) {
			ctrl := gomock.NewController(t)
			db := NewMockstdDB(ctrl)
			w := WrapDB(db)
			sqlRes := NewMockResult(ctrl)

			db.EXPECT().Exec("test query", "arg1", 2).Return(sqlRes, nil)

			res, err := w.Exec("test query", "arg1", 2)

			assert.NoError(t, err)
			assert.Equal(t, sqlRes, res)
			ctrl.Finish()
		})
		t.Run("Error", func(t *testing.T) {
			ctrl := gomock.NewController(t)
			db := NewMockstdDB(ctrl)
			w := WrapDB(db)

			db.EXPECT().Exec("test query", "arg1", 2).Return(nil, assert.AnError)

			_, err := w.Exec("test query", "arg1", 2)

			assert.Equal(t, assert.AnError, err)
			ctrl.Finish()
		})
	})
}

func TestTxWrapper(t *testing.T) {
	t.Run("BeginTX", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		tx := NewMockstdTx(ctrl)
		w := wrapTx(tx)

		rtx, err := w.BeginTX()

		assert.NoError(t, err)
		assert.NotNil(t, rtx)
		assert.NotEqual(t, tx, rtx)
		assert.NotEqual(t, w, rtx)
		ctrl.Finish()
	})

	t.Run("Query", func(t *testing.T) {
		t.Run("Success", func(t *testing.T) {
			ctrl := gomock.NewController(t)
			tx := NewMockstdTx(ctrl)
			w := wrapTx(tx)

			rows := &sql.Rows{}
			tx.EXPECT().Query("test query", "arg1", 2).Return(rows, nil)

			res, err := w.Query("test query", "arg1", 2)

			assert.NoError(t, err)
			assert.Equal(t, rows, res)
			ctrl.Finish()
		})
		t.Run("Error", func(t *testing.T) {
			ctrl := gomock.NewController(t)
			tx := NewMockstdTx(ctrl)
			w := wrapTx(tx)

			tx.EXPECT().Query("test query", "arg1", 2).Return(nil, assert.AnError)

			_, err := w.Query("test query", "arg1", 2)

			assert.Equal(t, assert.AnError, err)
			ctrl.Finish()
		})
	})

	t.Run("Exec", func(t *testing.T) {
		t.Run("Success", func(t *testing.T) {
			ctrl := gomock.NewController(t)
			tx := NewMockstdTx(ctrl)
			w := wrapTx(tx)
			sqlRes := NewMockResult(ctrl)

			tx.EXPECT().Exec("test query", "arg1", 2).Return(sqlRes, nil)

			res, err := w.Exec("test query", "arg1", 2)

			assert.NoError(t, err)
			assert.Equal(t, sqlRes, res)
			ctrl.Finish()
		})
		t.Run("Error", func(t *testing.T) {
			ctrl := gomock.NewController(t)
			tx := NewMockstdTx(ctrl)
			w := wrapTx(tx)

			tx.EXPECT().Exec("test query", "arg1", 2).Return(nil, assert.AnError)

			_, err := w.Exec("test query", "arg1", 2)

			assert.Equal(t, assert.AnError, err)
			ctrl.Finish()
		})
	})
}

func TestRecursiveTXWrapper(t *testing.T) {
	t.Run("BeginTX", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		tx := NewMockTX(ctrl)
		w := wrapTxWrapper(tx)

		rtx, err := w.BeginTX()

		assert.NoError(t, err)
		assert.NotNil(t, rtx)
		assert.NotEqual(t, w, rtx)
		assert.NotEqual(t, tx, rtx)
		ctrl.Finish()
	})

	t.Run("Query", func(t *testing.T) {
		t.Run("Success", func(t *testing.T) {
			ctrl := gomock.NewController(t)
			tx := NewMockTX(ctrl)
			w := wrapTxWrapper(tx)

			rows := &sql.Rows{}
			tx.EXPECT().Query("test query", "arg1", 2).Return(rows, nil)

			res, err := w.Query("test query", "arg1", 2)

			assert.NoError(t, err)
			assert.Equal(t, rows, res)
			ctrl.Finish()
		})
		t.Run("Error", func(t *testing.T) {
			ctrl := gomock.NewController(t)
			tx := NewMockTX(ctrl)
			w := wrapTxWrapper(tx)

			tx.EXPECT().Query("test query", "arg1", 2).Return(nil, assert.AnError)

			_, err := w.Query("test query", "arg1", 2)

			assert.Equal(t, assert.AnError, err)
			ctrl.Finish()
		})
	})

	t.Run("Exec", func(t *testing.T) {
		t.Run("Success", func(t *testing.T) {
			ctrl := gomock.NewController(t)
			tx := NewMockTX(ctrl)
			w := wrapTxWrapper(tx)
			sqlRes := NewMockResult(ctrl)

			tx.EXPECT().Exec("test query", "arg1", 2).Return(sqlRes, nil)

			res, err := w.Exec("test query", "arg1", 2)

			assert.NoError(t, err)
			assert.Equal(t, sqlRes, res)
			ctrl.Finish()
		})
		t.Run("Error", func(t *testing.T) {
			ctrl := gomock.NewController(t)
			tx := NewMockTX(ctrl)
			w := wrapTxWrapper(tx)

			tx.EXPECT().Exec("test query", "arg1", 2).Return(nil, assert.AnError)

			_, err := w.Exec("test query", "arg1", 2)

			assert.Equal(t, assert.AnError, err)
			ctrl.Finish()
		})
	})
}

func TestNestedTransactions(t *testing.T) {
	t.Run("Commit all", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		stdDB := NewMockstdDB(ctrl)

		fakeTx := &sql.Tx{}
		stdDB.EXPECT().Begin().Return(fakeTx, nil)

		db := WrapDB(stdDB)
		tx, err := db.BeginTX()
		assert.NoError(t, err)
		rtx, err := tx.BeginTX()
		assert.NoError(t, err)
		rtx2, err := rtx.BeginTX()
		assert.NoError(t, err)

		err = rtx2.Commit()
		assert.NoError(t, err)
		err = rtx.Commit()
		assert.NoError(t, err)

		ctrl.Finish()
	})

	t.Run("Rollback all", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		stdDB := NewMockstdDB(ctrl)

		fakeTx := &sql.Tx{}
		stdDB.EXPECT().Begin().Return(fakeTx, nil)

		db := WrapDB(stdDB)
		tx, err := db.BeginTX()
		assert.NoError(t, err)
		rtx, err := tx.BeginTX()
		assert.NoError(t, err)
		rtx2, err := rtx.BeginTX()
		assert.NoError(t, err)

		err = rtx2.Rollback()
		assert.NoError(t, err)
		err = rtx.Rollback()
		assert.NoError(t, err)

		ctrl.Finish()
	})

	t.Run("Commit fails after Rollback", func(t *testing.T) {
		t.Run("Rollback-Commit", func(t *testing.T) {
			ctrl := gomock.NewController(t)
			stdDB := NewMockstdDB(ctrl)

			fakeTx := &sql.Tx{}
			stdDB.EXPECT().Begin().Return(fakeTx, nil)

			db := WrapDB(stdDB)
			tx, err := db.BeginTX()
			assert.NoError(t, err)
			rtx, err := tx.BeginTX()
			assert.NoError(t, err)
			rtx2, err := rtx.BeginTX()
			assert.NoError(t, err)

			err = rtx2.Rollback()
			assert.NoError(t, err)
			err = rtx.Commit()
			assert.Equal(t, errCommitAfterChildRollback, err)

			ctrl.Finish()
		})
		t.Run("Rollback-Rollback-Commit", func(t *testing.T) {
			ctrl := gomock.NewController(t)
			stdDB := NewMockstdDB(ctrl)

			fakeTx := &sql.Tx{}
			stdDB.EXPECT().Begin().Return(fakeTx, nil)

			db := WrapDB(stdDB)
			tx, err := db.BeginTX()
			assert.NoError(t, err)
			rtx, err := tx.BeginTX()
			assert.NoError(t, err)
			rtx2, err := rtx.BeginTX()
			assert.NoError(t, err)

			err = rtx2.Rollback()
			assert.NoError(t, err)
			err = rtx.Rollback()
			assert.NoError(t, err)
			err = tx.Commit()
			assert.Equal(t, errCommitAfterChildRollback, err)

			ctrl.Finish()
		})
		t.Run("Commit-Rollback-Commit", func(t *testing.T) {
			ctrl := gomock.NewController(t)
			stdDB := NewMockstdDB(ctrl)

			fakeTx := &sql.Tx{}
			stdDB.EXPECT().Begin().Return(fakeTx, nil)

			db := WrapDB(stdDB)
			tx, err := db.BeginTX()
			assert.NoError(t, err)
			rtx, err := tx.BeginTX()
			assert.NoError(t, err)
			rtx2, err := rtx.BeginTX()
			assert.NoError(t, err)

			err = rtx2.Commit()
			assert.NoError(t, err)
			err = rtx.Rollback()
			assert.NoError(t, err)
			err = tx.Commit()
			assert.Equal(t, errCommitAfterChildRollback, err)

			ctrl.Finish()
		})
	})

	t.Run("Commit or Rollback fails on finished TX", func(t *testing.T) {
		t.Run("Commit first", func(t *testing.T) {
			t.Run("Level3", func(t *testing.T) {
				ctrl := gomock.NewController(t)
				stdDB := NewMockstdDB(ctrl)

				fakeTx := &sql.Tx{}
				stdDB.EXPECT().Begin().Return(fakeTx, nil)

				db := WrapDB(stdDB)
				tx, err := db.BeginTX()
				assert.NoError(t, err)
				rtx, err := tx.BeginTX()
				assert.NoError(t, err)
				rtx2, err := rtx.BeginTX()
				assert.NoError(t, err)

				err = rtx2.Commit()
				assert.NoError(t, err)
				err = rtx2.Commit()
				assert.Equal(t, errCommitFinishedTX, err)
				err = rtx2.Rollback()
				assert.Equal(t, errRollbackFinishedTX, err)

				ctrl.Finish()
			})
			t.Run("Level2", func(t *testing.T) {
				ctrl := gomock.NewController(t)
				stdDB := NewMockstdDB(ctrl)

				fakeTx := &sql.Tx{}
				stdDB.EXPECT().Begin().Return(fakeTx, nil)

				db := WrapDB(stdDB)
				tx, err := db.BeginTX()
				assert.NoError(t, err)
				rtx, err := tx.BeginTX()
				assert.NoError(t, err)
				rtx2, err := rtx.BeginTX()
				assert.NoError(t, err)

				err = rtx2.Commit()
				assert.NoError(t, err)
				err = rtx.Commit()
				assert.NoError(t, err)
				err = rtx.Commit()
				assert.Equal(t, errCommitFinishedTX, err)
				err = rtx.Rollback()
				assert.Equal(t, errRollbackFinishedTX, err)

				ctrl.Finish()
			})
		})

		t.Run("Rollback first", func(t *testing.T) {
			t.Run("Level3", func(t *testing.T) {
				ctrl := gomock.NewController(t)
				stdDB := NewMockstdDB(ctrl)

				fakeTx := &sql.Tx{}
				stdDB.EXPECT().Begin().Return(fakeTx, nil)

				db := WrapDB(stdDB)
				tx, err := db.BeginTX()
				assert.NoError(t, err)
				rtx, err := tx.BeginTX()
				assert.NoError(t, err)
				rtx2, err := rtx.BeginTX()
				assert.NoError(t, err)

				err = rtx2.Rollback()
				assert.NoError(t, err)
				err = rtx2.Commit()
				assert.Equal(t, errCommitFinishedTX, err)
				err = rtx2.Rollback()
				assert.Equal(t, errRollbackFinishedTX, err)

				ctrl.Finish()
			})
			t.Run("Level2", func(t *testing.T) {
				ctrl := gomock.NewController(t)
				stdDB := NewMockstdDB(ctrl)

				fakeTx := &sql.Tx{}
				stdDB.EXPECT().Begin().Return(fakeTx, nil)

				db := WrapDB(stdDB)
				tx, err := db.BeginTX()
				assert.NoError(t, err)
				rtx, err := tx.BeginTX()
				assert.NoError(t, err)
				rtx2, err := rtx.BeginTX()
				assert.NoError(t, err)

				err = rtx2.Rollback()
				assert.NoError(t, err)
				err = rtx.Rollback()
				assert.NoError(t, err)
				err = rtx.Commit()
				assert.Equal(t, errCommitAfterChildRollback, err)
				err = rtx.Rollback()
				assert.Equal(t, errRollbackFinishedTX, err)

				ctrl.Finish()
			})
		})
	})

	t.Run("Unfinished child tx", func(t *testing.T) {
		t.Run("Commit with unfinished child tx", func(t *testing.T) {
			t.Run("Level2", func(t *testing.T) {
				ctrl := gomock.NewController(t)
				stdDB := NewMockstdDB(ctrl)

				fakeTx := &sql.Tx{}
				stdDB.EXPECT().Begin().Return(fakeTx, nil)

				db := WrapDB(stdDB)
				tx, err := db.BeginTX()
				assert.NoError(t, err)
				rtx, err := tx.BeginTX()
				assert.NoError(t, err)
				_, err = rtx.BeginTX()
				assert.NoError(t, err)

				err = rtx.Commit()
				assert.Equal(t, errCommitWithUnfinishedChildren, err)

				ctrl.Finish()
			})
			t.Run("Level1", func(t *testing.T) {
				ctrl := gomock.NewController(t)
				stdDB := NewMockstdDB(ctrl)

				fakeTx := &sql.Tx{}
				stdDB.EXPECT().Begin().Return(fakeTx, nil)

				db := WrapDB(stdDB)
				tx, err := db.BeginTX()
				assert.NoError(t, err)
				_, err = tx.BeginTX()
				assert.NoError(t, err)

				err = tx.Commit()
				assert.Equal(t, errCommitWithUnfinishedChildren, err)

				ctrl.Finish()
			})
		})
		t.Run("Rollback with unfinished child tx", func(t *testing.T) {
			t.Run("Level2", func(t *testing.T) {
				ctrl := gomock.NewController(t)
				stdDB := NewMockstdDB(ctrl)

				fakeTx := &sql.Tx{}
				stdDB.EXPECT().Begin().Return(fakeTx, nil)

				db := WrapDB(stdDB)
				tx, err := db.BeginTX()
				assert.NoError(t, err)
				rtx, err := tx.BeginTX()
				assert.NoError(t, err)
				_, err = rtx.BeginTX()
				assert.NoError(t, err)

				err = rtx.Rollback()
				assert.Equal(t, errRollbackWithUnfinishedChildren, err)

				ctrl.Finish()
			})
			t.Run("Level1", func(t *testing.T) {
				ctrl := gomock.NewController(t)
				stdDB := NewMockstdDB(ctrl)

				fakeTx := &sql.Tx{}
				stdDB.EXPECT().Begin().Return(fakeTx, nil)

				db := WrapDB(stdDB)
				tx, err := db.BeginTX()
				assert.NoError(t, err)
				_, err = tx.BeginTX()
				assert.NoError(t, err)

				err = tx.Rollback()
				assert.Equal(t, errRollbackWithUnfinishedChildren, err)

				ctrl.Finish()
			})
		})
	})
}
