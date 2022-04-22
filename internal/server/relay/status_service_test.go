package relay

import (
	"context"
	"testing"

	"github.com/google/uuid"
	model "github.com/kong/koko/internal/gen/grpc/kong/admin/model/v1"
	relay "github.com/kong/koko/internal/gen/grpc/kong/relay/service/v1"
	"github.com/kong/koko/internal/log"
	"github.com/kong/koko/internal/resource"
	serverUtil "github.com/kong/koko/internal/server/util"
	"github.com/kong/koko/internal/status"
	"github.com/kong/koko/internal/store"
	"github.com/kong/koko/internal/test/util"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
)

func TestRelayStatusServiceUpdate(t *testing.T) {
	ctx := context.Background()
	persister, err := util.GetPersister(t)
	require.Nil(t, err)
	db := store.New(persister, log.Logger).ForCluster("default")
	opts := StatusServiceOpts{
		StoreLoader: serverUtil.DefaultStoreLoader{Store: db},
		Logger:      log.Logger,
	}
	server := NewStatusService(opts)
	require.NotNil(t, server)
	l := setup()
	s := grpc.NewServer(grpc.ChainUnaryInterceptor(serverUtil.LoggerInterceptor(opts.Logger)))
	relay.RegisterStatusServiceServer(s, server)
	cc := clientConn(t, l)
	client := relay.NewStatusServiceClient(cc)
	go func() {
		_ = s.Serve(l)
	}()
	defer s.Stop()

	t.Run("updates a given status", func(t *testing.T) {
		defer func() {
			util.CleanDB(t)
		}()
		ctx, cancel := context.WithCancel(ctx)
		defer cancel()
		_, err := client.UpdateStatus(ctx, &relay.UpdateStatusRequest{
			Item: &model.Status{
				ContextReference: &model.EntityReference{
					Type: "node",
					Id:   uuid.NewString(),
				},
				Conditions: []*model.Condition{
					{
						Code:     status.DPMissingPlugin,
						Message:  "foo bar",
						Severity: resource.SeverityError,
					},
				},
			},
		})
		require.Nil(t, err)
		list := resource.NewList(resource.TypeStatus)
		err = db.List(ctx, list)
		require.NoError(t, err)
		require.Len(t, list.GetAll(), 1)
		item := list.GetAll()[0]
		status, ok := item.Resource().(*model.Status)
		require.True(t, ok)
		require.Equal(t, status.ContextReference.Type, "node")
		require.Equal(t, status.Conditions[0].Message, "foo bar")
	})
	t.Run("update current upserts status for an existing reference", func(t *testing.T) {
		defer func() {
			util.CleanDB(t)
		}()
		ctx, cancel := context.WithCancel(ctx)
		defer cancel()
		refID := uuid.NewString()
		_, err := client.UpdateStatus(ctx, &relay.UpdateStatusRequest{
			Item: &model.Status{
				ContextReference: &model.EntityReference{
					Type: "node",
					Id:   refID,
				},
				Conditions: []*model.Condition{
					{
						Code:     status.DPMissingPlugin,
						Message:  "foo",
						Severity: resource.SeverityError,
					},
				},
			},
		})
		require.Nil(t, err)
		list := resource.NewList(resource.TypeStatus)
		err = db.List(ctx, list)
		require.NoError(t, err)
		require.Len(t, list.GetAll(), 1)
		item := list.GetAll()[0]
		currentStatus, ok := item.Resource().(*model.Status)
		require.True(t, ok)
		require.Equal(t, currentStatus.ContextReference.Type, "node")
		require.Equal(t, currentStatus.ContextReference.Id, refID)
		require.Equal(t, currentStatus.Conditions[0].Message, "foo")
		currentStatusID := currentStatus.Id

		_, err = client.UpdateStatus(ctx, &relay.UpdateStatusRequest{
			Item: &model.Status{
				ContextReference: &model.EntityReference{
					Type: "node",
					Id:   refID,
				},
				Conditions: []*model.Condition{
					{
						Code:     status.DPMissingPlugin,
						Message:  "foobar",
						Severity: resource.SeverityError,
					},
				},
			},
		})
		require.Nil(t, err)
		list = resource.NewList(resource.TypeStatus)
		err = db.List(ctx, list)
		require.NoError(t, err)
		require.Len(t, list.GetAll(), 1)
		item = list.GetAll()[0]
		currentStatus, ok = item.Resource().(*model.Status)
		require.True(t, ok)
		require.Equalf(t, currentStatusID, currentStatus.Id,
			"status id to not change across updated")
		require.Equal(t, currentStatus.ContextReference.Type, "node")
		require.Equal(t, currentStatus.ContextReference.Id, refID)
		require.Equal(t, currentStatus.Conditions[0].Message, "foobar")
	})
	t.Run("update without a condition errors", func(t *testing.T) {
		ctx, cancel := context.WithCancel(ctx)
		defer cancel()
		_, err := client.UpdateStatus(ctx, &relay.UpdateStatusRequest{
			Item: &model.Status{
				ContextReference: &model.EntityReference{
					Type: "node",
					Id:   uuid.NewString(),
				},
			},
		})
		require.Error(t, err)
	})
	t.Run("update without a reference errors", func(t *testing.T) {
		ctx, cancel := context.WithCancel(ctx)
		defer cancel()
		_, err := client.UpdateStatus(ctx, &relay.UpdateStatusRequest{
			Item: &model.Status{
				Conditions: []*model.Condition{
					{
						Code:     status.DPMissingPlugin,
						Message:  "node",
						Severity: resource.SeverityError,
					},
				},
			},
		})
		require.Error(t, err)
	})
	t.Run("updates fails with an invalid ref type", func(t *testing.T) {
		ctx, cancel := context.WithCancel(ctx)
		defer cancel()
		_, err := client.UpdateStatus(ctx, &relay.UpdateStatusRequest{
			Item: &model.Status{
				ContextReference: &model.EntityReference{
					Type: "foo",
					Id:   uuid.NewString(),
				},
				Conditions: []*model.Condition{
					{
						Code:     status.DPMissingPlugin,
						Message:  "foo bar",
						Severity: resource.SeverityError,
					},
				},
			},
		})
		require.Error(t, err)
	})
	t.Run("updates fails with an invalid ref id", func(t *testing.T) {
		ctx, cancel := context.WithCancel(ctx)
		defer cancel()
		_, err := client.UpdateStatus(ctx, &relay.UpdateStatusRequest{
			Item: &model.Status{
				ContextReference: &model.EntityReference{
					Type: "node",
					Id:   "borked-on-purpose",
				},
				Conditions: []*model.Condition{
					{
						Code:     status.DPMissingPlugin,
						Message:  "foo bar",
						Severity: resource.SeverityError,
					},
				},
			},
		})
		require.Error(t, err)
	})
}

