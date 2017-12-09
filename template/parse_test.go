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
					{Parameter: []string{""}},
				},
			},
			{
				template: "{}{}",
				sections: []Section{
					{Parameter: []string{""}},
					{Parameter: []string{""}},
				},
			},
			{
				template: "a",
				sections: []Section{
					{RawString: "a"},
				},
			},
			{
				template: "a{}{}",
				sections: []Section{
					{RawString: "a"},
					{Parameter: []string{""}},
					{Parameter: []string{""}},
				},
			},
			{
				template: "{}a{}",
				sections: []Section{
					{Parameter: []string{""}},
					{RawString: "a"},
					{Parameter: []string{""}},
				},
			},
			{
				template: "{}{}a",
				sections: []Section{
					{Parameter: []string{""}},
					{Parameter: []string{""}},
					{RawString: "a"},
				},
			},
			{
				template: "{}a1{}b2",
				sections: []Section{
					{Parameter: []string{""}},
					{RawString: "a1"},
					{Parameter: []string{""}},
					{RawString: "b2"},
				},
			},
			{
				template: "a1{}b2{}c3",
				sections: []Section{
					{RawString: "a1"},
					{Parameter: []string{""}},
					{RawString: "b2"},
					{Parameter: []string{""}},
					{RawString: "c3"},
				},
			},
			{
				template: "\\{",
				sections: []Section{
					{RawString: "{"},
				},
			},
			{
				template: "\\\\",
				sections: []Section{
					{RawString: "\\"},
				},
			},
			{
				template: "\\r\\n\\t",
				sections: []Section{
					{RawString: "\r\n\t"},
				},
			},
			{
				template: "\\\\\\{",
				sections: []Section{
					{RawString: "\\{"},
				},
			},
			{
				template: "{\\:}",
				sections: []Section{
					{Parameter: []string{":"}},
				},
			},
			{
				template: "{a\\:b}",
				sections: []Section{
					{Parameter: []string{"a:b"}},
				},
			},
			{
				template: "{\\}}",
				sections: []Section{
					{Parameter: []string{"}"}},
				},
			},
			{
				template: "{\\}\\:\\}{}",
				sections: []Section{
					{Parameter: []string{"}:}{"}},
				},
			},
			{
				template: "}",
				sections: []Section{
					{RawString: "}"},
				},
			},
			{
				template: ":",
				sections: []Section{
					{RawString: ":"},
				},
			},
			{
				template: "a1:b2}c3",
				sections: []Section{
					{RawString: "a1:b2}c3"},
				},
			},
			{
				template: "{:}",
				sections: []Section{
					{Parameter: []string{"", ""}},
				},
			},
			{
				template: "{::}",
				sections: []Section{
					{Parameter: []string{"", "", ""}},
				},
			},
			{
				template: "{a::}",
				sections: []Section{
					{Parameter: []string{"a", "", ""}},
				},
			},
			{
				template: "{:a:}",
				sections: []Section{
					{Parameter: []string{"", "a", ""}},
				},
			},
			{
				template: "{::a}",
				sections: []Section{
					{Parameter: []string{"", "", "a"}},
				},
			},
			{
				template: "{a1:b2:c3}",
				sections: []Section{
					{Parameter: []string{"a1", "b2", "c3"}},
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
				{Parameter: []string{""}},
			},
		},
		{
			template: "[][]",
			sections: []Section{
				{Parameter: []string{""}},
				{Parameter: []string{""}},
			},
		},
		{
			template: "a",
			sections: []Section{
				{RawString: "a"},
			},
		},
		{
			template: "a[][]",
			sections: []Section{
				{RawString: "a"},
				{Parameter: []string{""}},
				{Parameter: []string{""}},
			},
		},
		{
			template: "[]a[]",
			sections: []Section{
				{Parameter: []string{""}},
				{RawString: "a"},
				{Parameter: []string{""}},
			},
		},
		{
			template: "[][]a",
			sections: []Section{
				{Parameter: []string{""}},
				{Parameter: []string{""}},
				{RawString: "a"},
			},
		},
		{
			template: "[]a1[]b2",
			sections: []Section{
				{Parameter: []string{""}},
				{RawString: "a1"},
				{Parameter: []string{""}},
				{RawString: "b2"},
			},
		},
		{
			template: "a1[]b2[]c3",
			sections: []Section{
				{RawString: "a1"},
				{Parameter: []string{""}},
				{RawString: "b2"},
				{Parameter: []string{""}},
				{RawString: "c3"},
			},
		},
		{
			template: "@[",
			sections: []Section{
				{RawString: "["},
			},
		},
		{
			template: "@@",
			sections: []Section{
				{RawString: "@"},
			},
		},
		{
			template: "@@@[",
			sections: []Section{
				{RawString: "@["},
			},
		},
		{
			template: "[@|]",
			sections: []Section{
				{Parameter: []string{"|"}},
			},
		},
		{
			template: "[a@|b]",
			sections: []Section{
				{Parameter: []string{"a|b"}},
			},
		},
		{
			template: "[@]]",
			sections: []Section{
				{Parameter: []string{"]"}},
			},
		},
		{
			template: "[@]@|@][]",
			sections: []Section{
				{Parameter: []string{"]|]["}},
			},
		},
		{
			template: "]",
			sections: []Section{
				{RawString: "]"},
			},
		},
		{
			template: "|",
			sections: []Section{
				{RawString: "|"},
			},
		},
		{
			template: "a1|b2]c3",
			sections: []Section{
				{RawString: "a1|b2]c3"},
			},
		},
		{
			template: "[|]",
			sections: []Section{
				{Parameter: []string{"", ""}},
			},
		},
		{
			template: "[||]",
			sections: []Section{
				{Parameter: []string{"", "", ""}},
			},
		},
		{
			template: "[a||]",
			sections: []Section{
				{Parameter: []string{"a", "", ""}},
			},
		},
		{
			template: "[|a|]",
			sections: []Section{
				{Parameter: []string{"", "a", ""}},
			},
		},
		{
			template: "[||a]",
			sections: []Section{
				{Parameter: []string{"", "", "a"}},
			},
		},
		{
			template: "[a1|b2|c3]",
			sections: []Section{
				{Parameter: []string{"a1", "b2", "c3"}},
			},
		},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("%q", test.template), func(t *testing.T) {
			opts := &ParseOptions{
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
