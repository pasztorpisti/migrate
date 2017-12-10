package template

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestExecute(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		tests := []*struct {
			name     string
			template string
			commands map[string]string
			vars     map[string]string
			result   string
		}{
			{
				name:     "raw string",
				template: `just ra\w string`,
				result:   `just raw string`,
			},
			{
				name:     "variable",
				template: `a{b}c{dd}e`,
				result:   `a<b_val>c<dd_val>e`,
				vars: map[string]string{
					"b":  "<b_val>",
					"dd": "<dd_val>",
				},
			},
			{
				name:     "cmd",
				template: `a{cmd:cmd1}b{cmd:cmd2}c`,
				result:   `a<cmd1_result>b<cmd2_result>c`,
				commands: map[string]string{
					"cmd1": "<cmd1_result>",
					"cmd2": "<cmd2_result>",
				},
			},
			{
				name:     "variable+cmd",
				template: `a{var1}b{cmd:cmd1}c`,
				result:   `a<var1_val>b<cmd1_result>c`,
				vars: map[string]string{
					"var1": "<var1_val>",
				},
				commands: map[string]string{
					"cmd1": "<cmd1_result>",
				},
			},
		}

		for _, test := range tests {
			t.Run(test.name, func(t *testing.T) {
				sections, err := Parse(test.template)
				require.NoError(t, err)

				result, err := Execute(sections, &ExecuteOptions{
					LookupVar: func(name string) (string, bool) {
						v, ok := test.vars[name]
						return v, ok
					},
					ExecCmd: func(command string, dir string, env []string) (string, error) {
						if v, ok := test.commands[command]; ok {
							return v, nil
						}
						return "", fmt.Errorf("test command not found: %q", command)
					},
				})
				require.NoError(t, err)
				assert.Equal(t, test.result, result)
			})
		}
	})

	t.Run("command dir", func(t *testing.T) {
		sections, err := Parse("{cmd:my_command:my_dir}")
		require.NoError(t, err)

		numExecCalls := 0
		result, err := Execute(sections, &ExecuteOptions{
			LookupVar: func(name string) (string, bool) {
				t.Fatal("unexpected LookupVar call")
				return "", false
			},
			ExecCmd: func(command string, dir string, env []string) (string, error) {
				numExecCalls += 1
				assert.Equal(t, command, "my_command")
				assert.Equal(t, dir, "my_dir")
				return "", nil
			},
		})
		require.NoError(t, err)
		assert.Equal(t, "", result)
		assert.Equal(t, 1, numExecCalls)
	})

	t.Run("command env", func(t *testing.T) {
		sections, err := Parse("{cmd:my_command:my_dir:env1=val1:env2=val2}")
		require.NoError(t, err)

		numExecCalls := 0
		result, err := Execute(sections, &ExecuteOptions{
			LookupVar: func(name string) (string, bool) {
				t.Fatal("unexpected LookupVar call")
				return "", false
			},
			ExecCmd: func(command string, dir string, env []string) (string, error) {
				numExecCalls += 1
				assert.Equal(t, command, "my_command")
				assert.Equal(t, dir, "my_dir")
				assert.Equal(t, env, []string{"env1=val1", "env2=val2"})
				return "", nil
			},
		})
		require.NoError(t, err)
		assert.Equal(t, "", result)
		assert.Equal(t, 1, numExecCalls)
	})

	t.Run("undefined variable", func(t *testing.T) {
		sections, err := Parse("{undefined}")
		require.NoError(t, err)

		_, err = Execute(sections, &ExecuteOptions{
			LookupVar: func(name string) (string, bool) {
				return "", false
			},
			ExecCmd: func(command string, dir string, env []string) (string, error) {
				t.Fatal("unexpected ExecCmd call")
				return "", nil
			},
		})
		assert.EqualError(t, err, `template variable "undefined" isn't defined`)
	})

	t.Run("command failure", func(t *testing.T) {
		sections, err := Parse("{cmd:my_command}")
		require.NoError(t, err)

		_, err = Execute(sections, &ExecuteOptions{
			LookupVar: func(name string) (string, bool) {
				t.Fatal("unexpected LookupVar call")
				return "", false
			},
			ExecCmd: func(command string, dir string, env []string) (string, error) {
				return "", assert.AnError
			},
		})
		assert.EqualError(t, err, `error executing command "my_command": `+assert.AnError.Error())
	})

	t.Run("invalid instruction", func(t *testing.T) {
		sections, err := Parse("{invalid_instruction:instruction_param1:instruction_param2}")
		require.NoError(t, err)

		_, err = Execute(sections, &ExecuteOptions{
			LookupVar: func(name string) (string, bool) {
				t.Fatal("unexpected LookupVar call")
				return "", false
			},
			ExecCmd: func(command string, dir string, env []string) (string, error) {
				t.Fatal("unexpected ExecCmd call")
				return "", nil
			},
		})
		assert.EqualError(t, err, `unknown template instruction: "invalid_instruction"`)
	})
}
