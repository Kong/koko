package resource

import (
	"fmt"
	"strings"
	"testing"

	v1 "github.com/kong/koko/internal/gen/grpc/kong/admin/model/v1"
	"github.com/kong/koko/internal/model"
	"github.com/kong/koko/internal/model/json/validation"
	"github.com/kong/koko/internal/model/json/validation/typedefs"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/reflect/protoreflect"
)

var (
	// All allowed tag delimiters.
	// NOTE: The space character has restrictions, see tests.
	tagDelimiters = ".: -_~"

	// All allowed tag characters. This also forms a valid tag.
	allowedTagChars = tagDelimiters + getAlphanumericChars()

	// Various characters denied in tags. This is not an exhaustive list, but the
	// commonly denied ASCII characters, along with UTF8 & extended ASCII characters.
	disallowedTagChars = "🙂♥ü`!@#$%^&*()=+[]\\{}|;'\",/<>?\t"
)

// Errors by result of JSON Schema validation.
var (
	errTagsMatchPattern = "must match pattern '" + typedefs.Tags.Items.Pattern + "'"
	errTagsMinLength    = "length must be >= 1, but got 0"
	errTagsMaxLength    = fmt.Sprintf(
		"length must be <= %d, but got %d",
		typedefs.Tags.Items.MaxLength,
		typedefs.Tags.Items.MaxLength+1,
	)
	errTagsMaxItems = fmt.Sprintf(
		"maximum %d items required, but found %d items",
		typedefs.Tags.MaxItems,
		typedefs.Tags.MaxItems+1,
	)
)

var validateTagTests = map[string][]*validateTagTest{
	"allowed characters": {
		{tags: []string{allowedTagChars}},
		{tags: []string{"tag"}},
	},
	"length check: min": {{
		tags:         []string{""},
		expectedErrs: []*v1.ErrorDetail{getValidationErr("tags[0]", errTagsMinLength)},
	}},
	"length check: max": {{
		tags:         []string{strings.Repeat("x", typedefs.Tags.Items.MaxLength+1)},
		expectedErrs: []*v1.ErrorDetail{getValidationErr("tags[0]", errTagsMaxLength)},
	}},
	"length check": {{
		tags: lo.Times(typedefs.Tags.MaxItems+1, func(i int) string {
			return fmt.Sprintf("tag-%d", i+1)
		}),
		expectedErrs: []*v1.ErrorDetail{getValidationErr("tags", errTagsMaxItems)},
	}},
	"unique items": {{
		tags:         []string{"tag1", "tag2", "tag2", "tag 1", "tag 2"},
		expectedErrs: []*v1.ErrorDetail{getValidationErr("tags", "items at index 1 and 2 are equal")},
	}},
	"multiple errors": {{
		tags: []string{"tag-1", "tag-2", "tag-2", "", "!"},
		expectedErrs: []*v1.ErrorDetail{
			getValidationErr("tags", "items at index 1 and 2 are equal"),
			getValidationErr("tags[3]", errTagsMinLength),
			getValidationErr("tags[4]", errTagsMatchPattern),
		},
	}},
}

type validateTagTest struct {
	tags         []string
	expectedErrs []*v1.ErrorDetail
}

