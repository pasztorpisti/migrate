package dir

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestFilenamePattern(t *testing.T) {
	t.Run("successful format and parse", func(t *testing.T) {
		const id = int64(42)
		const description = "test desc"

		tests := []*struct {
			Pattern             string
			Forward             string
			Backward            string
			Description         string
			IDSequence          bool
			HasDescription      bool
			OptionalDescription bool
			HasDirection        bool
		}{
			{
				Pattern:             "{id}{description:_:_}.{direction:fw:bw}.sql",
				Forward:             "0042_test_desc.fw.sql",
				Backward:            "0042_test_desc.bw.sql",
				Description:         description,
				IDSequence:          true,
				HasDescription:      true,
				OptionalDescription: true,
				HasDirection:        true,
			},
			{
				Pattern:             "{id}{description:_:_}.{direction:fw:bw}.sql",
				Forward:             "0042.fw.sql",
				Backward:            "0042.bw.sql",
				IDSequence:          true,
				HasDescription:      true,
				OptionalDescription: true,
				HasDirection:        true,
			},
			{
				Pattern:             "{description:_::_}{id}.{direction:fw:bw}.sql",
				Forward:             "test_desc_0042.fw.sql",
				Backward:            "test_desc_0042.bw.sql",
				Description:         description,
				IDSequence:          true,
				HasDescription:      true,
				OptionalDescription: true,
				HasDirection:        true,
			},
			{
				Pattern:             "{description:_::_}{id}.{direction:foreward:backward}.sql",
				Forward:             "0042.foreward.sql",
				Backward:            "0042.backward.sql",
				IDSequence:          true,
				HasDescription:      true,
				OptionalDescription: true,
				HasDirection:        true,
			},
			{
				Pattern:             "{direction:fwd:bwd}.{description:.::.}{id}.sql",
				Forward:             "fwd.test.desc.0042.sql",
				Backward:            "bwd.test.desc.0042.sql",
				Description:         description,
				IDSequence:          true,
				HasDescription:      true,
				OptionalDescription: true,
				HasDirection:        true,
			},
			{
				Pattern:             "{direction:fwd:bwd}.{description:_::_}{id}.sql",
				Forward:             "fwd.0042.sql",
				Backward:            "bwd.0042.sql",
				IDSequence:          true,
				HasDescription:      true,
				OptionalDescription: true,
				HasDirection:        true,
			},
			{
				Pattern:             "{direction}.{description:_::_}{id:unix_time:5}.sql",
				Forward:             defaultForwardStr + ".00042.sql",
				Backward:            defaultBackwardStr + ".00042.sql",
				IDSequence:          false,
				HasDescription:      true,
				OptionalDescription: true,
				HasDirection:        true,
			},
			{
				Pattern:             "{description}_{id:unix_time:5}.sql",
				Forward:             "test_desc_00042.sql",
				Backward:            "test_desc_00042.sql",
				Description:         description,
				IDSequence:          false,
				HasDescription:      true,
				OptionalDescription: false,
				HasDirection:        false,
			},
			{
				Pattern:             "{description}_{id:unix_time:5}.sql",
				Forward:             "_00042.sql",
				Backward:            "_00042.sql",
				IDSequence:          false,
				HasDescription:      true,
				OptionalDescription: false,
				HasDirection:        false,
			},
			{
				Pattern:             "{id:sequence:1}.sql",
				Forward:             "42.sql",
				Backward:            "42.sql",
				IDSequence:          true,
				HasDescription:      false,
				OptionalDescription: false,
				HasDirection:        false,
			},
			{
				Pattern:             "{id:unix_time:2}.sql",
				Forward:             "42.sql",
				Backward:            "42.sql",
				IDSequence:          false,
				HasDescription:      false,
				OptionalDescription: false,
				HasDirection:        false,
			},
			{
				Pattern:             "{id:unix_time:3}.sql",
				Forward:             "042.sql",
				Backward:            "042.sql",
				IDSequence:          false,
				HasDescription:      false,
				OptionalDescription: false,
				HasDirection:        false,
			},
		}

		for _, test := range tests {
			t.Run(test.Pattern, func(t *testing.T) {
				fp, err := parseFilenamePattern(test.Pattern)
				require.NoError(t, err)
				assert.Equal(t, test.IDSequence, fp.IDSequence)
				assert.Equal(t, test.HasDescription, fp.HasDescription)
				assert.Equal(t, test.OptionalDescription, fp.OptionalDescription)
				assert.Equal(t, test.HasDirection, fp.HasDirection)

				f := fp.FormatFilename(id, test.Description, true)
				assert.Equal(t, test.Forward, f)

				b := fp.FormatFilename(id, test.Description, false)
				assert.Equal(t, test.Backward, b)

				p, err := fp.ParseFilename(f)
				if assert.NoError(t, err) {
					assert.Equal(t, test.Description, p.Description)
					assert.Equal(t, id, p.ID.Number)
					assert.True(t, p.Forward)
				}
			})
		}
	})

	t.Run("{id} is required", func(t *testing.T) {
		_, err := parseFilenamePattern("woof")
		assert.Equal(t, err, errRequiredIDParameter)
	})

	t.Run("duplicate pattern parameter", func(t *testing.T) {
		tests := []*struct {
			Pattern string
			Error   error
		}{
			{
				Pattern: "{id}woof{id}",
				Error:   errDuplicateFilenamePatternParameter("id"),
			},
			{
				Pattern: "{description}woof{description:.}",
				Error:   errDuplicateFilenamePatternParameter("description"),
			},
			{
				Pattern: "{direction}woof{direction:f:b}",
				Error:   errDuplicateFilenamePatternParameter("direction"),
			},
		}

		for _, test := range tests {
			t.Run(test.Pattern, func(t *testing.T) {
				_, err := parseFilenamePattern(test.Pattern)
				assert.Equal(t, test.Error, err)
			})
		}
	})
}
