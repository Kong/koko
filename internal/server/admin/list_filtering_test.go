package admin

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/gavv/httpexpect/v2"
	"github.com/google/uuid"
	v1 "github.com/kong/koko/internal/gen/grpc/kong/admin/model/v1"
	"github.com/kong/koko/internal/model"
	"github.com/kong/koko/internal/model/json/validation"
	"github.com/kong/koko/internal/resource"
	"github.com/kong/koko/internal/test/seed"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
)

type listFilterTestData struct {
	name             string
	pageRequest      *v1.PaginationRequest
	expectedPagedIDs [][]string
	expectedErr      *validation.Error
}

func TestListFiltering(t *testing.T) {
	server, cleanup := setup(t)
	defer cleanup()

	seeder, err := seed.New(seed.NewSeederOpts{URL: server.URL})
	require.NoError(t, err)

	seedOpts, err := seed.NewOptionsBuilder().
		// Create 25 resources for each resource type, which is a sane amount for these tests. However,
		// this can be anything, as the tests will automatically adapt to the number of resources created.
		WithResourceCount(25).
		// The test cases are hard-coded to expect no more than two tags on a resource.
		WithRandomTagCount(2, true).
		// For resources that do not have the tags field, we'll skip testing
		// that resource & log it on the test output as skipped.
		WithIgnoredErrors(seed.ErrRequiredFieldMissing).
		Build()
	require.NoError(t, err)

	// Seed all test resources with random tags, so that we can test filtering resources by tag with a CEL expression.
	_, err = seeder.SeedAllTypes(context.Background(), seedOpts)
	require.NoError(t, err)

	results := seeder.Results()
	require.Greater(t, len(results.All()), 0)

	// Run tests for every resource that has been registered.
	for _, resourceInfo := range seeder.AllResourceInfo() {
		expectedResources := results.ByType(resourceInfo.Name)
		resourceIDsByTag := getResourcesByTag(expectedResources)

		// The purpose of these tests are not to validate all possible expression combinations, as that is tested
		// in `internal/server/admin/cel_test.go`. The sole purpose of these tests are to ensure the filtering is
		// working as expected, reporting errors properly, and that pagination is working as intended.
		tests := []listFilterTestData{
			{
				name:             "no filter",
				expectedPagedIDs: [][]string{expectedResources.IDs()},
			},
			{
				name:        "invalid expression",
				pageRequest: &v1.PaginationRequest{Filter: `"tag-1" in something`},
				expectedErr: &validation.Error{
					Errs: []*v1.ErrorDetail{{
						Type:     v1.ErrorType_ERROR_TYPE_FIELD,
						Field:    "page.filter",
						Messages: []string{"invalid filter expression: undeclared reference to 'something'"},
					}},
				},
			},
			{
				name:             "single tag: single result",
				pageRequest:      &v1.PaginationRequest{Filter: `"tag-2" in tags`},
				expectedPagedIDs: [][]string{resourceIDsByTag["tag-2"]},
			},
			{
				name:             "single tag: multiple results",
				pageRequest:      &v1.PaginationRequest{Filter: `"tag-1" in tags`},
				expectedPagedIDs: [][]string{resourceIDsByTag["tag-1"]},
			},
			{
				name:        "single tag: no results",
				pageRequest: &v1.PaginationRequest{Filter: `"tag-3" in tags`},
			},
			{
				name:             "logical or",
				pageRequest:      &v1.PaginationRequest{Filter: `"tag-1" in tags || "tag-2" in tags`},
				expectedPagedIDs: [][]string{resourceIDsByTag["all"]},
			},
			{
				name:             "logical and",
				pageRequest:      &v1.PaginationRequest{Filter: `"tag-1" in tags && "tag-2" in tags`},
				expectedPagedIDs: [][]string{resourceIDsByTag["tag-1, tag-2"]},
			},
			{
				name:             "exists() macro: with results",
				pageRequest:      &v1.PaginationRequest{Filter: `["tag-1", "tag-2"].exists(x, x in tags)`},
				expectedPagedIDs: [][]string{resourceIDsByTag["all"]},
			},
			{
				name:             "all() macro: with results",
				pageRequest:      &v1.PaginationRequest{Filter: `["tag-1", "tag-2"].all(x, x in tags)`},
				expectedPagedIDs: [][]string{resourceIDsByTag["tag-1, tag-2"]},
			},
			{
				name:        "exists() macro: no results",
				pageRequest: &v1.PaginationRequest{Filter: `["tag-3", "tag-4"].exists(x, x in tags)`},
			},
			{
				name:        "all() macro: no results",
				pageRequest: &v1.PaginationRequest{Filter: `["tag-1", "tag-3"].all(x, x in tags)`},
			},
			{
				name:             "empty slice with macro",
				pageRequest:      &v1.PaginationRequest{Filter: `[].all(x, x in tags)`},
				expectedPagedIDs: [][]string{expectedResources.IDs()},
			},
			{
				name:             "duplicate tags",
				pageRequest:      &v1.PaginationRequest{Filter: `["tag-1", "tag-1"].exists(x, x in tags)`},
				expectedPagedIDs: [][]string{resourceIDsByTag["tag-1"]},
			},
			{
				name:             "pagination",
				pageRequest:      &v1.PaginationRequest{Size: 1, Filter: `"tag-1" in tags`},
				expectedPagedIDs: lo.Chunk(resourceIDsByTag["tag-1"], 1),
			},
		}

		t.Run(string(resourceInfo.Name), func(t *testing.T) {
			skipListFilteringTestForResource(t, resourceInfo)

			for i, tt := range tests {
				t.Run(tt.name, func(t *testing.T) {
					testListFilteringForType(t, httpexpect.New(t, server.URL), &tests[i], resourceInfo)
				})
			}
		})
	}
}

