package admin

import (
	"context"
	"net/http"
	"testing"

	"github.com/gavv/httpexpect/v2"
	"github.com/google/uuid"
	v1 "github.com/kong/koko/internal/gen/grpc/kong/admin/model/v1"
	"github.com/kong/koko/internal/log"
	"github.com/kong/koko/internal/resource"
	"github.com/kong/koko/internal/store"
	"github.com/kong/koko/internal/test/util"
	"github.com/stretchr/testify/require"
)

func TestStatusRead(t *testing.T) {
	p, err := util.GetPersister(t)
	require.Nil(t, err)
	objectStore := store.New(p, log.Logger)
	store := objectStore.ForCluster("default")
	status := resource.NewStatus()
	statusID := uuid.NewString()
	contextID := uuid.NewString()
	status.Status = &v1.Status{
		Id: statusID,
		ContextReference: &v1.EntityReference{
			Type: string(resource.TypeService),
			Id:   contextID,
		},
		Conditions: []*v1.Condition{
			{
				Code:     "R0023",
				Message:  "foo bar",
				Severity: resource.SeverityError,
			},
		},
	}
	err = store.Create(context.Background(), status)
	require.Nil(t, err)

	s, cleanup := setupWithDB(t, store)
	defer cleanup()
	c := httpexpect.New(t, s.URL)

	t.Run("reading a non-existent status returns 404", func(t *testing.T) {
		randomID := "071f5040-3e4a-46df-9d98-451e79e318fd"
		c.GET("/v1/statuses/" + randomID).Expect().Status(http.StatusNotFound)
	})
	t.Run("reading a status return 200", func(t *testing.T) {
		res := c.GET("/v1/statuses/" + statusID).Expect().Status(http.StatusOK)
		body := res.JSON().Path("$.item").Object()
		body.ValueEqual("id", statusID)
		body.Path("$.context_reference.id").String().Equal(contextID)
		body.Path("$.context_reference.type").String().Equal("service")
		body.Path("$.conditions[0].code").String().Equal("R0023")
		body.Path("$.conditions[0].message").String().Equal("foo bar")
		body.Path("$.conditions[0].severity").String().Equal("error")
	})
	t.Run("read request without an ID returns 400", func(t *testing.T) {
		c.GET("/v1/statuses/").Expect().Status(http.StatusBadRequest)
	})
}

func TestStatusDelete(t *testing.T) {
	p, err := util.GetPersister(t)
	require.Nil(t, err)
	objectStore := store.New(p, log.Logger)
	store := objectStore.ForCluster("default")
	status := resource.NewStatus()
	statusID := uuid.NewString()
	contextID := uuid.NewString()
	status.Status = &v1.Status{
		Id: statusID,
		ContextReference: &v1.EntityReference{
			Type: string(resource.TypeService),
			Id:   contextID,
		},
		Conditions: []*v1.Condition{
			{
				Code:     "R0023",
				Message:  "foo bar",
				Severity: resource.SeverityError,
			},
		},
	}
	err = store.Create(context.Background(), status)
	require.Nil(t, err)

	s, cleanup := setupWithDB(t, store)
	defer cleanup()
	c := httpexpect.New(t, s.URL)

	t.Run("deleting a non-existent status returns 404", func(t *testing.T) {
		randomID := "071f5040-3e4a-46df-9d98-451e79e318fd"
		c.DELETE("/v1/statuses/" + randomID).Expect().Status(http.StatusNotFound)
	})
	t.Run("deleting a status return 204", func(t *testing.T) {
		c.DELETE("/v1/statuses/" + statusID).Expect().Status(http.StatusNoContent)
	})
	t.Run("delete request without an ID returns 400", func(t *testing.T) {
		c.DELETE("/v1/statuses/").Expect().Status(http.StatusBadRequest)
	})
}

func TestStatusList(t *testing.T) {
	p, err := util.GetPersister(t)
	require.Nil(t, err)
	objectStore := store.New(p, log.Logger)
	store := objectStore.ForCluster("default")
	status := resource.NewStatus()
	id1 := uuid.NewString()
	contextID := uuid.NewString()
	status.Status = &v1.Status{
		Id: id1,
		ContextReference: &v1.EntityReference{
			Type: string(resource.TypeService),
			Id:   contextID,
		},
		Conditions: []*v1.Condition{
			{
				Code:     "R0023",
				Message:  "foo bar",
				Severity: resource.SeverityError,
			},
		},
	}
	err = store.Create(context.Background(), status)
	require.Nil(t, err)
	id2 := uuid.NewString()
	status.Status = &v1.Status{
		Id: id2,
		ContextReference: &v1.EntityReference{
			Type: string(resource.TypeRoute),
			Id:   contextID,
		},
		Conditions: []*v1.Condition{
			{
				Code:     "R0023",
				Message:  "foo bar",
				Severity: resource.SeverityError,
			},
		},
	}
	err = store.Create(context.Background(), status)
	require.Nil(t, err)

	s, cleanup := setupWithDB(t, store)
	defer cleanup()
	c := httpexpect.New(t, s.URL)

	t.Run("list returns multiple statuses", func(t *testing.T) {
		body := c.GET("/v1/statuses").Expect().Status(http.StatusOK).JSON().Object()
		items := body.Value("items").Array()
		items.Length().Equal(2)
		var gotIDs []string
		for _, item := range items.Iter() {
			gotIDs = append(gotIDs, item.Object().Value("id").String().Raw())
		}
		require.ElementsMatch(t, []string{id1, id2}, gotIDs)
	})
	t.Run("list returns multiple statuses with paging", func(t *testing.T) {
		// Get First Page
		body := c.GET("/v1/statuses").
			WithQuery("page.size", "1").
			WithQuery("page.number", "1").
			Expect().Status(http.StatusOK).JSON().Object()
		items := body.Value("items").Array()
		items.Length().Equal(1)
		id1Got := items.Element(0).Object().Value("id").String().Raw()
		body.Value("page").Object().Value("total_count").Number().Equal(2)
		body.Value("page").Object().Value("next_page_num").Number().Equal(2)
		// Get second page
		body = c.GET("/v1/statuses").
			WithQuery("page.size", "1").
			WithQuery("page.number", "2").
			Expect().Status(http.StatusOK).JSON().Object()
		items = body.Value("items").Array()
		items.Length().Equal(1)
		id2Got := items.Element(0).Object().Value("id").String().Raw()
		body.Value("page").Object().Value("total_count").Number().Equal(2)
		body.Value("page").Object().NotContainsKey("next_page_num")
		require.ElementsMatch(t, []string{id1, id2}, []string{id1Got, id2Got})
	})
	t.Run("list returns status by type id", func(t *testing.T) {
		body := c.GET("/v1/statuses").
			WithQuery("ref_type", "route").
			WithQuery("ref_id", contextID).
			Expect().
			Status(http.StatusOK).
			JSON().Object()
		items := body.Value("items").Array()
		items.Length().Equal(1)
	})
	t.Run("list returns 200 with no results on no match", func(t *testing.T) {
		body := c.GET("/v1/statuses").
			WithQuery("ref_type", "route").
			WithQuery("ref_id", uuid.NewString()).
			Expect().
			Status(http.StatusOK).
			JSON().Object()
		body.Empty()
	})
}
