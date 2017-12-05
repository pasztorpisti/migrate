package migrate

import (
	"errors"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSQLExecStep_AllowsTransaction(t *testing.T) {
	t.Run("NoTransaction=false", func(t *testing.T) {
		step := &SQLExecStep{
			NoTransaction: false,
		}
		assert.True(t, step.AllowsTransaction())
	})
	t.Run("NoTransaction=true", func(t *testing.T) {
		step := &SQLExecStep{
			NoTransaction: true,
		}
		assert.False(t, step.AllowsTransaction())
	})
}

func TestSQLExecStep_Execute(t *testing.T) {
	t.Run("DB Success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		db := NewMockDB(ctrl)
		printer := NewMockPrinter(ctrl)

		step := &SQLExecStep{
			Query:  "fake query",
			Args:   []interface{}{"str", 42},
			IsMeta: false,
		}
		ctx := ExecCtx{
			DB:     db,
			Output: printer,
		}

		db.EXPECT().Exec("fake query", []interface{}{"str", 42})

		err := step.Execute(ctx)
		assert.NoError(t, err)
	})
	t.Run("DB Error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		db := NewMockDB(ctrl)
		printer := NewMockPrinter(ctrl)

		step := &SQLExecStep{
			Query:  "fake query",
			Args:   []interface{}{"str", 42},
			IsMeta: false,
		}
		ctx := ExecCtx{
			DB:     db,
			Output: printer,
		}
		testErr := errors.New("test error")

		db.EXPECT().Exec("fake query", []interface{}{"str", 42}).Return(nil, testErr)

		err := step.Execute(ctx)
		assert.Equal(t, testErr, err)
	})
}

func TestSQLExecStep_Print(t *testing.T) {
	t.Run("IsMeta=false PrintSQL=false", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		printer := NewMockPrinter(ctrl)

		step := &SQLExecStep{
			Query:  "fake query",
			Args:   []interface{}{"str", 42},
			IsMeta: false,
		}
		ctx := PrintCtx{
			Output:   printer,
			PrintSQL: false,
		}
		step.Print(ctx)
	})
	t.Run("IsMeta=false PrintSQL=true", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		writer := NewMockWriter(ctrl)
		printer := NewPrinter(writer)

		step := &SQLExecStep{
			Query:  "fake query",
			Args:   []interface{}{"str", 42},
			IsMeta: false,
		}
		ctx := PrintCtx{
			Output:   printer,
			PrintSQL: true,
		}

		writer.EXPECT().Write(gomock.Any()).MinTimes(1)

		step.Print(ctx)
	})
	t.Run("IsMeta=true PrintMetaSQL=false", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		printer := NewMockPrinter(ctrl)

		step := &SQLExecStep{
			Query:  "fake query",
			Args:   []interface{}{"str", 42},
			IsMeta: true,
		}
		ctx := PrintCtx{
			Output:       printer,
			PrintMetaSQL: false,
		}
		step.Print(ctx)
	})
	t.Run("IsMeta=true PrintMetaSQL=true", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		writer := NewMockWriter(ctrl)
		printer := NewPrinter(writer)

		step := &SQLExecStep{
			Query:  "fake query",
			Args:   []interface{}{"str", 42},
			IsMeta: true,
		}
		ctx := PrintCtx{
			Output:       printer,
			PrintMetaSQL: true,
		}

		writer.EXPECT().Write(gomock.Any()).MinTimes(1)

		step.Print(ctx)
	})
}

