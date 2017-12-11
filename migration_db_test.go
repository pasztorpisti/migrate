package migrate

import (
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestRegisterDriverFactory(t *testing.T) {
	t.Run("Register duplicate driver name", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		df := NewMockDriverFactory(ctrl)
		df2 := NewMockDriverFactory(ctrl)
		const name = "my_driver"

		registry := make(driverFactoryMap)
		registry.RegisterDriverFactory(name, df)

		panicked := false
		defer func() {
			r := recover()
			if r == nil {
				return
			}
			panicked = true
			assert.Equal(t, "duplicate database driver name: "+name, r)
		}()

		registry.RegisterDriverFactory(name, df2)
		assert.True(t, panicked)
		ctrl.Finish()
	})

	t.Run("Register nil driver", func(t *testing.T) {
		panicked := false
		defer func() {
			r := recover()
			if r == nil {
				return
			}
			panicked = true
			assert.Equal(t, "driver is nil", r)
		}()

		registry := make(driverFactoryMap)
		registry.RegisterDriverFactory("nil_driver", nil)
		assert.True(t, panicked)
	})
}

func TestGetDriverFactory(t *testing.T) {
	t.Run("Get existing driver", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		df := NewMockDriverFactory(ctrl)
		const name = "my_driver"

		registry := make(driverFactoryMap)
		registry.RegisterDriverFactory(name, df)

		res, ok := registry.GetDriverFactory(name)
		assert.True(t, ok)
		assert.Equal(t, df, res)
		ctrl.Finish()
	})

	t.Run("Get missing driver", func(t *testing.T) {
		registry := make(driverFactoryMap)
		_, ok := registry.GetDriverFactory("missing_driver")
		assert.False(t, ok)
	})
}
