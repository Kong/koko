package seed

import (
	"errors"
	"net/http"

	"github.com/kong/koko/internal/model"
	"github.com/kong/koko/internal/resource"
	"google.golang.org/protobuf/proto"
)

// ModifyHTTPRequestFunc defines a function that can manipulate the passed in HTTP
// request for the given type. The function must be safe for concurrent use.
type ModifyHTTPRequestFunc func(model.Type, *http.Request) error

// OptionsBuilder allows seeder options to be built, used for each seed call.
type OptionsBuilder struct{ opts *Options }

// Options defines the non-exported options that the Seeder seed
// methods can take in. Must be formed using NewOptionsBuilder().
type Options struct {
	// Number of resources to create for a given type.
	count int

	// Various tag options used for resource creation.
	tags tagOptions

	// Seed-specific errors (defined as seed.Err* vars) that can be ignored.
	ignoredErrs []error

	// Functions that can modify the HTTP requests that the seeder generates.
	modifyReqFuncs []ModifyHTTPRequestFunc

	// Overrides any default new resource functions set on the seeder.
	newResourceFuncs map[model.Type]NewResourceFunc
}

type tagOptions struct {
	// When true, a unique tag will be created for every resource, starting
	// from "tag-1". When false, tags will be created randomly.
	useIncremental bool

	// Maximum number of tags to create.
	count int

	// When true, allows for no tags at all on a resource. When false, at least one tag will always exist.
	allowEmpty bool

	// Optional tags to set on every resource.
	staticTags []string
}

// NewOptionsBuilder instantiates a new instance of an OptionsBuilder.
func NewOptionsBuilder() *OptionsBuilder {
	return &OptionsBuilder{opts: &Options{
		ignoredErrs:      make([]error, 0),
		modifyReqFuncs:   make([]ModifyHTTPRequestFunc, 0),
		newResourceFuncs: make(map[model.Type]NewResourceFunc),
	}}
}

// WithResourceCount sets the number of resources to create for
// each type being seeded. Must be a non-zero, positive number.
func (b *OptionsBuilder) WithResourceCount(count int) *OptionsBuilder {
	b.opts.count = count
	return b
}

// WithRandomTagCount sets the max number of tags to be created, starting from "tag-1".
//
// When zero (or option is redacted), no tags will be set on the resources.
// When the field is non-zero, random tags will be created for each object
// (using NewSeederOpts.RandSourceFunc), up to the given amount.
//
// The allowEmpty flag controls whether resources can be created with no tags.
//
// E.g.: If `2` is set & allowEmpty is true, a resource can have any of the
// given combinations of tags: (no tags), `tag-1`, `tag-2`, `tag-1, tag-2`.
func (b *OptionsBuilder) WithRandomTagCount(count int, allowEmpty bool) *OptionsBuilder {
	b.opts.tags.count, b.opts.tags.allowEmpty = count, allowEmpty

	// Disable the use of incremental tags, as it's allowed to be random now.
	b.opts.tags.useIncremental = false

	return b
}

// WithIncrementalTags sets the specified number of unique tags on each resource, starting from "tag-1".
//
// The count argument must be a non-zero, positive number, and it represents the max number of tags that
// can be created on a resource. If it's set to one, each resource will only have one tag. If it's set
// to any higher number, resources can have more than one tag, but each tag is only used once.
//
// E.g.: If ten resources are being seeded, with the count set to one, "tag-1"-"tag-10" will be used.
// In the event the seeder is run multiple times, it will still ensure unique tags (as long as this
// option is specified).
func (b *OptionsBuilder) WithIncrementalTags(count int) *OptionsBuilder {
	b.opts.tags.count, b.opts.tags.useIncremental, b.opts.tags.allowEmpty = count, true, false
	return b
}

// WithStaticTags sets optional tags to set on each resource. When used, this overrides
// all default tag generation behavior (as in, no other tags will be set but these).
func (b *OptionsBuilder) WithStaticTags(tags []string) *OptionsBuilder {
	b.opts.tags.count, b.opts.tags.staticTags = len(tags), tags
	return b
}

// WithIgnoredErrors can be used to skip specific errors when seeding all types with
// the Seeder.SeedAllTypes() call. For example,the ErrRequiredFieldMissing error can
// be set to continue seeding when a type is missing a field that is expected.
func (b *OptionsBuilder) WithIgnoredErrors(errs ...error) *OptionsBuilder {
	b.opts.ignoredErrs = append(b.opts.ignoredErrs, errs...)
	return b
}

// WithModifyHTTPRequestFuncs can be used to modify the HTTP request to create
// a resource. The functions are executed in the order they are provided.
func (b *OptionsBuilder) WithModifyHTTPRequestFuncs(modifyRequestFuncs ...ModifyHTTPRequestFunc) *OptionsBuilder {
	b.opts.modifyReqFuncs = append(b.opts.modifyReqFuncs, modifyRequestFuncs...)
	return b
}

// WithNewResourceFunc overrides a default NewResourceFunc for a single seed call.
//
// When inheritDefault is true, the new resource function on the seeder will be executed before, and
// if said function is not defined, it will then default to those defined in DefaultNewResourceFuncs.
func (b *OptionsBuilder) WithNewResourceFunc(typ model.Type, inheritDefault bool, f NewResourceFunc) *OptionsBuilder {
	if b.opts.newResourceFuncs == nil {
		b.opts.newResourceFuncs = make(map[model.Type]NewResourceFunc, 1)
	}
	b.opts.newResourceFuncs[typ] = func(s Seeder, m proto.Message, i int) error {
		if inheritDefault {
			var defaultFunc NewResourceFunc
			var ok bool
			if defaultFunc, ok = s.(*seeder).opts.NewResourceFuncs[resource.TypeService]; !ok {
				defaultFunc = DefaultNewResourceFuncs[typ]
			}
			if defaultFunc != nil {
				if err := defaultFunc(s, m, i); err != nil {
					return err
				}
			}
		}
		return f(s, m, i)
	}
	return b
}

// Build validates the passed in options & returns the generated options, that can be passed to the seed methods.
func (b *OptionsBuilder) Build() (*Options, error) {
	if b.opts.count < 0 {
		return nil, errors.New("count passed to WithResourceCount() must be greater than zero")
	}

	if b.opts.tags.count < 0 {
		if b.opts.tags.useIncremental {
			return nil, errors.New("count passed to WithIncrementalTags() must be greater than zero")
		}
		return nil, errors.New("count passed to WithRandomTagCount() must be greater than zero")
	}

	return b.opts, nil
}
