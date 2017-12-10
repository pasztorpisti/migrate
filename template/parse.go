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
	return ParseWithOptions(s, defaultOptions)
}

type Section struct {
	// RawString is always the raw string section from the original template
	// string in the original raw (escaped) format.
	RawString string

	// If IsParameter()==false then this is a non-empty unescaped string.
	String string

	// If IsParameter()==true then this is a non-empty list of unescaped
	// template params.
	Parameter []string
}

func (o *Section) IsParameter() bool {
	return len(o.Parameter) != 0
}

type Options struct {
	ParamOpen  byte
	ParamClose byte
	ParamSplit byte
	Escape     byte
}

var defaultOptions = &Options{
	ParamOpen:  '{',
	ParamClose: '}',
	ParamSplit: ':',
	Escape:     '\\',
}

var errLonelyTrailingEscape = errors.New("lonely trailing escape at the end of template string")
var errUnclosedTrailingTemplateParam = errors.New("reached the end of template string before closing a template parameter")

// ParseWithOptions is the same as Parse but you can use different special
// characters instead of \ { : }.
func ParseWithOptions(s string, opts *Options) ([]Section, error) {
	var res []Section

	var str bytes.Buffer
	length := len(s)
	pos := 0

	for {
		rawStrBegin := pos

	rawStrLoop:
		for pos < length {
			switch b := s[pos]; b {
			default:
				str.WriteByte(b)
				pos++
			case opts.ParamOpen:
				break rawStrLoop
			case opts.Escape:
				if pos+1 >= length {
					return nil, errLonelyTrailingEscape
				}
				pos++
				str.WriteByte(translateEscapedByte(s[pos]))
				pos++
			}
		}

		if str.Len() > 0 {
			res = append(res, Section{
				RawString: s[rawStrBegin:pos],
				String:    str.String(),
			})
			str.Reset()
		}

		if pos >= length {
			break
		}

		rawStrBegin = pos
		// skipping an opts.ParamOpen
		pos++

		var params []string

	paramsLoop:
		for pos < length {
			switch b := s[pos]; b {
			default:
				str.WriteByte(b)
				pos++
			case opts.ParamClose:
				break paramsLoop
			case opts.ParamSplit:
				params = append(params, str.String())
				str.Reset()
				pos++
			case opts.Escape:
				if pos+1 >= length {
					return nil, errLonelyTrailingEscape
				}
				pos++
				str.WriteByte(translateEscapedByte(s[pos]))
				pos++
			}
		}

		if pos >= length {
			return nil, errUnclosedTrailingTemplateParam
		}

		// skipping an opts.ParamClose
		pos++

		params = append(params, str.String())
		str.Reset()

		res = append(res, Section{
			RawString: s[rawStrBegin:pos],
			Parameter: params,
		})
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

// Escape returns a string in which all meta-characters ('{', '}', ':' and '\\')
// have been escaped with a backslash ('\\').
func Escape(s string) string {
	return EscapeWithOptions(s, defaultOptions)
}

// EscapeWithOptions is the same as Escape but you can customise the meta-characters.
func EscapeWithOptions(s string, opts *Options) string {
	isMeta := func(b byte) bool {
		return b == opts.ParamOpen || b == opts.ParamClose || b == opts.ParamSplit || b == opts.Escape
	}

	numMetas := 0
	for i := 0; i < len(s); i++ {
		if isMeta(s[i]) {
			numMetas++
		}
	}
	if numMetas == 0 {
		return s
	}

	var buf bytes.Buffer
	buf.Grow(len(s) + numMetas)

	for i := 0; i < len(s); i++ {
		if isMeta(s[i]) {
			buf.WriteByte(opts.Escape)
		}
		buf.WriteByte(s[i])
	}
	return buf.String()
}
