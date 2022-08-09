package seed

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"hash/fnv"
	"io"
	"math/rand"
	"net/http"
	"net/http/httputil"
	"runtime"
	"sync"
	"time"

	"github.com/kong/koko/internal/json"
	"github.com/kong/koko/internal/model"
	"github.com/kong/koko/internal/model/json/extension"
	"github.com/kong/koko/internal/model/json/schema"
	"github.com/samber/lo"
	"go.uber.org/atomic"
	"golang.org/x/sync/errgroup"
	"google.golang.org/genproto/googleapis/api/annotations"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
	"google.golang.org/protobuf/types/descriptorpb"
)

// ErrRequiredFieldMissing is returned during a seed call when a type does not have a
// required field to properly seed the resource. For example, this can happen when
// tags should be created, however, the resource is missing the `tags` field.
var ErrRequiredFieldMissing = errors.New("field does not exist for type")

// DefaultRandSourceFunc defines the default function
// used for determining the random source based on a given type.
var DefaultRandSourceFunc = func(typ model.Type) (rand.Source, error) {
	// The randomness source will be set based on the FNV-1A hash of the type's name. This allows for
	// deterministic randomness, which is helpful when writing tests against resources created.
	h := fnv.New32a()
	if _, err := h.Write([]byte(typ)); err != nil {
		return nil, err
	}
	return rand.NewSource(int64(h.Sum32())), nil
}

// Seeder handles creating test resources, usually used to ease integration
// testing. All methods defined on the interface are safe for concurrent use.
type Seeder interface {
	// Results returns all resources that were created by result of running the seeder (in the event
	// the seeder was run multiple times, all resources created across each run will be returned).
	Results() *Results

	// AllResourceInfo returns extended information about all registered resources that are capable of being seeded.
	//
	// Resources that cannot be seeded include:
	// - Resources missing `$["x-koko-config"].resourceAPIPath` on the JSON schema.
	// - Resources missing a `POST /v1/(resource)` HTTP binding defined on the gRPC service.
	AllResourceInfo() []*ResourceInfo

	// ResourceInfoByType returns extended information for the provided resource.
	ResourceInfoByType(typ model.Type) *ResourceInfo

	// SeedAllTypes creates resources, based on the provided options, for all registered
	// resources. For more information, read the documentation for Seeder.Seed().
	SeedAllTypes(context.Context, *Options) (*Results, error)

	// Seed creates the given number of resources for the provided resource. The newly created results will be
	// returned & added to the underlining seeder object.
	//
	// In the event this method is called multiple times, any newly created resources will be appended to the
	// seeder instance. This will affect the results when calling seeder.Results().
	//
	// The ErrNoResourcePath & ErrRequiredFieldMissing errors may be returned as a wrapped error, and can be
	// checked using errors.Is(). See documentation for the mentioned errors in this file for more detail.
	Seed(context.Context, *Options, ...model.Type) (*Results, error)
}

// seeder handles seeding via the REST API.
type seeder struct {
	opts                   *NewSeederOpts
	resourcesToCreateFirst []model.Type
	resourcesToCreateLast  []model.Type
	resourceInfoByType     map[model.Type]*ResourceInfo

	// Contains the ResourceInfo for resourcesToCreateFirst and then resourcesToCreateLast.
	orderedResources []*ResourceInfo

	// Keep track of all resources that were created (even across multiple seed calls).
	results Results

	// Ensure unique tags (even across multiple seed calls) for a specific type when requested.
	lastTagNumberForType   map[model.Type]*atomic.Uint64
	lastTagNumberForTypeMu sync.Mutex

	// Each type has its own source of random, used for things like tag generation.
	randByType   map[model.Type]*rand.Rand
	randByTypeMu sync.RWMutex
}

var _ Seeder = &seeder{}

