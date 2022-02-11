package admin

import (
	"context"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gavv/httpexpect/v2"
	"github.com/google/uuid"
	model "github.com/kong/koko/internal/gen/grpc/kong/admin/model/v1"
	service "github.com/kong/koko/internal/gen/grpc/kong/admin/service/v1"
	"github.com/kong/koko/internal/log"
	"github.com/kong/koko/internal/resource"
	serverUtil "github.com/kong/koko/internal/server/util"
	"github.com/kong/koko/internal/store"
	"github.com/kong/koko/internal/test/util"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"
)

func goodNode() *model.Node {
	return &model.Node{
		Id:       uuid.NewString(),
		Hostname: "secure-server",
		Version:  "42.1.0",
		Type:     resource.NodeTypeKongProxy,
		LastPing: 42,
	}
}

func TestNodeCreateUpsert(t *testing.T) {
	p, err := util.GetPersister()
	require.Nil(t, err)
	objectStore := store.New(p, log.Logger)

	storeLoader := serverUtil.DefaultStoreLoader{
		Store: objectStore.ForCluster("default"),
	}
	nodeService := &NodeService{
		CommonOpts: CommonOpts{
			logger:      log.Logger,
			storeLoader: storeLoader,
		},
	}

	l := setupBufConn()
	grpcServer := grpc.NewServer()
	service.RegisterNodeServiceServer(grpcServer, nodeService)
	cc := clientConn(t, l)

	client := service.NewNodeServiceClient(cc)
	go func() {
		_ = grpcServer.Serve(l)
	}()
	defer grpcServer.Stop()
	ctx := context.Background()

	t.Run("creates a valid node", func(t *testing.T) {
		resp, err := client.CreateNode(ctx, &service.CreateNodeRequest{
			Item: goodNode(),
		})
		require.Nil(t, err)
		require.NotNil(t, resp)
	})
	t.Run("creating invalid node fails with 400", func(t *testing.T) {
		node := &model.Node{
			Id:       uuid.NewString(),
			Hostname: "secure-server",
			Version:  "",
			Type:     resource.NodeTypeKongProxy,
			LastPing: -42,
		}
		resp, err := client.CreateNode(ctx, &service.CreateNodeRequest{
			Item: node,
		})
		require.NotNil(t, err)
		require.Nil(t, resp)
	})
	t.Run("upserts a valid node", func(t *testing.T) {
		resp, err := client.UpsertNode(ctx, &service.UpsertNodeRequest{
			Item: goodNode(),
		})
		require.Nil(t, err)
		require.NotNil(t, resp)
	})
	t.Run("upserts a valid node", func(t *testing.T) {
		resp, err := client.UpsertNode(ctx, &service.UpsertNodeRequest{
			Item: goodNode(),
		})
		require.Nil(t, err)
		require.NotNil(t, resp)
	})
	t.Run("upsert correctly updates a route", func(t *testing.T) {
		nid := uuid.NewString()
		node := goodNode()
		node.Id = nid

		resp, err := client.UpsertNode(ctx, &service.UpsertNodeRequest{
			Item: goodNode(),
		})
		require.Nil(t, err)
		require.NotNil(t, resp)
		require.NotNil(t, resp.Item)

		node.Version = "new-version"
		node.Hostname = "new-hostname"
		resp, err = client.UpsertNode(ctx, &service.UpsertNodeRequest{
			Item: node,
		})
		require.Nil(t, err)
		require.NotNil(t, resp)
		require.NotNil(t, resp.Item)
		require.Equal(t, "new-hostname", resp.Item.Hostname)
		require.Equal(t, "new-version", resp.Item.Version)
	})
}

func TestNodeDelete(t *testing.T) {
	p, err := util.GetPersister()
	require.Nil(t, err)
	objectStore := store.New(p, log.Logger)

	storeLoader := serverUtil.DefaultStoreLoader{
		Store: objectStore.ForCluster("default"),
	}
	nodeService := &NodeService{
		CommonOpts: CommonOpts{
			logger:      log.Logger,
			storeLoader: storeLoader,
		},
	}

	l := setupBufConn()
	grpcServer := grpc.NewServer()
	service.RegisterNodeServiceServer(grpcServer, nodeService)
	cc := clientConn(t, l)

	client := service.NewNodeServiceClient(cc)
	go func() {
		_ = grpcServer.Serve(l)
	}()
	defer grpcServer.Stop()
	ctx := context.Background()
	node := goodNode()
	nodeID := node.Id
	resp, err := client.CreateNode(ctx, &service.CreateNodeRequest{
		Item: node,
	})
	require.Nil(t, err)
	require.NotNil(t, resp)

	handler, err := NewHandler(HandlerOpts{
		Logger: log.Logger,
		StoreLoader: serverUtil.DefaultStoreLoader{
			Store: objectStore.ForCluster("default"),
		},
	})
	require.Nil(t, err)

	s := httptest.NewServer(handler)
	defer s.Close()
	c := httpexpect.New(t, s.URL)

	t.Run("deleting a non-existent node returns 404", func(t *testing.T) {
		c.DELETE("/v1/nodes/" + uuid.NewString()).Expect().Status(http.StatusNotFound)
	})
	t.Run("deleting a node return 204", func(t *testing.T) {
		c.DELETE("/v1/nodes/" + nodeID).Expect().Status(http.StatusNoContent)
	})
	t.Run("delete request without an ID returns 400", func(t *testing.T) {
		c.DELETE("/v1/nodes/").Expect().Status(http.StatusBadRequest)
	})
}

