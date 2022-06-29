package admin

import (
	"context"
	"net/http"
	"strings"
	"testing"

	"github.com/gavv/httpexpect/v2"
	v1 "github.com/kong/koko/internal/gen/grpc/kong/admin/model/v1"
	"github.com/kong/koko/internal/model/json/validation"
	"github.com/kong/koko/internal/test/seed"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

	// Seed all test resources with random tags, so that we can test filtering resources by tag with a CEL expression.
	seeder, err := seed.New(seed.Config{URL: server.URL})
	require.NoError(t, err)
	_, err = seeder.SeedAllTypes(
		context.Background(),

		// Create 25 resources for each resource type, which is a sane amount for these tests. However,
		// this can be anything, as the tests will automatically adapt to the number of resources created.
		seed.WithResourceCount(25),

		// The test cases are hard-coded to expect no more than two tags on a resource.
		seed.WithRandomTagCount(2, true),

		// We don't want to seed resources that aren't exposed over the REST
		// API. For resources that do not have the tags field, we'll skip
		// testing that resource & log it on the test output as skipped.
		seed.WithIgnoredErrors(seed.ErrRequiredFieldMissing, seed.ErrNoResourcePath),
	)
	require.NoError(t, err)

	results := seeder.Results()
	require.Greater(t, len(results.All()), 0)

	// Run tests for every resource that has been registered.
	for _, resourceInfo := range seeder.AllResourceInfo() {
		expectedResources := results.ByType(resourceInfo.Name)

		// Store resource IDs that contain the keyed tag(s). Resources that have
		// multiple tags will be duplicated for each tag.
		//
		// The "all" key contains all resources with a tag. Additionally, each
		// combination of tags will have its own key, e.g.: `tag1, tag2` is a
		// valid key defining resources that contain both `tag1` and `tag2.
		resourceIDsByTag := map[string][]string{"all": make([]string, 0)}
		for _, r := range expectedResources.All() {
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
				pageRequest: &v1.PaginationRequest{Filter: `"tag1" in something`},
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
				pageRequest:      &v1.PaginationRequest{Filter: `"tag2" in tags`},
				expectedPagedIDs: [][]string{resourceIDsByTag["tag2"]},
			},
			{
				name:             "single tag: multiple results",
				pageRequest:      &v1.PaginationRequest{Filter: `"tag1" in tags`},
				expectedPagedIDs: [][]string{resourceIDsByTag["tag1"]},
			},
			{
				name:        "single tag: no results",
				pageRequest: &v1.PaginationRequest{Filter: `"tag3" in tags`},
			},
			{
				name:             "logical or",
				pageRequest:      &v1.PaginationRequest{Filter: `"tag1" in tags || "tag2" in tags`},
				expectedPagedIDs: [][]string{resourceIDsByTag["all"]},
			},
			{
				name:             "logical and",
				pageRequest:      &v1.PaginationRequest{Filter: `"tag1" in tags && "tag2" in tags`},
				expectedPagedIDs: [][]string{resourceIDsByTag["tag1, tag2"]},
			},
			{
				name:             "exists() macro: with results",
				pageRequest:      &v1.PaginationRequest{Filter: `["tag1", "tag2"].exists(x, x in tags)`},
				expectedPagedIDs: [][]string{resourceIDsByTag["all"]},
			},
			{
				name:             "all() macro: with results",
				pageRequest:      &v1.PaginationRequest{Filter: `["tag1", "tag2"].all(x, x in tags)`},
				expectedPagedIDs: [][]string{resourceIDsByTag["tag1, tag2"]},
			},
			{
				name:        "exists() macro: no results",
				pageRequest: &v1.PaginationRequest{Filter: `["tag3", "tag4"].exists(x, x in tags)`},
			},
			{
				name:        "all() macro: no results",
				pageRequest: &v1.PaginationRequest{Filter: `["tag1", "tag3"].all(x, x in tags)`},
			},
			{
				name:             "empty slice with macro",
				pageRequest:      &v1.PaginationRequest{Filter: `[].all(x, x in tags)`},
				expectedPagedIDs: [][]string{expectedResources.IDs()},
			},
			{
				name:             "duplicate tags",
				pageRequest:      &v1.PaginationRequest{Filter: `["tag1", "tag1"].exists(x, x in tags)`},
				expectedPagedIDs: [][]string{resourceIDsByTag["tag1"]},
			},
			{
				name:             "pagination",
				pageRequest:      &v1.PaginationRequest{Size: 1, Filter: `"tag1" in tags`},
				expectedPagedIDs: lo.Chunk(resourceIDsByTag["tag1"], 1),
			},
		}

		t.Run(string(resourceInfo.Name), func(t *testing.T) {
			// Skip tests that we cannot run list filters on.
			if resourceInfo.JSONSchemaConfig.ResourceAPIPath == "" {
				t.Skipf(
					"Tag-based listing tests for type %s skipped, as it does"+
						" not set the API resource path per the JSON schema.",
					resourceInfo.Name,
				)
				return
			}
			if !resourceInfo.HasField("tags") {
				t.Skipf(
					"Tag-based listing tests for type %s skipped, as it does"+
						" not have a tags field per the JSON schema.",
					resourceInfo.Name,
				)
				return
			}

			for i, tt := range tests {
				t.Run(tt.name, func(t *testing.T) {
					testListFilteringForType(t, httpexpect.New(t, server.URL), &tests[i], resourceInfo)
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