func init() {
	// Generate tests to ensure all single characters pass validation (except for the space character).
	tt := make([]*validateTagTest, 0)
	validateTagTests["single characters"] = tt
	for _, char := range allowedTagChars {
		singleCharTest := &validateTagTest{tags: []string{string(char)}}
		if char == ' ' {
			singleCharTest.expectedErrs = []*v1.ErrorDetail{getValidationErr("tags[0]", errTagsMatchPattern)}
		}
		tt = append(tt, singleCharTest)
	}

	// Generate tests using all delimiters. This will create a wide range of permutations,
	// such as: `-tag`, `tag-`, `--tag`, `some-tag`, `-some-tag-`, `some--tag`, etc.
	tt = make([]*validateTagTest, 0)
	validateTagTests["delimiters"] = tt
	for _, char := range tagDelimiters {
		for _, strs := range [][]string{{"tag"}, {"some", "tag"}, {"some", "awesome", "tag"}} {
			// We'll create tests using one & two delimiters as the separator.
			for i := 1; i <= 2; i++ {
				sep := strings.Repeat(string(char), i)
				tag := strings.Join(strs, sep)

				// Only test words split by a delimiter.
				if len(strs) > 1 {
					tt = append(tt, &validateTagTest{tags: []string{tag}})
				}

				// Every delimiter but a space can be prefixed/suffixed.
				prefixSuffixTagTests := []*validateTagTest{
					{tags: []string{sep + tag}},
					{tags: []string{tag + sep}},
					{tags: []string{sep + tag + sep}},
				}
				if char == ' ' {
					for _, pt := range prefixSuffixTagTests {
						pt.expectedErrs = []*v1.ErrorDetail{getValidationErr("tags[0]", errTagsMatchPattern)}
					}
				}
				tt = append(tt, prefixSuffixTagTests...)
			}
		}
	}

	// Generate tests to ensure disallowed characters are rejected.
	tt = make([]*validateTagTest, 0)
	validateTagTests["disallowed characters"] = tt
	for _, char := range disallowedTagChars {
		tt = append(tt, &validateTagTest{
			tags:         []string{string(char)},
			expectedErrs: []*v1.ErrorDetail{getValidationErr("tags[0]", errTagsMatchPattern)},
		})
	}
}

// TestValidateTags ensures tag validation works as intended for all applicable registered resources.
func TestValidateTags(t *testing.T) {
	for name, groupedTests := range validateTagTests {
		t.Run(name, func(t *testing.T) {
			for _, typ := range model.AllTypes() {
				for _, tt := range groupedTests {
					testValidateTagForType(t, typ, tt)
				}
			}
		})
	}
}

func testValidateTagForType(t *testing.T, typ model.Type, tt *validateTagTest) {
	r, err := model.NewObject(typ)
	require.NoError(t, err)

	// Set tags on the resources by leveraging its Protobuf message.
	protoMsg := r.Resource()
	fieldDescriptors := protoMsg.ProtoReflect().Descriptor().Fields()
	var tagsField protoreflect.FieldDescriptor
	if tagsField = fieldDescriptors.ByName("tags"); tagsField == nil {
		return
	}
	tagFieldList := protoMsg.ProtoReflect().Mutable(tagsField).List()
	for _, tag := range tt.tags {
		tagFieldList.Append(protoreflect.ValueOfString(tag))
	}

	// Validate & only capture the tag validation errors, as everything else we are not concerned about.
	var tagErrs []*v1.ErrorDetail
	if err := validation.Validate(string(typ), protoMsg); err != nil {
		if err, ok := err.(validation.Error); ok {
			tagErrs = lo.Filter(err.Errs, func(err *v1.ErrorDetail, _ int) bool {
				// Tag validation errors will either have a field of `tags` or `tags[(index)]`.
				return strings.HasPrefix(err.Field, "tags")
			})
		}
	}

	if len(tt.expectedErrs) > 0 {
		require.NotEmpty(t, tagErrs)
		assert.Equal(t, tt.expectedErrs, tagErrs)
		return
	}

	require.Empty(t, tagErrs)
}

// getAlphanumericChars returns a string containing a-z, A-Z, & 0-9.
func getAlphanumericChars() string {
	var chars string
	for _, arr := range [][]rune{{'a', 'z'}, {'A', 'Z'}, {'0', '9'}} {
		for char := arr[0]; char <= arr[1]; char++ {
			chars += fmt.Sprintf("%c", char)
		}
	}
	return chars
}

// getValidationErr is a helper to builder a validation error for the given field & message.
func getValidationErr(field, message string) *v1.ErrorDetail {
	return &v1.ErrorDetail{
		Type:     v1.ErrorType_ERROR_TYPE_FIELD,
		Field:    field,
		Messages: []string{message},
	}
}