func TestNodeRead(t *testing.T) {
	p, err := util.GetPersister()
	require.Nil(t, err)
	objectStore := store.New(p, log.Logger)

	storeLoader := serverUtil.DefaultStoreLoader{
		Store: objectStore.ForCluster("default"),
	}
	nodeService := &NodeService{
		CommonOpts: CommonOpts{
			logger:      log.Logger,
			storeLoader: storeLoader,
		},
	}

	l := setupBufConn()
	grpcServer := grpc.NewServer()
	service.RegisterNodeServiceServer(grpcServer, nodeService)
	cc := clientConn(t, l)

	client := service.NewNodeServiceClient(cc)
	go func() {
		_ = grpcServer.Serve(l)
	}()
	defer grpcServer.Stop()
	ctx := context.Background()
	node := goodNode()
	nodeID := node.Id
	resp, err := client.CreateNode(ctx, &service.CreateNodeRequest{
		Item: node,
	})
	require.Nil(t, err)
	require.NotNil(t, resp)

	handler, err := NewHandler(HandlerOpts{
		Logger: log.Logger,
		StoreLoader: serverUtil.DefaultStoreLoader{
			Store: objectStore.ForCluster("default"),
		},
	})
	require.Nil(t, err)

	s := httptest.NewServer(handler)
	defer s.Close()
	c := httpexpect.New(t, s.URL)

	t.Run("reading a non-existent node returns 404", func(t *testing.T) {
		c.GET("/v1/nodes/" + uuid.NewString()).Expect().Status(http.StatusNotFound)
	})
	t.Run("reading a node return 200", func(t *testing.T) {
		res := c.GET("/v1/nodes/" + nodeID).Expect().Status(http.StatusOK)
		body := res.JSON().Path("$.item").Object()
		body.ValueEqual("id", nodeID)
		body.ValueEqual("hostname", "secure-server")
		body.ValueEqual("version", "42.1.0")
		body.ValueEqual("type", resource.NodeTypeKongProxy)
		body.ValueEqual("last_ping", 42)
	})
	t.Run("read request without an ID returns 400", func(t *testing.T) {
		c.GET("/v1/nodes/").Expect().Status(http.StatusBadRequest)
	})
}

func TestNodeList(t *testing.T) {
	p, err := util.GetPersister()
	require.Nil(t, err)
	objectStore := store.New(p, log.Logger)

	storeLoader := serverUtil.DefaultStoreLoader{
		Store: objectStore.ForCluster("default"),
	}
	nodeService := &NodeService{
		CommonOpts: CommonOpts{
			logger:      log.Logger,
			storeLoader: storeLoader,
		},
	}

	l := setupBufConn()
	grpcServer := grpc.NewServer()
	service.RegisterNodeServiceServer(grpcServer, nodeService)
	cc := clientConn(t, l)

	client := service.NewNodeServiceClient(cc)
	go func() {
		_ = grpcServer.Serve(l)
	}()
	defer grpcServer.Stop()
	ctx := context.Background()

	// create node 1
	node1 := goodNode()
	id1 := node1.Id
	resp, err := client.CreateNode(ctx, &service.CreateNodeRequest{
		Item: node1,
	})
	require.Nil(t, err)
	require.NotNil(t, resp)

	// create node 2
	node2 := goodNode()
	id2 := node2.Id
	resp, err = client.CreateNode(ctx, &service.CreateNodeRequest{
		Item: node2,
	})
	require.Nil(t, err)
	require.NotNil(t, resp)

	handler, err := NewHandler(HandlerOpts{
		Logger: log.Logger,
		StoreLoader: serverUtil.DefaultStoreLoader{
			Store: objectStore.ForCluster("default"),
		},
	})
	require.Nil(t, err)

	s := httptest.NewServer(handler)
	defer s.Close()
	c := httpexpect.New(t, s.URL)

	t.Run("list returns multiple nodes", func(t *testing.T) {
		body := c.GET("/v1/nodes").Expect().Status(http.StatusOK).JSON().Object()
		items := body.Value("items").Array()
		items.Length().Equal(2)
		var gotIDs []string
		for _, item := range items.Iter() {
			gotIDs = append(gotIDs, item.Object().Value("id").String().Raw())
		}
		require.ElementsMatch(t, []string{id1, id2}, gotIDs)
	})
	t.Run("list returns multiple nodes with paging", func(t *testing.T) {
		// Get First Page
		body := c.GET("/v1/nodes").
			WithQuery("pagination.size", "1").
			WithQuery("pagination.page", "1").
			Expect().Status(http.StatusOK).JSON().Object()
		items := body.Value("items").Array()
		items.Length().Equal(1)
		id1Got := items.Element(0).Object().Value("id").String().Raw()
		body.Value("pagination").Object().Value("total_count").Number().Equal(2)
		body.Value("pagination").Object().Value("next_page").Number().Equal(2)
		body = c.GET("/v1/nodes").
			WithQuery("pagination.size", "1").
			WithQuery("pagination.page", "2").
			Expect().Status(http.StatusOK).JSON().Object()
		items = body.Value("items").Array()
		items.Length().Equal(1)
		id2Got := items.Element(0).Object().Value("id").String().Raw()
		body.Value("pagination").Object().Value("total_count").Number().Equal(2)
		body.Value("pagination").Object().NotContainsKey("next_page")
		require.ElementsMatch(t, []string{id1, id2}, []string{id1Got, id2Got})
	})
}

func setupBufConn() *bufconn.Listener {
	const bufSize = 1024 * 1024
	return bufconn.Listen(bufSize)
}

func clientConn(t *testing.T, l *bufconn.Listener) grpc.ClientConnInterface {
	conn, err := grpc.DialContext(context.Background(),
		"bufnet",
		grpc.WithContextDialer(func(ctx context.Context, s string) (net.Conn, error) {
			return l.Dial()
		}),
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.Nil(t, err)
	return conn
}