// TestListFilteringWithSpaces is a simple test to ensure we can properly query tags with spaces across all resources.
func TestListFilteringWithSpaces(t *testing.T) {
	server, cleanup := setup(t)
	defer cleanup()

	seeder, err := seed.New(seed.NewSeederOpts{URL: server.URL})
	require.NoError(t, err)

	// We'll be creating one of every resource with the below tags.
	validTags := []string{"some tag", "another  tag", "yet another tag"}

	for _, tag := range append([]string{
		// Should not appear in our filter tests, however we'll create some resources
		// with these tags to ensure exact match filtering is working as expected.
		"sometag",
		"some  tag",
		"another tag",
	}, validTags...) {
		seedOpts, err := seed.NewOptionsBuilder().
			WithResourceCount(1).
			WithStaticTags([]string{tag}).
			WithIgnoredErrors(seed.ErrRequiredFieldMissing).
			Build()
		require.NoError(t, err)

		_, err = seeder.SeedAllTypes(context.Background(), seedOpts)
		require.NoError(t, err)
	}

	results := seeder.Results()
	require.Greater(t, len(results.All()), 0)

	for _, resourceInfo := range seeder.AllResourceInfo() {
		t.Run(string(resourceInfo.Name), func(t *testing.T) {
			skipListFilteringTestForResource(t, resourceInfo)

			resourceIDsByTag := getResourcesByTag(results.ByType(resourceInfo.Name))
			for _, tag := range validTags {
				tests := []*listFilterTestData{
					// Should return results.
					{
						pageRequest:      &v1.PaginationRequest{Size: 1, Filter: fmt.Sprintf("%q in tags", tag)},
						expectedPagedIDs: lo.Chunk(resourceIDsByTag[tag], 1),
					},

					// Should not return results.
					{pageRequest: &v1.PaginationRequest{Size: 1, Filter: fmt.Sprintf(`" %s" in tags`, tag)}},
					{pageRequest: &v1.PaginationRequest{Size: 1, Filter: fmt.Sprintf(`"%s " in tags`, tag)}},
					{pageRequest: &v1.PaginationRequest{Size: 1, Filter: fmt.Sprintf(`" %s " in tags`, tag)}},
				}
				for _, tt := range tests {
					t.Run(tt.pageRequest.Filter, func(t *testing.T) {
						testListFilteringForType(t, httpexpect.New(t, server.URL), tt, resourceInfo)
					})
				}
			}
		})
	}
}

func TestListFilteringWithReferenceListing(t *testing.T) {
	s, cleanup := setup(t)
	defer cleanup()

	refID := uuid.NewString()

	tests := []struct {
		apiPath   string
		refFields map[string]model.Type
	}{
		{apiPath: "plugins", refFields: map[string]model.Type{
			"consumer_id": resource.TypeConsumer,
			"route_id":    resource.TypeRoute,
			"service_id":  resource.TypeService,
		}},
		{apiPath: "routes", refFields: map[string]model.Type{"service_id": resource.TypeService}},
		{apiPath: "snis", refFields: map[string]model.Type{"certificate_id": resource.TypeCertificate}},
		{apiPath: "targets", refFields: map[string]model.Type{"upstream_id": resource.TypeUpstream}},
	}
	for _, tt := range tests {
		t.Run(tt.apiPath, func(t *testing.T) {
			for queryArg, refField := range tt.refFields {
				t.Run("by "+string(refField), func(t *testing.T) {
					res := httpexpect.New(t, s.URL).
						GET("/v1/"+tt.apiPath).
						WithQuery("page.filter", `"tag-1" in tags`).
						WithQuery(queryArg, refID).
						Expect()

					res.Status(http.StatusBadRequest)
					body := res.JSON().Object()
					body.Value("code").Number().Equal(codes.FailedPrecondition)
					body.Value("message").String().Equal(
						"listing resources scoped to a resource while " +
							"applying a filter are not yet supported",
					)
				})
			}
		})
	}
}

