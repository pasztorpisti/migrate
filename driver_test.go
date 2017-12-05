package migrate

import (
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestRegisterDriver(t *testing.T) {
	t.Run("Register duplicate driver name", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		driver := NewMockDriver(ctrl)
		driver2 := NewMockDriver(ctrl)
		const name = "my_driver"

		registry := make(driverMap)
		registry.RegisterDriver(name, driver)

		panicked := false
		defer func() {
			r := recover()
			if r == nil {
				return
			}
			panicked = true
			assert.Equal(t, "duplicate database driver name: "+name, r)
		}()

		registry.RegisterDriver(name, driver2)
		assert.True(t, panicked)
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

		registry := make(driverMap)
		registry.RegisterDriver("nil_driver", nil)
		assert.True(t, panicked)
	})
}

func TestGetDriver(t *testing.T) {
	t.Run("Get existing driver", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		driver := NewMockDriver(ctrl)
		const name = "my_driver"

		registry := make(driverMap)
		registry.RegisterDriver(name, driver)

		res, err := registry.GetDriver(name)
		assert.NoError(t, err)
		assert.Equal(t, driver, res)
	})

	t.Run("Get missing driver", func(t *testing.T) {
		registry := make(driverMap)
		_, err := registry.GetDriver("missing_driver")
		assert.Equal(t, ErrDriverNotFound, err)
	})
}
