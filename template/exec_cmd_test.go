package template

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestRemoveTrailingNewlines(t *testing.T) {
	tests := []*struct {
		input, output string
	}{
		{"output", "output"},
		{"output\n", "output"},
		{"output\n\n", "output"},
		{"\noutput", "\noutput"},
		{"\noutput\n", "\noutput"},
		{"\noutput\n\n", "\noutput"},
		{"output\noutput", "output\noutput"},
		{"output\noutput\n", "output\noutput"},
		{"output\noutput\n\n", "output\noutput"},
	}

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			execCmd := RemoveTrailingNewlines(func(command, dir string, env []string) (string, error) {
				return test.input, nil
			})
			result, err := execCmd("", "", nil)
			require.NoError(t, err)
			assert.Equal(t, test.output, result)
		})
	}
}

func TestExecCmd(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		s, err := ExecCmd(">&2 echo \"err output\"; echo \"${MY_ENV} woof woof\"; >&2 echo \"err output 2\";", "", []string{"MY_ENV=MY_VAL"})
		require.NoError(t, err)
		assert.Equal(t, "MY_VAL woof woof\n", s)
	})

	t.Run("error", func(t *testing.T) {
		_, err := ExecCmd("exit 1;", "", nil)
		assert.Error(t, err)
	})
}