func testListFilteringForType(t *testing.T, c *httpexpect.Expect, tt *listFilterTestData, typ *seed.ResourceInfo) {
	var actualResults [][]string
	pageNumber := 1
	for {
		req := c.GET("/v1/" + typ.JSONSchemaConfig.ResourceAPIPath)
		if tt.pageRequest != nil {
			req.WithQuery("page.number", pageNumber)
			if tt.pageRequest.Size > 0 {
				req.WithQuery("page.size", tt.pageRequest.Size)
			}
			if tt.pageRequest.Filter != "" {
				req.WithQuery("page.filter", tt.pageRequest.Filter)
			}
		}
		res := req.Expect()
		body := res.JSON().Object()

		if tt.expectedErr != nil {
			res.Status(http.StatusBadRequest)
			body.Value("message").String().Equal("validation error")
			body.Value("details").Array().Length().Equal(len(tt.expectedErr.Errs))
			for i, expectedErr := range tt.expectedErr.Errs {
				errDetails := body.Value("details").Array().Element(i).Object()
				errDetails.Value("type").String().Equal(expectedErr.Type.String())
				messages := errDetails.Value("messages").Array()
				messages.Length().Equal(len(expectedErr.Messages))
				for i, expectedMessage := range expectedErr.Messages {
					messages.Element(i).String().Equal(expectedMessage)
				}
			}
			return
		}

		res.Status(http.StatusOK)

		if len(tt.expectedPagedIDs) == 0 {
			body.Empty()
			return
		}

		body.NotEmpty()

		body.Path("$.page.total_count").Number().Equal(lo.Reduce(
			tt.expectedPagedIDs,
			func(count int, resourceIDs []string, _ int) int {
				count += len(resourceIDs)
				return count
			},
			0,
		))

		resourceIDs := body.Path("$.items[*].id").Array().Raw()
		if len(actualResults) < pageNumber {
			actualResults = append(actualResults, make([]string, len(resourceIDs)))
		}
		for i, resourceID := range resourceIDs {
			actualResults[pageNumber-1][i] = resourceID.(string)
		}

		pageObj := body.Path("$.page").Object().Raw()
		nextPageNum, ok := pageObj["next_page_num"]
		if !ok {
			// No more pages.
			break
		}
		pageNumber = int(nextPageNum.(float64))
	}

	// Ensure actual resources IDs match to what is expected, for each page.
	assert.Len(t, actualResults, len(tt.expectedPagedIDs))
	assert.ElementsMatch(t, lo.Flatten(tt.expectedPagedIDs), lo.Flatten(actualResults))
}

// getResourcesByTag stores resource IDs by tag. Resources that have
// multiple tags will be duplicated for each tag.
//
// The "all" key contains all resources with a tag. Additionally, each
// combination of tags will have its own key, e.g.: `tag-1, tag-2` is a
// valid key defining resources that contain both `tag-1` and `tag-2.
func getResourcesByTag(resources *seed.Results) map[string][]string {
	resourceIDsByTag := map[string][]string{"all": make([]string, 0)}
	for _, r := range resources.All() {
		// Keep track of resources by each tag.
		for _, tag := range r.Tags {
			if _, ok := resourceIDsByTag[tag]; !ok {
				resourceIDsByTag[tag] = make([]string, 0)
			}
			resourceIDsByTag[tag] = append(resourceIDsByTag[tag], r.ID)
		}

		// Keep track of all resources that have at least one tag.
		if len(r.Tags) > 0 {
			resourceIDsByTag["all"] = append(resourceIDsByTag["all"], r.ID)
		}

		// Keep track of resources that share the same combination of tags.
		if len(r.Tags) > 1 {
			joinedTags := strings.Join(r.Tags, ", ")
			if _, ok := resourceIDsByTag[joinedTags]; !ok {
				resourceIDsByTag[joinedTags] = make([]string, 0)
			}
			resourceIDsByTag[joinedTags] = append(resourceIDsByTag[joinedTags], r.ID)
		}
	}

	return resourceIDsByTag
}

// skipListFilteringTestForResource handles skipping tests that we cannot run list filters on.
func skipListFilteringTestForResource(t *testing.T, resourceInfo *seed.ResourceInfo) {
	if resourceInfo.JSONSchemaConfig.ResourceAPIPath == "" {
		t.Skipf(
			"Tag-based listing tests for type %s skipped, as it does"+
				" not set the API resource path per the JSON schema.",
			resourceInfo.Name,
		)
	}
	if !resourceInfo.HasField("tags") {
		t.Skipf(
			"Tag-based listing tests for type %s skipped, as it does"+
				" not have a tags field per the JSON schema.",
			resourceInfo.Name,
		)
	}
}
