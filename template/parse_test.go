package template

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestParse(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		tests := []*struct {
			template string
			sections []Section
		}{
			{
				template: "",
				sections: []Section(nil),
			},
			{
				template: "{}",
				sections: []Section{
					{RawString: "{}", Parameter: []string{""}},
				},
			},
			{
				template: "{}{}",
				sections: []Section{
					{RawString: "{}", Parameter: []string{""}},
					{RawString: "{}", Parameter: []string{""}},
				},
			},
			{
				template: "a",
				sections: []Section{
					{RawString: "a", String: "a"},
				},
			},
			{
				template: "a{}{}",
				sections: []Section{
					{RawString: "a", String: "a"},
					{RawString: "{}", Parameter: []string{""}},
					{RawString: "{}", Parameter: []string{""}},
				},
			},
			{
				template: "{}a{}",
				sections: []Section{
					{RawString: "{}", Parameter: []string{""}},
					{RawString: "a", String: "a"},
					{RawString: "{}", Parameter: []string{""}},
				},
			},
			{
				template: "{}{}a",
				sections: []Section{
					{RawString: "{}", Parameter: []string{""}},
					{RawString: "{}", Parameter: []string{""}},
					{RawString: "a", String: "a"},
				},
			},
			{
				template: "{}a1{}b2",
				sections: []Section{
					{RawString: "{}", Parameter: []string{""}},
					{RawString: "a1", String: "a1"},
					{RawString: "{}", Parameter: []string{""}},
					{RawString: "b2", String: "b2"},
				},
			},
			{
				template: "a1{}b2{}c3",
				sections: []Section{
					{RawString: "a1", String: "a1"},
					{RawString: "{}", Parameter: []string{""}},
					{RawString: "b2", String: "b2"},
					{RawString: "{}", Parameter: []string{""}},
					{RawString: "c3", String: "c3"},
				},
			},
			{
				template: "\\{",
				sections: []Section{
					{RawString: "\\{", String: "{"},
				},
			},
			{
				template: "\\\\",
				sections: []Section{
					{RawString: "\\\\", String: "\\"},
				},
			},
			{
				template: "\\r\\n\\t",
				sections: []Section{
					{RawString: "\\r\\n\\t", String: "\r\n\t"},
				},
			},
			{
				template: "\\\\\\{",
				sections: []Section{
					{RawString: "\\\\\\{", String: "\\{"},
				},
			},
			{
				template: "{\\:}",
				sections: []Section{
					{RawString: "{\\:}", Parameter: []string{":"}},
				},
			},
			{
				template: "{a\\:b}",
				sections: []Section{
					{RawString: "{a\\:b}", Parameter: []string{"a:b"}},
				},
			},
			{
				template: "{\\}}",
				sections: []Section{
					{RawString: "{\\}}", Parameter: []string{"}"}},
				},
			},
			{
				template: "{\\}\\:\\}{}",
				sections: []Section{
					{RawString: "{\\}\\:\\}{}", Parameter: []string{"}:}{"}},
				},
			},
			{
				template: "}",
				sections: []Section{
					{RawString: "}", String: "}"},
				},
			},
			{
				template: ":",
				sections: []Section{
					{RawString: ":", String: ":"},
				},
			},
			{
				template: "a1:b2}c3",
				sections: []Section{
					{RawString: "a1:b2}c3", String: "a1:b2}c3"},
				},
			},
			{
				template: "{:}",
				sections: []Section{
					{RawString: "{:}", Parameter: []string{"", ""}},
				},
			},
			{
				template: "{::}",
				sections: []Section{
					{RawString: "{::}", Parameter: []string{"", "", ""}},
				},
			},
			{
				template: "{a::}",
				sections: []Section{
					{RawString: "{a::}", Parameter: []string{"a", "", ""}},
				},
			},
			{
				template: "{:a:}",
				sections: []Section{
					{RawString: "{:a:}", Parameter: []string{"", "a", ""}},
				},
			},
			{
				template: "{::a}",
				sections: []Section{
					{RawString: "{::a}", Parameter: []string{"", "", "a"}},
				},
			},
			{
				template: "{a1:b2:c3}",
				sections: []Section{
					{RawString: "{a1:b2:c3}", Parameter: []string{"a1", "b2", "c3"}},
				},
			},
		}

		for _, test := range tests {
			t.Run(fmt.Sprintf("%q", test.template), func(t *testing.T) {
				sections, err := Parse(test.template)
				require.NoError(t, err)
				assert.Equal(t, test.sections, sections)
			})
		}
	})

	t.Run("lonely trailing escape", func(t *testing.T) {
		t.Run("outside template param", func(t *testing.T) {
			_, err := Parse("\\")
			assert.Equal(t, errLonelyTrailingEscape, err)
		})

		t.Run("inside template param", func(t *testing.T) {
			_, err := Parse("{\\")
			assert.Equal(t, errLonelyTrailingEscape, err)
		})
	})

	t.Run("missing closing '}'", func(t *testing.T) {
		for _, s := range []string{"{", "{a:"} {
			t.Run(fmt.Sprintf("%q", s), func(t *testing.T) {
				_, err := Parse(s)
				assert.Equal(t, errUnclosedTrailingTemplateParam, err)
			})
		}
	})
}