// NewSeederOpts defines the configuration used to instantiate a new seeder.
// The config should not be updated after it is set on a seeder.
type NewSeederOpts struct {
	// The URL to the control plane RESTful API, with the scheme. e.g.: http://127.0.0.1:8080
	URL string

	// Optional max amount of resources to be seeded at one time. When zero,
	// defaults to runtime.NumCPU(). Otherwise, must be a positive number.
	ConcurrencyLimit int

	// Optional HTTP client to use. When nil, will default to http.DefaultClient.
	HTTPClient *http.Client

	// Optional function that handles setting the random source used for deterministically generating random
	// data, like tags, based on the given type name. When nil, defaults to DefaultRandSourceFunc.
	RandSourceFunc func(typ model.Type) (rand.Source, error)

	// Optional map allowing to set custom NewResourceFunc functions, that will be called for the associated
	// type when creating the resource. When empty, will default to DefaultNewResourceFuncs.
	NewResourceFuncs map[model.Type]NewResourceFunc

	// Optional Protobuf registry, used for HTTP rule binding verification.
	// When nil, defaults to protoregistry.GlobalFiles.
	ProtoRegistry *protoregistry.Files
}

// New instantiates a seeder with the given config.
//
// In the event new resources are registered, you must instantiate
// a new seeder for those resources to be picked up.
func New(opts NewSeederOpts) (Seeder, error) {
	if opts.URL == "" {
		return nil, errors.New("must provide an API URL")
	}

	if opts.ConcurrencyLimit < 0 {
		return nil, errors.New("concurrency limit must not be negative")
	} else if opts.ConcurrencyLimit == 0 {
		opts.ConcurrencyLimit = runtime.NumCPU()
	}

	if opts.HTTPClient == nil {
		opts.HTTPClient = http.DefaultClient
	}

	if opts.RandSourceFunc == nil {
		opts.RandSourceFunc = DefaultRandSourceFunc
	}

	if len(opts.NewResourceFuncs) == 0 {
		opts.NewResourceFuncs = DefaultNewResourceFuncs
	}

	if opts.ProtoRegistry == nil {
		opts.ProtoRegistry = protoregistry.GlobalFiles
	}

	s := &seeder{
		opts: &opts,
		results: Results{
			all:    make([]*Result, 0),
			byType: make(map[model.Type][]*Result),
		},
		resourceInfoByType:   make(map[model.Type]*ResourceInfo),
		lastTagNumberForType: make(map[model.Type]*atomic.Uint64),
		randByType:           make(map[model.Type]*rand.Rand),
	}
	if err := s.generateOrderedResourceList(); err != nil {
		return nil, fmt.Errorf("unable to determine the order of which resources should be created: %w", err)
	}

	return s, nil
}

// Results implements the Seeder.Results interface.
func (s *seeder) Results() *Results { return &s.results }

// AllResourceInfo implements the Seeder.AllResourceInfo interface.
func (s *seeder) AllResourceInfo() []*ResourceInfo { return s.orderedResources }

// ResourceInfoByType implements the Seeder.ResourceInfoByType interface.
func (s *seeder) ResourceInfoByType(typ model.Type) *ResourceInfo { return s.resourceInfoByType[typ] }

// SeedAllTypes implements the Seeder.SeedAllTypes interface.
func (s *seeder) SeedAllTypes(ctx context.Context, opts *Options) (*Results, error) {
	return s.seedTypes(ctx, opts, lo.Keys(s.resourceInfoByType)...)
}

// Seed implements the Seeder.Seed interface.
func (s *seeder) Seed(ctx context.Context, opts *Options, typesToSeed ...model.Type) (*Results, error) {
	return s.seedTypes(ctx, opts, typesToSeed...)
}

// seedTypes handles seeding multiple resource types concurrently.
func (s *seeder) seedTypes(ctx context.Context, opts *Options, typesToSeed ...model.Type) (*Results, error) {
	// Used to keep track of the results created during this specific seed.
	newResults := &Results{
		all:    make([]*Result, 0),
		byType: make(map[model.Type][]*Result, len(s.orderedResources)),
	}

	g := errgroup.Group{}
	g.SetLimit(s.opts.ConcurrencyLimit)
	fn := func(typ model.Type) func() error {
		return func() error {
			results, err := s.seedType(ctx, typ, opts)
			if err != nil {
				// Continue seeding in the event the caller asked to ignore this error.
				for _, ignoredErr := range opts.ignoredErrs {
					if errors.Is(err, ignoredErr) {
						return nil
					}
				}

				return fmt.Errorf("unable to seed type %q: %w", typ, err)
			}
			newResults.Add(typ, results)
			return nil
		}
	}

	// Create test resources that don't require any dependencies.
	for _, typ := range lo.Intersect(s.resourcesToCreateFirst, typesToSeed) {
		g.Go(fn(typ))
	}
	if err := g.Wait(); err != nil {
		return nil, err
	}

	// Create test resources that require dependencies.
	for _, typ := range lo.Intersect(s.resourcesToCreateLast, typesToSeed) {
		g.Go(fn(typ))
	}
	if err := g.Wait(); err != nil {
		return nil, err
	}

	return newResults, nil
}