func TestRelayStatusServiceClear(t *testing.T) {
	ctx := context.Background()
	persister, err := util.GetPersister(t)
	require.Nil(t, err)
	db := store.New(persister, log.Logger).ForCluster("default")
	opts := StatusServiceOpts{
		StoreLoader: serverUtil.DefaultStoreLoader{Store: db},
		Logger:      log.Logger,
	}
	server := NewStatusService(opts)
	require.NotNil(t, server)
	l := setup()
	s := grpc.NewServer(grpc.ChainUnaryInterceptor(serverUtil.LoggerInterceptor(opts.Logger)))
	relay.RegisterStatusServiceServer(s, server)
	cc := clientConn(t, l)
	client := relay.NewStatusServiceClient(cc)
	go func() {
		_ = s.Serve(l)
	}()
	defer s.Stop()

	t.Run("clears a status", func(t *testing.T) {
		ctx, cancel := context.WithCancel(ctx)
		defer cancel()

		refID := uuid.NewString()
		_, err = client.UpdateStatus(ctx, &relay.UpdateStatusRequest{
			Item: &model.Status{
				ContextReference: &model.EntityReference{
					Type: "node",
					Id:   refID,
				},
				Conditions: []*model.Condition{
					{
						Code:     status.DPMissingPlugin,
						Message:  "foo bar",
						Severity: resource.SeverityError,
					},
				},
			},
		})

		_, err = client.ClearStatus(ctx, &relay.ClearStatusRequest{
			ContextReference: &model.EntityReference{
				Type: "node",
				Id:   refID,
			},
		})
		require.NoError(t, err)

		list := resource.NewList(resource.TypeStatus)
		err = db.List(ctx, list)
		require.NoError(t, err)
		require.Len(t, list.GetAll(), 0)
	})
	t.Run("clear status throws an error with invalid type", func(t *testing.T) {
		ctx, cancel := context.WithCancel(ctx)
		defer cancel()
		_, err := client.ClearStatus(ctx, &relay.ClearStatusRequest{
			ContextReference: &model.EntityReference{
				Type: "borked",
				Id:   uuid.NewString(),
			},
		})
		require.Error(t, err)
	})
	t.Run("clear status throws an error with invalid id", func(t *testing.T) {
		ctx, cancel := context.WithCancel(ctx)
		defer cancel()
		_, err := client.ClearStatus(ctx, &relay.ClearStatusRequest{
			ContextReference: &model.EntityReference{
				Type: "node",
				Id:   "borked",
			},
		})
		require.Error(t, err)
	})
}
