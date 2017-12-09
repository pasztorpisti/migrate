package template

import (
	"bytes"
	"errors"
)

// Parse parses a template string that contains template parameters in the
// form {param[:param2[:param3[...]]]}. The interpretation of the parameters
// is up to the caller. To treat a special character (e.g.: '{') as a normal
// one escape it with a backslash.
//
// There are some "printf" style escape sequences available: \\, \t, \r, \n.
//
// For this parser a template string is simply a sequence of raw strings and
// parameters (where a parameter is a slice of strings).
//
// Example template strings: "my name is {name}" or "my name is {name:3:3}"
func Parse(s string) ([]Section, error) {
	return ParseWithOptions(s, defaultParseOptions)
}

type Section struct {
	RawString string
	Parameter []string
}

func (o *Section) IsRawString() bool {
	return len(o.Parameter) == 0
}

type ParseOptions struct {
	ParamOpen  byte
	ParamClose byte
	ParamSplit byte
	Escape     byte
}

var defaultParseOptions = &ParseOptions{
	ParamOpen:  '{',
	ParamClose: '}',
	ParamSplit: ':',
	Escape:     '\\',
}

var errLonelyTrailingEscape = errors.New("lonely trailing escape at the end of template string")
var errUnclosedTrailingTemplateParam = errors.New("reached the end of template string before closing a template parameter")

// ParseWithOptions is the same as Parse but you can use different special
// characters instead of \ { : }.
func ParseWithOptions(s string, opts *ParseOptions) ([]Section, error) {
	var res []Section

	var rawStr bytes.Buffer
	length := len(s)
	pos := 0

	for {
	rawStrLoop:
		for pos < length {
			switch b := s[pos]; b {
			default:
				rawStr.WriteByte(b)
				pos++
			case opts.ParamOpen:
				break rawStrLoop
			case opts.Escape:
				if pos+1 >= length {
					return nil, errLonelyTrailingEscape
				}
				pos++
				rawStr.WriteByte(translateEscapedByte(s[pos]))
				pos++
			}
		}

		if rawStr.Len() > 0 {
			res = append(res, Section{RawString: rawStr.String()})
			rawStr.Reset()
		}

		if pos >= length {
			break
		}

		// skipping an opts.ParamOpen
		pos++

		var params []string

	paramsLoop:
		for pos < length {
			switch b := s[pos]; b {
			default:
				rawStr.WriteByte(b)
				pos++
			case opts.ParamClose:
				break paramsLoop
			case opts.ParamSplit:
				params = append(params, rawStr.String())
				rawStr.Reset()
				pos++
			case opts.Escape:
				if pos+1 >= length {
					return nil, errLonelyTrailingEscape
				}
				pos++
				rawStr.WriteByte(translateEscapedByte(s[pos]))
				pos++
			}
		}

		if pos >= length {
			return nil, errUnclosedTrailingTemplateParam
		}

		// skipping an opts.ParamClose
		pos++

		params = append(params, rawStr.String())
		rawStr.Reset()

		res = append(res, Section{Parameter: params})
	}

	return res, nil
}

func translateEscapedByte(e byte) byte {
	switch e {
	case 'r':
		return '\r'
	case 'n':
		return '\n'
	case 't':
		return '\t'
	default:
		return e
	}
}