// seedTypes handles seeding a single resource type.
func (s *seeder) seedType(ctx context.Context, typeToSeed model.Type, opts *Options) (*Results, error) {
	typ, ok := s.resourceInfoByType[typeToSeed]
	if !ok {
		return nil, fmt.Errorf("cannot find type %q", typeToSeed)
	}

	createdResources, err := s.createResourcesForType(ctx, opts, typ)
	if err != nil {
		return nil, err
	}

	// Add in the results to the seeder, so that seeder.Seed() can
	// be called multiple times & store all created resources.
	s.results.Add(typ.Name, createdResources)

	return createdResources, nil
}

// createResourcesForType handles generating the resources to seed
// and issuing the proper HTTP calls to create said resources.
func (s *seeder) createResourcesForType(ctx context.Context, opts *Options, typ *ResourceInfo) (*Results, error) {
	idField := typ.fieldDescriptors.ByName("id")

	// In the event we need to set tags, we'll need to know what field to set them on.
	var tagsField protoreflect.FieldDescriptor
	if tagsField = typ.fieldDescriptors.ByName("tags"); opts.tags.count > 0 && tagsField == nil {
		return nil, fmt.Errorf(`"tags" field missing for type %q: %w`, typ.Name, ErrRequiredFieldMissing)
	}

	// Allow for seemingly random tags, but deterministically for each
	// resource type (when the default random source function is used).
	r, err := s.getRandForType(typ.Name)
	if err != nil {
		return nil, err
	}

	// In the event the tag options have changed since the last seed run, we'll reset the underlining counter keeping
	// track of the tags used. We'll ensure this is only reset once, as concurrent seed runs are supported.
	lastTagNumber := s.resetLastTagNumber(typ.Name, &opts.tags)

	// Allow callers to override the default new resource functions defined on the seeder.
	newResourceFuncs := lo.Assign(s.opts.NewResourceFuncs, opts.newResourceFuncs)

	// Create the desired amount of resources.
	createdResources := &Results{all: make([]*Result, opts.count)}
	createdResources.byType = map[model.Type][]*Result{typ.Name: createdResources.all}
	currentResultsCount := len(s.Results().ByType(typ.Name).All())
	for i := 0; i < opts.count; i++ {
		protoMsg := proto.Clone(typ.object.Resource())
		createdResource := &Result{Tags: make([]string, 0), Resource: protoMsg}
		createdResources.all[i] = createdResource

		// Set random tags (e.g.: when `s.opts.TagsToCreate == 2`, any
		// combination of `tag-1` or `tag-2`, including no tags at all).
		if createdResource.Tags = s.generateTags(r, lastTagNumber, opts); len(createdResource.Tags) > 0 {
			tagFieldList := protoMsg.ProtoReflect().Mutable(tagsField).List()
			for _, tag := range createdResource.Tags {
				tagFieldList.Append(protoreflect.ValueOfString(tag))
			}
		}

		// Set any required fields for this resource.
		if f := newResourceFuncs[typ.Name]; f != nil {
			if err := f(s, protoMsg, i+currentResultsCount); err != nil {
				return nil, err
			}
		}

		// Create the resource via the REST API.
		ctxTimeout, cancel := context.WithTimeout(ctx, 30*time.Second) //nolint:gomnd
		defer cancel()
		req, err := s.getHTTPRequest(ctxTimeout, typ, protoMsg, opts)
		if err != nil {
			return nil, err
		}
		if err := s.doHTTPRequest(req, protoMsg); err != nil {
			return nil, err
		}

		// We're letting the API automatically generate the ID, so we'll need to fetch what was generated.
		if createdResource.ID, err = getResourceIDFromProto(typ, idField, protoMsg); err != nil {
			return nil, err
		}
	}

	return createdResources, nil
}

