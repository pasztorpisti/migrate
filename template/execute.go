package template

import (
	"bytes"
	"fmt"
)

type ExecCmdFunc func(command, dir string, env []string) (string, error)

// ExecuteOptions provides parameters for Execute.
type ExecuteOptions struct {
	// LookupVar tells what to substitute in place of a "{name}" template parameter.
	// Implement your own or use os.LookupEnv.
	LookupVar func(name string) (string, bool)

	// ExecCmd tells what to substitute in place of a
	// "{cmd:<command>[:dir[:env1=value1[:env2=value2[...]]]]}" template parameter.
	//
	// You can provide your own implementation or use an existing one
	// which is the ExecCmd function or RemoveTrailingNewlines(ExecCmd).
	ExecCmd ExecCmdFunc
}

// Execute uses the output of Parse and replaces the template parameters with
// (environment) variables and the output of command executions.
func Execute(sections []Section, opts *ExecuteOptions) (string, error) {
	var buf bytes.Buffer
	for _, section := range sections {
		if section.IsRawString() {
			buf.WriteString(section.RawString)
			continue
		}

		// we know that len(section.Parameter) >= 1

		param := section.Parameter[0]
		if len(section.Parameter) == 1 {
			if v, ok := opts.LookupVar(param); ok {
				buf.WriteString(v)
				continue
			}
			return "", fmt.Errorf("template variable %q isn't defined", param)
		}

		if param != "cmd" {
			return "", fmt.Errorf("unknown template instruction: %q", param)
		}

		// param == "cmd" && len(section.Parameter) >= 2

		command := section.Parameter[1]
		dir := ""
		var env []string

		if len(section.Parameter) >= 3 {
			dir = section.Parameter[2]
			env = section.Parameter[3:]
		}

		s, err := opts.ExecCmd(command, dir, env)
		if err != nil {
			return "", fmt.Errorf("error executing command %q: %s", command, err)
		}
		buf.WriteString(s)
	}
	return buf.String(), nil
}
