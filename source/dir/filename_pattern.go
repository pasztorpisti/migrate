package dir

import (
	"errors"
	"fmt"
	"github.com/pasztorpisti/migrate/template"
	"regexp"
	"strconv"
	"strings"
)

const (
	maxMigrationNameLength = 255
	defaultForwardStr      = "fw"
	defaultBackwardStr     = "fw"
)

type migrationID struct {
	// Number is the parsed form of the ID.
	Number int64

	// Names contains the valid names you can use to refer to this migration.
	// E.g.: In case of 0001_initial.sql you could use the following names:
	// "0001_initial.sql", "0001", "1"
	//
	// When the backward and forward migrations are split into separate files
	// the Names array doesn't contain the filename. In that case it has only
	// the ID and the zero padded ID.
	Names []string
}

func (o migrationID) String() string {
	return "migrationID(" + strconv.FormatInt(o.Number, 10) + ")"
}

// parsedFilenamePattern is the output of the parseFilenamePattern function.
type parsedFilenamePattern struct {
	IDSequence          bool
	HasDescription      bool
	OptionalDescription bool
	HasDirection        bool

	filenamePattern     string
	formatStr           string
	descriptionSpace    string
	descriptionPrefix   string
	descriptionSuffix   string
	forwardStr          string
	backwardStr         string
	formatArgs          []string
	regex               *regexp.Regexp
	idRegexIdx          int
	descriptionRegexIdx int
	directionRegexIdx   int
}

// FormatFilename returns a filename based on this parsedFilenamePattern and
// the provided template parameters.
// If parsedFilenamePattern.HasDescription == false then the description parameter is ignored.
// If parsedFilenamePattern.HasDirection == false then the forward parameter is ignored.
func (o *parsedFilenamePattern) FormatFilename(id int64, description string, forward bool) string {
	var args []interface{}
	for _, a := range o.formatArgs {
		switch a {
		case "id":
			args = append(args, id)
		case "direction":
			if forward {
				args = append(args, o.forwardStr)
			} else {
				args = append(args, o.backwardStr)
			}
		case "description":
			d := ""
			if !o.OptionalDescription || description != "" {
				d = strings.Replace(description, " ", o.descriptionSpace, -1)
				d = o.descriptionPrefix + d + o.descriptionSuffix
			}
			args = append(args, d)
		}
	}

	return fmt.Sprintf(o.formatStr, args...)
}

type parsedFilename struct {
	ID          migrationID
	Description string
	// Forward is always true when the filename pattern doesn't contain {direction}.
	Forward bool
}

func (o *parsedFilename) equals(other *parsedFilename) bool {
	return o.ID.Number == other.ID.Number && o.Description == other.Description
}

func (o *parsedFilenamePattern) ParseFilename(filename string) (*parsedFilename, error) {
	if len(filename) > maxMigrationNameLength {
		return nil, fmt.Errorf("migration file name is longer than the maximum=%v: %q", maxMigrationNameLength, filename)
	}
	a := o.regex.FindStringSubmatch(filename)
	if a == nil {
		return nil, fmt.Errorf("filename %q doesn't match the %q pattern", filename, o.filenamePattern)
	}

	zeroPaddedNumber := a[o.idRegexIdx]
	number, err := strconv.ParseInt(zeroPaddedNumber, 10, 64)
	if err != nil {
		return nil, err
	}

	id := migrationID{Number: number}
	if !o.HasDirection {
		id.Names = append(id.Names, filename)
	}
	id.Names = append(id.Names, zeroPaddedNumber)
	notZeroPadded := strconv.FormatInt(id.Number, 10)
	// The two are equal when zeroPaddedNumber has no leading zero digits.
	if notZeroPadded != zeroPaddedNumber {
		id.Names = append(id.Names, notZeroPadded)
	}

	description := ""
	if o.descriptionRegexIdx >= 0 {
		description = a[o.descriptionRegexIdx]
	}

	forward := o.directionRegexIdx < 0 || a[o.directionRegexIdx] == o.forwardStr

	return &parsedFilename{
		ID:          id,
		Description: strings.Replace(description, o.descriptionSpace, " ", -1),
		Forward:     forward,
	}, nil
}

var errRequiredIDParameter = errors.New("the filename pattern doesn't contain the required {id} parameter")

type errDuplicateFilenamePatternParameter string

func (o errDuplicateFilenamePatternParameter) Error() string {
	return fmt.Sprintf("duplicate {%v} parameter", string(o))
}