// generateTags handles creating tags, either randomly or incrementally based on the provided options.
func (s *seeder) generateTags(r *rand.Rand, lastNum *atomic.Uint64, opts *Options) []string {
	var tags []string
	for j := 1; j <= opts.tags.count; j++ {
		// Use a 50% chance whether the tag will be set on the resource.
		if r.Int()%2 == 0 {
			// Ensure there is at least one tag when we're not allowed to create resources with empty tags.
			if opts.tags.allowEmpty || j != opts.tags.count || len(tags) != 0 {
				continue
			}
		}
		// By default, use random tags, otherwise, use incremental tags.
		tagNumber := uint64(j)
		if lastNum != nil {
			tagNumber = lastNum.Add(1)
		}
		tagName := fmt.Sprintf("tag-%d", tagNumber)
		tags = append(tags, tagName)
	}
	return tags
}

// getHTTPRequest forms an HTTP call to create a given resource, optionally
// modifying the request with any caller-provided modify request functions.
func (s *seeder) getHTTPRequest(
	ctx context.Context,
	typ *ResourceInfo,
	msg proto.Message,
	opts *Options,
) (*http.Request, error) {
	requestJSON, err := json.ProtoJSONMarshal(msg)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(
		ctx,
		"POST",
		s.opts.URL+typ.createEndpoint(),
		bytes.NewBuffer(requestJSON),
	)
	if err != nil {
		return nil, err
	}
	// Callers of the seeder are allowed to manipulate the requests that we generate.
	for _, f := range opts.modifyReqFuncs {
		if err := f(typ.Name, req); err != nil {
			return nil, err
		}
	}

	return req, nil
}

// doHTTPRequest issues a create resource HTTP call & alters the passed in `msg` argument with the updated resource.
func (s *seeder) doHTTPRequest(req *http.Request, msg proto.Message) error {
	requestStr, _ := httputil.DumpRequestOut(req, true)
	resp, err := s.opts.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("unable to create resource via the API: %w", err)
	}
	defer resp.Body.Close()

	// Replace the resource with the API response, as the API can automatically set fields & so
	// forth. This assumes the `$.item` key contains the created resource in the response.
	respJSON, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf(
			"received %d HTTP status code when creating %T resource:\n\n"+
				"Request:\n%s\n\n----\n\n"+
				"Response:\n%s",
			resp.StatusCode,
			msg,
			string(requestStr),
			string(respJSON),
		)
	}

	var respWithItem struct {
		Item json.RawMessage `json:"item"`
	}
	if err := json.Unmarshal(respJSON, &respWithItem); err != nil {
		return fmt.Errorf("unable to unmarshal create resource API respose: %w", err)
	}

	return protojson.Unmarshal(respWithItem.Item, msg)
}

// resetLastTagNumber keeps track of any tag option changes since the last seed run, and
// sets/removes an internal tag counter in order to ensure unique tags when required.
func (s *seeder) resetLastTagNumber(typ model.Type, opts *tagOptions) *atomic.Uint64 {
	var lastTagNumber *atomic.Uint64
	s.lastTagNumberForTypeMu.Lock()
	var ok bool
	if lastTagNumber, ok = s.lastTagNumberForType[typ]; opts.useIncremental != ok {
		if opts.useIncremental {
			lastTagNumber = &atomic.Uint64{}
			s.lastTagNumberForType[typ] = lastTagNumber
		} else {
			delete(s.lastTagNumberForType, typ)
		}
	}
	s.lastTagNumberForTypeMu.Unlock()
	if !opts.useIncremental {
		lastTagNumber = nil
	}
	return lastTagNumber
}

// getRandForType returns the source of random for a given type.
func (s *seeder) getRandForType(typ model.Type) (*rand.Rand, error) {
	// In the event the seeder was already called for this resource type, we'll grab the pre-computed random source.
	s.randByTypeMu.RLock()
	r, ok := s.randByType[typ]
	s.randByTypeMu.RUnlock()
	if ok {
		return r, nil
	}

	randSrc, err := s.opts.RandSourceFunc(typ)
	if err != nil {
		return nil, fmt.Errorf("unable to get random source for resource type %q: %w", typ, err)
	}

	// We're okay with the additional cost of acquiring a write-lock, as this will only be done once for each type.
	s.randByTypeMu.Lock()
	defer s.randByTypeMu.Unlock()
	s.randByType[typ] = rand.New(randSrc) //nolint:gosec

	return s.randByType[typ], nil
}