func TestSteps_AllowsTransaction(t *testing.T) {
	t.Run("NumSteps=0", func(t *testing.T) {
		steps := Steps{}
		assert.True(t, steps.AllowsTransaction())
	})
	t.Run("NumSteps=1 AllowTransaction=false", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		step0 := NewMockStep(ctrl)
		steps := Steps{step0}

		step0.EXPECT().AllowsTransaction().Return(false)

		assert.False(t, steps.AllowsTransaction())
	})
	t.Run("NumSteps=1 AllowTransaction=true", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		step0 := NewMockStep(ctrl)
		steps := Steps{step0}

		step0.EXPECT().AllowsTransaction().Return(true)

		assert.True(t, steps.AllowsTransaction())
	})
	t.Run("NumSteps=2 AllowTransaction=[false,false]", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		step0 := NewMockStep(ctrl)
		step1 := NewMockStep(ctrl)
		steps := Steps{step0, step1}

		step0.EXPECT().AllowsTransaction().Return(false)
		step1.EXPECT().AllowsTransaction().Return(false)

		assert.False(t, steps.AllowsTransaction())
	})
	t.Run("NumSteps=2 AllowTransaction=[false,true]", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		step0 := NewMockStep(ctrl)
		step1 := NewMockStep(ctrl)
		steps := Steps{step0, step1}

		step0.EXPECT().AllowsTransaction().Return(false)
		step1.EXPECT().AllowsTransaction().Return(true)

		assert.False(t, steps.AllowsTransaction())
	})
	t.Run("NumSteps=2 AllowTransaction=[true,false]", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		step0 := NewMockStep(ctrl)
		step1 := NewMockStep(ctrl)
		steps := Steps{step0, step1}

		step0.EXPECT().AllowsTransaction().Return(true)
		step1.EXPECT().AllowsTransaction().Return(false)

		assert.False(t, steps.AllowsTransaction())
	})
	t.Run("NumSteps=2 AllowTransaction=[true,true]", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		step0 := NewMockStep(ctrl)
		step1 := NewMockStep(ctrl)
		steps := Steps{step0, step1}

		step0.EXPECT().AllowsTransaction().Return(true)
		step1.EXPECT().AllowsTransaction().Return(true)

		assert.True(t, steps.AllowsTransaction())
	})
}

func TestSteps_Execute(t *testing.T) {
	t.Run("DB Success", func(t *testing.T) {
		t.Run("NumSteps=0", func(t *testing.T) {
			ctrl := gomock.NewController(t)
			db := NewMockDB(ctrl)
			printer := NewMockPrinter(ctrl)

			steps := Steps{}
			ctx := ExecCtx{
				DB:     db,
				Output: printer,
			}
			err := steps.Execute(ctx)
			assert.NoError(t, err)
		})
		t.Run("NumSteps=1", func(t *testing.T) {
			ctrl := gomock.NewController(t)
			db := NewMockDB(ctrl)
			printer := NewMockPrinter(ctrl)
			step0 := NewMockStep(ctrl)

			steps := Steps{
				step0,
			}
			ctx := ExecCtx{
				DB:     db,
				Output: printer,
			}

			step0.EXPECT().Execute(ctx)

			err := steps.Execute(ctx)
			assert.NoError(t, err)
		})
		t.Run("NumSteps=2", func(t *testing.T) {
			ctrl := gomock.NewController(t)
			db := NewMockDB(ctrl)
			printer := NewMockPrinter(ctrl)
			step0 := NewMockStep(ctrl)
			step1 := NewMockStep(ctrl)

			steps := Steps{
				step0,
				step1,
			}
			ctx := ExecCtx{
				DB:     db,
				Output: printer,
			}

			gomock.InOrder(
				step0.EXPECT().Execute(ctx),
				step1.EXPECT().Execute(ctx),
			)

			err := steps.Execute(ctx)
			assert.NoError(t, err)
		})
		t.Run("NumSteps=3", func(t *testing.T) {
			ctrl := gomock.NewController(t)
			db := NewMockDB(ctrl)
			printer := NewMockPrinter(ctrl)
			step0 := NewMockStep(ctrl)
			step1 := NewMockStep(ctrl)
			step2 := NewMockStep(ctrl)

			steps := Steps{
				step0,
				step1,
				step2,
			}
			ctx := ExecCtx{
				DB:     db,
				Output: printer,
			}

			gomock.InOrder(
				step0.EXPECT().Execute(ctx),
				step1.EXPECT().Execute(ctx),
				step2.EXPECT().Execute(ctx),
			)

			err := steps.Execute(ctx)
			assert.NoError(t, err)
		})
	})
	t.Run("DB Error", func(t *testing.T) {
		t.Run("NumSteps=3 FailIndex=0", func(t *testing.T) {
			ctrl := gomock.NewController(t)
			db := NewMockDB(ctrl)
			printer := NewMockPrinter(ctrl)
			step0 := NewMockStep(ctrl)
			step1 := NewMockStep(ctrl)
			step2 := NewMockStep(ctrl)

			steps := Steps{
				step0,
				step1,
				step2,
			}
			ctx := ExecCtx{
				DB:     db,
				Output: printer,
			}

			testErr := errors.New("test error")
			step0.EXPECT().Execute(ctx).Return(testErr)

			err := steps.Execute(ctx)
			assert.Equal(t, testErr, err)
		})
		t.Run("NumSteps=3 FailIndex=1", func(t *testing.T) {
			ctrl := gomock.NewController(t)
			db := NewMockDB(ctrl)
			printer := NewMockPrinter(ctrl)
			step0 := NewMockStep(ctrl)
			step1 := NewMockStep(ctrl)
			step2 := NewMockStep(ctrl)

			steps := Steps{
				step0,
				step1,
				step2,
			}
			ctx := ExecCtx{
				DB:     db,
				Output: printer,
			}

			testErr := errors.New("test error")

			gomock.InOrder(
				step0.EXPECT().Execute(ctx),
				step1.EXPECT().Execute(ctx).Return(testErr),
			)

			err := steps.Execute(ctx)
			assert.Equal(t, testErr, err)
		})
		t.Run("NumSteps=3 FailIndex=2", func(t *testing.T) {
			ctrl := gomock.NewController(t)
			db := NewMockDB(ctrl)
			printer := NewMockPrinter(ctrl)
			step0 := NewMockStep(ctrl)
			step1 := NewMockStep(ctrl)
			step2 := NewMockStep(ctrl)

			steps := Steps{
				step0,
				step1,
				step2,
			}
			ctx := ExecCtx{
				DB:     db,
				Output: printer,
			}

			testErr := errors.New("test error")

			gomock.InOrder(
				step0.EXPECT().Execute(ctx),
				step1.EXPECT().Execute(ctx),
				step2.EXPECT().Execute(ctx).Return(testErr),
			)

			err := steps.Execute(ctx)
			assert.Equal(t, testErr, err)
		})
	})
}