func parseFilenamePattern(filenamePattern string) (*parsedFilenamePattern, error) {
	sections, err := template.Parse(filenamePattern)
	if err != nil {
		return nil, err
	}

	idSequence := true

	forwardStr := defaultForwardStr
	backwardStr := defaultBackwardStr

	descriptionSpace := "_"
	descriptionPrefix := ""
	descriptionSuffix := ""
	optionalDescription := false

	hasID := false
	hasDirection := false
	hasDescription := false

	formatStr := ""
	var formatArgs []string
	regexStr := "^"

	for _, section := range sections {
		if section.IsRawString() {
			formatStr += strings.Replace(section.RawString, "%", "%%", -1)
			regexStr += regexp.QuoteMeta(section.RawString)
			continue
		}

		switch section.Parameter[0] {
		case "id":
			if hasID {
				return nil, errDuplicateFilenamePatternParameter("id")
			}
			hasID = true

			if len(section.Parameter) > 3 {
				return nil, fmt.Errorf("{id} has too many parameters: %q", section.Parameter)
			}
			if len(section.Parameter) >= 2 {
				switch idType := section.Parameter[1]; idType {
				case "sequence":
					idSequence = true
				case "unix_time":
					idSequence = false
				default:
					return nil, fmt.Errorf("found {id:<type>} with an invalid type: %v", idType)
				}
			}
			width := 4
			if len(section.Parameter) >= 3 {
				w, err := strconv.ParseInt(section.Parameter[2], 10, 0)
				if err != nil {
					return nil, fmt.Errorf("found {id:<type>:<width>} with invalid width: %v", section.Parameter[2])
				}
				if w < 1 {
					return nil, fmt.Errorf("width=%v must be greater than zero in {id:<type>:<width>}", w)
				}
				width = int(w)
			}

			formatStr += "%." + strconv.Itoa(width) + "d"
			formatArgs = append(formatArgs, "id")
			regexStr += `(?P<id>\d+)`

		case "direction":
			if hasDirection {
				return nil, errDuplicateFilenamePatternParameter("direction")
			}
			hasDirection = true

			switch len(section.Parameter) {
			case 1:
			case 3:
				forwardStr, backwardStr = section.Parameter[1], section.Parameter[2]
			default:
				return nil, fmt.Errorf("{section} and {section:<forwrad_string>:<backward_string>} are the only valid formats, got {%v}", strings.Join(section.Parameter, ":"))
			}

			formatStr += "%s"
			formatArgs = append(formatArgs, "direction")
			regexStr += fmt.Sprintf(`(?P<direction>%s|%s)`, regexp.QuoteMeta(forwardStr), regexp.QuoteMeta(backwardStr))

		case "description":
			if hasDescription {
				return nil, errDuplicateFilenamePatternParameter("description")
			}
			hasDescription = true

			n := len(section.Parameter)
			optionalDescription = n >= 3

			if n >= 2 {
				descriptionSpace = section.Parameter[1]
			}
			if n >= 3 {
				descriptionPrefix = section.Parameter[2]
			}
			if n >= 4 {
				descriptionSuffix = section.Parameter[3]
			}
			if n >= 5 {
				return nil, fmt.Errorf("{description} has too many parameters: %q", section.Parameter)
			}

			formatStr += "%s"
			formatArgs = append(formatArgs, "description")

			regexStr += `(` + regexp.QuoteMeta(descriptionPrefix) + `(?P<description>.*`
			if descriptionSuffix != "" {
				regexStr += `)` + regexp.QuoteMeta(descriptionSuffix)
			} else {
				regexStr += `?)`
			}
			regexStr += `)`
			if optionalDescription {
				regexStr += `?`
			}
		default:
			return nil, fmt.Errorf("invalid template parameter: {%v}", section.Parameter[0])
		}
	}

	regexStr += "$"

	if !hasID {
		return nil, errRequiredIDParameter
	}

	regex := regexp.MustCompile(regexStr)

	idRegexIdx := -1
	descriptionRegexIdx := -1
	directionRegexIdx := -1

	for i, name := range regex.SubexpNames() {
		switch name {
		case "id":
			idRegexIdx = i
		case "description":
			descriptionRegexIdx = i
		case "direction":
			directionRegexIdx = i
		}
	}

	return &parsedFilenamePattern{
		IDSequence:          idSequence,
		HasDescription:      hasDescription,
		OptionalDescription: optionalDescription,
		HasDirection:        hasDirection,

		filenamePattern:     filenamePattern,
		formatStr:           formatStr,
		descriptionSpace:    descriptionSpace,
		descriptionPrefix:   descriptionPrefix,
		descriptionSuffix:   descriptionSuffix,
		forwardStr:          forwardStr,
		backwardStr:         backwardStr,
		formatArgs:          formatArgs,
		regex:               regex,
		idRegexIdx:          idRegexIdx,
		descriptionRegexIdx: descriptionRegexIdx,
		directionRegexIdx:   directionRegexIdx,
	}, nil
}