func TestParseWithOptions(t *testing.T) {
	tests := []*struct {
		template string
		sections []Section
	}{
		{
			template: "",
			sections: []Section(nil),
		},
		{
			template: "[]",
			sections: []Section{
				{RawString: "[]", Parameter: []string{""}},
			},
		},
		{
			template: "[][]",
			sections: []Section{
				{RawString: "[]", Parameter: []string{""}},
				{RawString: "[]", Parameter: []string{""}},
			},
		},
		{
			template: "a",
			sections: []Section{
				{RawString: "a", String: "a"},
			},
		},
		{
			template: "a[][]",
			sections: []Section{
				{RawString: "a", String: "a"},
				{RawString: "[]", Parameter: []string{""}},
				{RawString: "[]", Parameter: []string{""}},
			},
		},
		{
			template: "[]a[]",
			sections: []Section{
				{RawString: "[]", Parameter: []string{""}},
				{RawString: "a", String: "a"},
				{RawString: "[]", Parameter: []string{""}},
			},
		},
		{
			template: "[][]a",
			sections: []Section{
				{RawString: "[]", Parameter: []string{""}},
				{RawString: "[]", Parameter: []string{""}},
				{RawString: "a", String: "a"},
			},
		},
		{
			template: "[]a1[]b2",
			sections: []Section{
				{RawString: "[]", Parameter: []string{""}},
				{RawString: "a1", String: "a1"},
				{RawString: "[]", Parameter: []string{""}},
				{RawString: "b2", String: "b2"},
			},
		},
		{
			template: "a1[]b2[]c3",
			sections: []Section{
				{RawString: "a1", String: "a1"},
				{RawString: "[]", Parameter: []string{""}},
				{RawString: "b2", String: "b2"},
				{RawString: "[]", Parameter: []string{""}},
				{RawString: "c3", String: "c3"},
			},
		},
		{
			template: "@[",
			sections: []Section{
				{RawString: "@[", String: "["},
			},
		},
		{
			template: "@@",
			sections: []Section{
				{RawString: "@@", String: "@"},
			},
		},
		{
			template: "@@@[",
			sections: []Section{
				{RawString: "@@@[", String: "@["},
			},
		},
		{
			template: "[@|]",
			sections: []Section{
				{RawString: "[@|]", Parameter: []string{"|"}},
			},
		},
		{
			template: "[a@|b]",
			sections: []Section{
				{RawString: "[a@|b]", Parameter: []string{"a|b"}},
			},
		},
		{
			template: "[@]]",
			sections: []Section{
				{RawString: "[@]]", Parameter: []string{"]"}},
			},
		},
		{
			template: "[@]@|@][]",
			sections: []Section{
				{RawString: "[@]@|@][]", Parameter: []string{"]|]["}},
			},
		},
		{
			template: "]",
			sections: []Section{
				{RawString: "]", String: "]"},
			},
		},
		{
			template: "|",
			sections: []Section{
				{RawString: "|", String: "|"},
			},
		},
		{
			template: "a1|b2]c3",
			sections: []Section{
				{RawString: "a1|b2]c3", String: "a1|b2]c3"},
			},
		},
		{
			template: "[|]",
			sections: []Section{
				{RawString: "[|]", Parameter: []string{"", ""}},
			},
		},
		{
			template: "[||]",
			sections: []Section{
				{RawString: "[||]", Parameter: []string{"", "", ""}},
			},
		},
		{
			template: "[a||]",
			sections: []Section{
				{RawString: "[a||]", Parameter: []string{"a", "", ""}},
			},
		},
		{
			template: "[|a|]",
			sections: []Section{
				{RawString: "[|a|]", Parameter: []string{"", "a", ""}},
			},
		},
		{
			template: "[||a]",
			sections: []Section{
				{RawString: "[||a]", Parameter: []string{"", "", "a"}},
			},
		},
		{
			template: "[a1|b2|c3]",
			sections: []Section{
				{RawString: "[a1|b2|c3]", Parameter: []string{"a1", "b2", "c3"}},
			},
		},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("%q", test.template), func(t *testing.T) {
			opts := &Options{
				ParamOpen:  '[',
				ParamClose: ']',
				ParamSplit: '|',
				Escape:     '@',
			}
			sections, err := ParseWithOptions(test.template, opts)
			require.NoError(t, err)
			assert.Equal(t, test.sections, sections)
		})
	}
}

func TestEscape(t *testing.T) {
	tests := []*struct {
		Str     string
		Escaped string
	}{
		{"", ""},
		{"ab c", "ab c"},
		{"{abc}", "\\{abc\\}"},
		{"{a:bc}", "\\{a\\:bc\\}"},
		{"\\{:}", "\\\\\\{\\:\\}"},
	}

	for _, test := range tests {
		t.Run(test.Str, func(t *testing.T) {
			escaped := Escape(test.Str)
			assert.Equal(t, test.Escaped, escaped)
		})
	}
}

func TestEscapeWithOptions(t *testing.T) {
	tests := []*struct {
		Str     string
		Escaped string
	}{
		{"", ""},
		{"ab c", "ab c"},
		{"[abc]", "@[abc@]"},
		{"[a|bc]", "@[a@|bc@]"},
		{"@[|]", "@@@[@|@]"},
	}

	for _, test := range tests {
		t.Run(test.Str, func(t *testing.T) {
			opts := &Options{
				ParamOpen:  '[',
				ParamClose: ']',
				ParamSplit: '|',
				Escape:     '@',
			}
			escaped := EscapeWithOptions(test.Str, opts)
			assert.Equal(t, test.Escaped, escaped)
		})
	}
}
