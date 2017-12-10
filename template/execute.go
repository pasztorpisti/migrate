package template

import (
	"bytes"
	"fmt"
)

type ExecCmdFunc func(command, dir string, env []string) (string, error)

// ExecuteInput provides parameters for Execute.
type ExecuteInput struct {
	// Sections is the output of Parse.
	Sections []Section

	// Options should be the options used to parse the specified sections.
	// Use nil to use the default options.
	Options *Options

	// LookupVar tells what to substitute in place of a "{var:<name>}" template parameter.
	// Implement your own or use os.LookupEnv.
	LookupVar func(name string) (string, bool)

	// ExecCmd tells what to substitute in place of a
	// "{cmd:<command>[:dir[:env1=value1[:env2=value2[...]]]]}" template parameter.
	//
	// You can provide your own implementation or use an existing one
	// which is the ExecCmd function or RemoveTrailingNewlines(ExecCmd).
	ExecCmd ExecCmdFunc

	// VarParamName is the <template_param_name> of the
	// {<template_param_name>:<name>} template parameters that are replaced
	// with variable values looked up by the ExecuteInput.LookupVar function.
	VarParamName string

	// CmdParamName is the <template_param_name> of the
	// {<template_param_name>:<command>[:workdir[:env1=val1[:env2=val2[...]]]]}
	// template parameters that are replaced with the output of command
	// executions performed by the ExecuteInput.ExecCmd function.
	CmdParamName string

	// IgnoreUnknownTemplateParams instructs Execute to leave unknown template
	// params untouched in the result string without failing.
	// The only known template parameter names are
	// ExecuteInput.VarParamName and ExecuteInput.CmdParamName.
	IgnoreUnknownTemplateParams bool

	// EscapedResult==true instructs Execute to substitute variables and
	// command outputs into the result string in escaped format.
	// Aside from this string sections residing outside template parameters
	// are also left escaped (those parts of the template string would
	// be unescaped with EscapedResult==false).
	//
	// This parameter is useful when you want to perform multiple Parse-Execute
	// call pairs on a template string. The first Parse-Execute call pair
	// has to use EscapedResult==true. The second Parse-Execute pair can use
	// the output of the first Parse-Execute with EscapedResult==false.
	// If there are multiple Parse-Execute pairs in the chain then only the
	// last one should use EscapedResult==false.
	EscapedResult bool
}

// Execute uses the output of Parse and replaces the template parameters with
// (environment) variables and the output of command executions.
func Execute(input *ExecuteInput) (string, error) {
	varParamName := input.VarParamName
	if varParamName == "" {
		varParamName = "var"
	}

	cmdParamName := input.CmdParamName
	if cmdParamName == "" {
		cmdParamName = "cmd"
	}

	opts := input.Options
	if opts == nil {
		opts = defaultOptions
	}

	var buf bytes.Buffer
	for _, section := range input.Sections {
		if !section.IsParameter() {
			if input.EscapedResult {
				buf.WriteString(section.RawString)
			} else {
				buf.WriteString(section.String)
			}
			continue
		}

		// we know that len(section.Parameter) >= 1

		param := section.Parameter[0]
		if param == varParamName {
			if len(section.Parameter) != 2 {
				return "", fmt.Errorf("invalid {%s:<name>} template parameter: %q", varParamName, section.RawString)
			}
			varName := section.Parameter[1]
			if v, ok := input.LookupVar(varName); ok {
				if input.EscapedResult {
					buf.WriteString(EscapeWithOptions(v, opts))
				} else {
					buf.WriteString(v)
				}
				continue
			}
			return "", fmt.Errorf("template variable %q isn't defined", section.RawString)
		}

		if param != cmdParamName {
			if !input.IgnoreUnknownTemplateParams {
				return "", fmt.Errorf("unknown template instruction: %q", param)
			}
			buf.WriteString(section.RawString)
			continue
		}

		// param == "cmd" && len(section.Parameter) >= 2

		command := section.Parameter[1]
		dir := ""
		var env []string

		if len(section.Parameter) >= 3 {
			dir = section.Parameter[2]
			env = section.Parameter[3:]
		}

		s, err := input.ExecCmd(command, dir, env)
		if err != nil {
			return "", fmt.Errorf("error executing command %q: %s", command, err)
		}
		if input.EscapedResult {
			buf.WriteString(EscapeWithOptions(s, opts))
		} else {
			buf.WriteString(s)
		}
	}
	return buf.String(), nil
}