func TestSteps_Print(t *testing.T) {
	t.Run("NumSteps=0", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		printer := NewMockPrinter(ctrl)
		steps := Steps{}
		ctx := PrintCtx{
			Output: printer,
		}
		steps.Print(ctx)
	})
	t.Run("NumSteps=1", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		printer := NewMockPrinter(ctrl)
		step0 := NewMockStep(ctrl)
		steps := Steps{step0}
		ctx := PrintCtx{
			Output: printer,
		}

		step0.EXPECT().Print(ctx)

		steps.Print(ctx)
	})
	t.Run("NumSteps=2", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		printer := NewMockPrinter(ctrl)
		step0 := NewMockStep(ctrl)
		step1 := NewMockStep(ctrl)
		steps := Steps{step0, step1}
		ctx := PrintCtx{
			Output: printer,
		}

		gomock.InOrder(
			step0.EXPECT().Print(ctx),
			step1.EXPECT().Print(ctx),
		)

		steps.Print(ctx)
	})
}

func TestTransactionIfAllowed_Execute(t *testing.T) {
	t.Run("DB Success", func(t *testing.T) {
		t.Run("NumSteps=0", func(t *testing.T) {
			ctrl := gomock.NewController(t)
			db := NewMockDB(ctrl)
			printer := NewMockPrinter(ctrl)

			steps := TransactionIfAllowed{Steps{}}
			ctx := ExecCtx{
				DB:     db,
				Output: printer,
			}
			err := steps.Execute(ctx)
			assert.NoError(t, err)
		})
		t.Run("NumSteps=1 AllowsTransaction=false", func(t *testing.T) {
			ctrl := gomock.NewController(t)
			db := NewMockDB(ctrl)
			printer := NewMockPrinter(ctrl)
			step0 := NewMockStep(ctrl)

			steps := TransactionIfAllowed{Steps{
				step0,
			}}
			ctx := ExecCtx{
				DB:     db,
				Output: printer,
			}

			gomock.InOrder(
				step0.EXPECT().AllowsTransaction().Return(false),
				step0.EXPECT().Execute(ctx),
			)

			err := steps.Execute(ctx)
			assert.NoError(t, err)
		})
		t.Run("NumSteps=1 AllowsTransaction=true", func(t *testing.T) {
			ctrl := gomock.NewController(t)
			db := NewMockDB(ctrl)
			tx := NewMockTX(ctrl)
			printer := NewMockPrinter(ctrl)
			step0 := NewMockStep(ctrl)

			steps := TransactionIfAllowed{Steps{
				step0,
			}}
			ctx := ExecCtx{
				DB:     db,
				Output: printer,
			}
			ctx2 := ExecCtx{
				DB:     tx,
				Output: printer,
			}
			gomock.InOrder(
				step0.EXPECT().AllowsTransaction().Return(true),
				db.EXPECT().BeginTX().Return(tx, nil),
				step0.EXPECT().Execute(ctx2),
				tx.EXPECT().Commit(),
			)

			err := steps.Execute(ctx)
			assert.NoError(t, err)
		})
		t.Run("NumSteps=2 AllowsTransaction=false", func(t *testing.T) {
			ctrl := gomock.NewController(t)
			db := NewMockDB(ctrl)
			printer := NewMockPrinter(ctrl)
			step0 := NewMockStep(ctrl)
			step1 := NewMockStep(ctrl)

			steps := TransactionIfAllowed{Steps{
				step0,
				step1,
			}}
			ctx := ExecCtx{
				DB:     db,
				Output: printer,
			}

			c0 := step0.EXPECT().AllowsTransaction().Return(true)
			c1 := step1.EXPECT().AllowsTransaction().Return(false)

			gomock.InOrder(
				step0.EXPECT().Execute(ctx).After(c0).After(c1),
				step1.EXPECT().Execute(ctx),
			)

			err := steps.Execute(ctx)
			assert.NoError(t, err)
		})
		t.Run("NumSteps=2 AllowsTransaction=true", func(t *testing.T) {
			ctrl := gomock.NewController(t)
			db := NewMockDB(ctrl)
			tx := NewMockTX(ctrl)
			printer := NewMockPrinter(ctrl)
			step0 := NewMockStep(ctrl)
			step1 := NewMockStep(ctrl)

			steps := TransactionIfAllowed{Steps{
				step0,
				step1,
			}}
			ctx := ExecCtx{
				DB:     db,
				Output: printer,
			}
			ctx2 := ExecCtx{
				DB:     tx,
				Output: printer,
			}

			c0 := step0.EXPECT().AllowsTransaction().Return(true)
			c1 := step1.EXPECT().AllowsTransaction().Return(true)
			gomock.InOrder(
				db.EXPECT().BeginTX().Return(tx, nil).After(c0).After(c1),
				step0.EXPECT().Execute(ctx2),
				step1.EXPECT().Execute(ctx2),
				tx.EXPECT().Commit(),
			)

			err := steps.Execute(ctx)
			assert.NoError(t, err)
		})
	})
	t.Run("DB Error", func(t *testing.T) {
		t.Run("NumSteps=2 AllowsTransaction=false FailIndex=0", func(t *testing.T) {
			ctrl := gomock.NewController(t)
			db := NewMockDB(ctrl)
			printer := NewMockPrinter(ctrl)
			step0 := NewMockStep(ctrl)
			step1 := NewMockStep(ctrl)

			steps := TransactionIfAllowed{Steps{
				step0,
				step1,
			}}
			ctx := ExecCtx{
				DB:     db,
				Output: printer,
			}

			testErr := errors.New("test error")

			c0 := step0.EXPECT().AllowsTransaction().Return(true)
			c1 := step1.EXPECT().AllowsTransaction().Return(false)
			step0.EXPECT().Execute(ctx).Return(testErr).After(c0).After(c1)

			err := steps.Execute(ctx)
			assert.Equal(t, testErr, err)
		})
		t.Run("NumSteps=2 AllowsTransaction=false FailIndex=1", func(t *testing.T) {
			ctrl := gomock.NewController(t)
			db := NewMockDB(ctrl)
			printer := NewMockPrinter(ctrl)
			step0 := NewMockStep(ctrl)
			step1 := NewMockStep(ctrl)

			steps := TransactionIfAllowed{Steps{
				step0,
				step1,
			}}
			ctx := ExecCtx{
				DB:     db,
				Output: printer,
			}

			testErr := errors.New("test error")
			c0 := step0.EXPECT().AllowsTransaction().Return(true)
			c1 := step1.EXPECT().AllowsTransaction().Return(false)
			gomock.InOrder(
				step0.EXPECT().Execute(ctx).After(c0).After(c1),
				step1.EXPECT().Execute(ctx).Return(testErr),
			)

			err := steps.Execute(ctx)
			assert.Equal(t, testErr, err)
		})
		t.Run("NumSteps=2 AllowsTransaction=true FailIndex=0", func(t *testing.T) {
			ctrl := gomock.NewController(t)
			db := NewMockDB(ctrl)
			tx := NewMockTX(ctrl)
			printer := NewMockPrinter(ctrl)
			step0 := NewMockStep(ctrl)
			step1 := NewMockStep(ctrl)

			steps := TransactionIfAllowed{Steps{
				step0,
				step1,
			}}
			ctx := ExecCtx{
				DB:     db,
				Output: printer,
			}
			ctx2 := ExecCtx{
				DB:     tx,
				Output: printer,
			}

			testErr := errors.New("test error")

			c0 := step0.EXPECT().AllowsTransaction().Return(true)
			c1 := step1.EXPECT().AllowsTransaction().Return(true)
			gomock.InOrder(
				db.EXPECT().BeginTX().Return(tx, nil).After(c0).After(c1),
				step0.EXPECT().Execute(ctx2).Return(testErr),
				tx.EXPECT().Rollback(),
			)

			err := steps.Execute(ctx)
			assert.Equal(t, testErr, err)
		})
		t.Run("NumSteps=2 AllowsTransaction=true FailIndex=1", func(t *testing.T) {
			ctrl := gomock.NewController(t)
			db := NewMockDB(ctrl)
			tx := NewMockTX(ctrl)
			printer := NewMockPrinter(ctrl)
			step0 := NewMockStep(ctrl)
			step1 := NewMockStep(ctrl)

			steps := TransactionIfAllowed{Steps{
				step0,
				step1,
			}}
			ctx := ExecCtx{
				DB:     db,
				Output: printer,
			}
			ctx2 := ExecCtx{
				DB:     tx,
				Output: printer,
			}

			testErr := errors.New("test error")

			c0 := step0.EXPECT().AllowsTransaction().Return(true)
			c1 := step1.EXPECT().AllowsTransaction().Return(true)
			gomock.InOrder(
				db.EXPECT().BeginTX().Return(tx, nil).After(c0).After(c1),
				step0.EXPECT().Execute(ctx2),
				step1.EXPECT().Execute(ctx2).Return(testErr),
				tx.EXPECT().Rollback(),
			)

			err := steps.Execute(ctx)
			assert.Equal(t, testErr, err)
		})
	})
}

func TestStepTitle(t *testing.T) {
	t.Run("Empty Title", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		writer := NewMockWriter(ctrl)
		printer := NewPrinter(writer)
		wrapped := NewMockStep(ctrl)
		step := StepTitle{
			Step: wrapped,
		}
		ctx := PrintCtx{
			Output: printer,
		}

		wrapped.EXPECT().Print(ctx)

		step.Print(ctx)
	})
	t.Run("NonEmpty Title", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		writer := NewMockWriter(ctrl)
		printer := NewPrinter(writer)
		wrapped := NewMockStep(ctrl)
		step := StepTitle{
			Step:  wrapped,
			Title: "test title",
		}
		ctx := PrintCtx{
			Output: printer,
		}

		gomock.InOrder(
			writer.EXPECT().Write(gomock.Any()).MinTimes(1),
			wrapped.EXPECT().Print(ctx),
		)

		step.Print(ctx)
	})
}