// generateOrderedResourceList determines in what order to create
// the resources & set resource metadata on the seeder object.
func (s *seeder) generateOrderedResourceList() error {
	// These resources require other resources in order to be created, so we'll create them last.
	//
	// TODO(tjasko): If in the event a resource requires two or more dependent resources to be
	//  created, we'll need to make some changes in order for that to be supported. Right now,
	//  we're leaning on the side of simplicity.
	s.resourcesToCreateLast = resourcesToCreateLast

	// Any remaining resources will be created prior to those that require dependencies.
	s.resourcesToCreateFirst, _ = lo.Difference(model.AllTypes(), s.resourcesToCreateLast)

	// Fetch all POST endpoints that all registered gRPC services expose.
	createResourceEndpoints := s.getCreateResourceEndpoints()

	var err error
	orderedTypes := s.resourcesToCreateFirst
	for _, t := range append(orderedTypes, s.resourcesToCreateLast...) {
		// Fill in some metadata for each type.
		typ := &ResourceInfo{Name: t}
		if typ.Schema, err = schema.Get(string(typ.Name)); err != nil {
			return err
		}
		configExtName := (&extension.Config{}).Name()
		if e := typ.Schema.Extensions; e != nil && e[configExtName] != nil {
			var ok bool
			if typ.JSONSchemaConfig, ok = e[configExtName].(*extension.Config); !ok {
				// Should not happen, but just a sanity check.
				return fmt.Errorf(
					"unexpected JSON schema %s custom config, expected: %T, got: %T",
					configExtName,
					&extension.Config{},
					e[configExtName],
				)
			}
		}

		// Ensure this type has an underlining `POST /v1/(resource)` API
		// endpoint, or else don't add it to the list of types we can seed.
		if typ.JSONSchemaConfig == nil || !createResourceEndpoints[typ.createEndpoint()] {
			continue
		}

		// Save the underlining Protobuf message to the resource info, so that we can clone it later.
		if typ.object, err = model.NewObject(typ.Name); err != nil {
			return err
		}
		typ.fieldDescriptors = typ.object.Resource().ProtoReflect().Descriptor().Fields()

		// Save the resource info for later use, as callers of the
		// Seeder can use it to determine business logic as well.
		s.orderedResources, s.resourceInfoByType[typ.Name] = append(s.orderedResources, typ), typ
	}

	return nil
}

// getCreateResourceEndpoints parses the registered gRPC services and outputs a map of POST endpoints
// that exist. This map can then be checked against to see if a desired POST endpoint exists.
func (s *seeder) getCreateResourceEndpoints() map[string]bool {
	endpoints := make(map[string]bool)

	var addPostBindingFromHTTPRule func(rule *annotations.HttpRule)
	addPostBindingFromHTTPRule = func(rule *annotations.HttpRule) {
		if rule == nil {
			return
		}
		if binding := rule.GetPost(); binding != "" {
			endpoints[binding] = true
		}
		for _, binding := range rule.AdditionalBindings {
			addPostBindingFromHTTPRule(binding)
		}
	}

	s.opts.ProtoRegistry.RangeFiles(func(descriptor protoreflect.FileDescriptor) bool {
		services := descriptor.Services()
		for i := 0; i < services.Len(); i++ {
			methods := services.Get(i).Methods()
			for i := 0; i < methods.Len(); i++ {
				messageOptions, ok := methods.Get(i).Options().(*descriptorpb.MethodOptions)
				if !ok {
					continue
				}
				if httpRule, ok := proto.GetExtension(messageOptions, annotations.E_Http).(*annotations.HttpRule); ok {
					addPostBindingFromHTTPRule(httpRule)
				}
			}
		}
		return true
	})

	return endpoints
}

// getResourceIDFromProto fetches the automatically generated ID for a given resource.
func getResourceIDFromProto(
	typ *ResourceInfo,
	idField protoreflect.FieldDescriptor,
	msg proto.Message,
) (string, error) {
	if idField != nil {
		return msg.ProtoReflect().Get(idField).String(), nil
	}

	// Not all resources use the "id" field for its primary key. In the event we can't
	// find it, we'll call the underlining model.Object.ID() method to fetch it.
	//
	// This isn't as performant, so we'll avoid doing this whenever possible.
	obj, err := model.NewObject(typ.Name)
	if err != nil {
		return "", err
	}
	if err := obj.SetResource(msg); err != nil {
		return "", err
	}
	id := obj.ID()
	if id == "" {
		// Should never happen, but just a sanity check.
		return "", fmt.Errorf("unable to determine identifier field for type %q", typ.Name)
	}

	return id, nil
}
