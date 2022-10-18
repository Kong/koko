package admin

import (
	"context"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gavv/httpexpect/v2"
	"github.com/google/uuid"
	model "github.com/kong/koko/internal/gen/grpc/kong/admin/model/v1"
	service "github.com/kong/koko/internal/gen/grpc/kong/admin/service/v1"
	nonPublic "github.com/kong/koko/internal/gen/grpc/kong/nonpublic/v1"
	"github.com/kong/koko/internal/log"
	"github.com/kong/koko/internal/resource"
	_ "github.com/kong/koko/internal/server/kong/ws/config/compat"
	serverUtil "github.com/kong/koko/internal/server/util"
	"github.com/kong/koko/internal/store"
	"github.com/kong/koko/internal/test/util"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"
)

func goodNode() *model.Node {
	return &model.Node{
		Id:         uuid.NewString(),
		Hostname:   "secure-server",
		Version:    "42.1.0",
		ConfigHash: strings.Repeat("foo0", 8),
		Type:       resource.NodeTypeKongProxy,
		LastPing:   42,
	}
}

func TestNodeCreateUpsert(t *testing.T) {
	p, err := util.GetPersister(t)
	require.Nil(t, err)
	objectStore := store.New(p, log.Logger)

	storeLoader := serverUtil.DefaultStoreLoader{
		Store: objectStore.ForCluster(store.DefaultCluster),
	}
	nodeService := &NodeService{
		CommonOpts: CommonOpts{
			loggerFields: []zapcore.Field{zap.String("admin-service", "node")},
			storeLoader:  storeLoader,
		},
	}

	l := setupBufConn()
	grpcServer := grpc.NewServer(grpc.ChainUnaryInterceptor(
		serverUtil.LoggerInterceptor(log.Logger),
		serverUtil.PanicInterceptor(log.Logger)))
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
	p, err := util.GetPersister(t)
	require.Nil(t, err)
	objectStore := store.New(p, log.Logger)

	db := objectStore.ForCluster(store.DefaultCluster)
	ctx := context.Background()

	handler, err := NewHandler(HandlerOpts{
		Logger: log.Logger,
		StoreLoader: serverUtil.DefaultStoreLoader{
			Store: objectStore.ForCluster(store.DefaultCluster),
		},
		Validator: validator,
	})
	require.Nil(t, err)
	handler = serverUtil.HandlerWithRecovery(serverUtil.HandlerWithLogger(handler, log.Logger), log.Logger)

	s := httptest.NewServer(handler)
	defer s.Close()
	c := httpexpect.New(t, s.URL)

	t.Run("deleting a non-existent node returns 404", func(t *testing.T) {
		c.DELETE("/v1/nodes/" + uuid.NewString()).Expect().Status(http.StatusNotFound)
	})
	t.Run("deleting a node return 204", func(t *testing.T) {
		n := resource.NewNode()
		n.Node = goodNode()
		nodeID := n.ID()
		err := db.Create(ctx, n)
		require.NoError(t, err)
		c.DELETE("/v1/nodes/" + nodeID).Expect().Status(http.StatusNoContent)
	})
	t.Run("delete request without an ID returns 400", func(t *testing.T) {
		c.DELETE("/v1/nodes/").Expect().Status(http.StatusBadRequest)
	})
	t.Run("deleting a node deletes the corresponding node-status", func(t *testing.T) {
		n := resource.NewNode()
		n.Node = goodNode()
		nodeID := n.ID()
		err := db.Create(ctx, n)
		require.NoError(t, err)

		nodeStatus := resource.NewNodeStatus()
		nodeStatus.NodeStatus = &nonPublic.NodeStatus{
			Id: n.ID(),
			Issues: []*model.CompatibilityIssue{
				{
					Code: "F424",
				},
			},
		}
		err = db.Create(ctx, nodeStatus)
		require.NoError(t, err)

		c.DELETE("/v1/nodes/" + nodeID).Expect().Status(http.StatusNoContent)

		nodeStatus = resource.NewNodeStatus()
		err = db.Read(ctx, nodeStatus, store.GetByID(nodeID))
		require.ErrorIs(t, err, store.ErrNotFound)
	})
}

func TestNodeRead(t *testing.T) {
	p, err := util.GetPersister(t)
	require.Nil(t, err)
	objectStore := store.New(p, log.Logger)
	db := objectStore.ForCluster(store.DefaultCluster)

	ctx := context.Background()

	handler, err := NewHandler(HandlerOpts{
		Logger: log.Logger,
		StoreLoader: serverUtil.DefaultStoreLoader{
			Store: objectStore.ForCluster(store.DefaultCluster),
		},
		Validator: validator,
	})
	require.Nil(t, err)
	handler = serverUtil.HandlerWithRecovery(serverUtil.HandlerWithLogger(handler, log.Logger), log.Logger)

	s := httptest.NewServer(handler)
	defer s.Close()
	c := httpexpect.New(t, s.URL)

	t.Run("reading a non-existent node returns 404", func(t *testing.T) {
		c.GET("/v1/nodes/" + uuid.NewString()).Expect().Status(http.StatusNotFound)
	})
	t.Run("reading a node return 200", func(t *testing.T) {
		n := resource.NewNode()
		n.Node = goodNode()
		nodeID := n.ID()
		err = db.Create(ctx, n)
		require.NoError(t, err)

		res := c.GET("/v1/nodes/" + nodeID).Expect().Status(http.StatusOK)
		body := res.JSON().Path("$.item").Object()
		body.ValueEqual("id", nodeID)
		body.ValueEqual("hostname", "secure-server")
		body.ValueEqual("version", "42.1.0")
		body.ValueEqual("type", resource.NodeTypeKongProxy)
		body.ValueEqual("last_ping", 42)
		compatibilityStatus := body.Path("$.compatibility_status").Object()
		compatibilityStatus.ValueEqual("state", "COMPATIBILITY_STATE_UNKNOWN")
	})
	t.Run("read request without an ID returns 400", func(t *testing.T) {
		c.GET("/v1/nodes/").Expect().Status(http.StatusBadRequest)
	})
	t.Run("read a node with incompatible node-status", func(t *testing.T) {
		n := resource.NewNode()
		n.Node = goodNode()
		nodeID := n.ID()
		err = db.Create(ctx, n)
		require.NoError(t, err)

		nodeStatus := resource.NewNodeStatus()
		nodeStatus.NodeStatus = &nonPublic.NodeStatus{
			Id: nodeID,
			Issues: []*model.CompatibilityIssue{
				{
					Code: "P101",
					AffectedResources: []*model.Resource{
						{
							Type: "plugin",
							Id:   "f43976aa-9342-46ef-a36c-8a154b29ed21",
						},
					},
				},
			},
		}
		err = db.Create(ctx, nodeStatus)
		require.NoError(t, err)

		res := c.GET("/v1/nodes/" + nodeID).Expect().Status(http.StatusOK)
		body := res.JSON().Path("$.item").Object()
		body.ValueEqual("id", nodeID)
		body.ValueEqual("hostname", "secure-server")
		body.ValueEqual("version", "42.1.0")
		body.ValueEqual("type", resource.NodeTypeKongProxy)
		body.ValueEqual("last_ping", 42)
		compatibilityStatus := body.Path("$.compatibility_status").Object()
		compatibilityStatus.ValueEqual("state",
			"COMPATIBILITY_STATE_INCOMPATIBLE")
		issues := compatibilityStatus.Value("issues").Array()
		issues.Length().Equal(1)
		issue := issues.First().Object()
		issue.ValueEqual("code", "P101")
		issue.ValueEqual("severity", "error")

		description := strings.ReplaceAll(`For the 'acme' plugin,
 one or more of the following 'config' fields are set: 'preferred_chain',
 'storage_config.vault.auth_method', 'storage_config.vault.auth_path',
 'storage_config.vault.auth_role', 'storage_config.vault.jwt_path' but
 Kong Gateway versions < 2.6 do not support these fields.
 Plugin features that rely on these fields are not working as intended.`, "\n", "")
		gotDescription := issue.Value("description").String().Raw()
		require.Equal(t, description, gotDescription)
		issue.ValueEqual("resolution", `Please upgrade Kong Gateway to version '2.6' or above.`)

		affectedResources := issue.Value("affected_resources").Array()
		affectedResources.Length().Equal(1)
		affectedResources.First().Object().Equal(map[string]interface{}{
			"type": "plugin",
			"id":   "f43976aa-9342-46ef-a36c-8a154b29ed21",
		})
	})
	t.Run("read a node with fully compatible node-status", func(t *testing.T) {
		n := resource.NewNode()
		n.Node = goodNode()
		nodeID := n.ID()
		err = db.Create(ctx, n)
		require.NoError(t, err)

		nodeStatus := resource.NewNodeStatus()
		nodeStatus.NodeStatus = &nonPublic.NodeStatus{
			Id: nodeID,
		}
		err = db.Create(ctx, nodeStatus)
		require.NoError(t, err)

		res := c.GET("/v1/nodes/" + nodeID).Expect().Status(http.StatusOK)
		body := res.JSON().Path("$.item").Object()
		body.ValueEqual("id", nodeID)
		body.ValueEqual("hostname", "secure-server")
		body.ValueEqual("version", "42.1.0")
		body.ValueEqual("type", resource.NodeTypeKongProxy)
		body.ValueEqual("last_ping", 42)
		body.ContainsKey("compatibility_status")
		compatibilityStatus := body.Path("$.compatibility_status").Object()
		compatibilityStatus.ValueEqual("state",
			"COMPATIBILITY_STATE_FULLY_COMPATIBLE")
	})
	t.Run("read a node with unknown node-status due to unregistered change-id", func(t *testing.T) {
		n := resource.NewNode()
		n.Node = goodNode()
		nodeID := n.ID()
		err = db.Create(ctx, n)
		require.NoError(t, err)

		nodeStatus := resource.NewNodeStatus()
		nodeStatus.NodeStatus = &nonPublic.NodeStatus{
			Id: nodeID,
			Issues: []*model.CompatibilityIssue{
				{
					Code: "Z101",
				},
			},
		}
		err = db.Create(ctx, nodeStatus)
		require.NoError(t, err)

		res := c.GET("/v1/nodes/" + nodeID).Expect().Status(http.StatusOK)
		body := res.JSON().Path("$.item").Object()
		body.ValueEqual("id", nodeID)
		body.ValueEqual("hostname", "secure-server")
		body.ValueEqual("version", "42.1.0")
		body.ValueEqual("type", resource.NodeTypeKongProxy)
		body.ValueEqual("last_ping", 42)
		compatibilityStatus := body.Path("$.compatibility_status").Object()
		compatibilityStatus.ValueEqual("state",
			"COMPATIBILITY_STATE_UNKNOWN")
	})
	t.Run("read a node with zero config hash", func(t *testing.T) {
		n := resource.NewNode()
		n.Node = goodNode()
		n.Node.ConfigHash = emptyConfigHash
		nodeID := n.ID()
		err = db.Create(ctx, n)
		require.NoError(t, err)

		nodeStatus := resource.NewNodeStatus()
		nodeStatus.NodeStatus = &nonPublic.NodeStatus{
			Id: nodeID,
		}
		err = db.Create(ctx, nodeStatus)
		require.NoError(t, err)

		res := c.GET("/v1/nodes/" + nodeID).Expect().Status(http.StatusOK)
		body := res.JSON().Path("$.item").Object()
		body.ValueEqual("id", nodeID)
		body.ValueEqual("hostname", "secure-server")
		body.ValueEqual("version", "42.1.0")
		body.ValueEqual("type", resource.NodeTypeKongProxy)
		body.ValueEqual("last_ping", 42)
		body.ContainsKey("compatibility_status")
		compatibilityStatus := body.Path("$.compatibility_status").Object()
		compatibilityStatus.ValueEqual("state",
			"COMPATIBILITY_STATE_UNKNOWN")
	})
	t.Run("read a node with empty config hash", func(t *testing.T) {
		n := resource.NewNode()
		n.Node = goodNode()
		n.Node.ConfigHash = ""
		nodeID := n.ID()
		err = db.Create(ctx, n)
		require.NoError(t, err)

		nodeStatus := resource.NewNodeStatus()
		nodeStatus.NodeStatus = &nonPublic.NodeStatus{
			Id: nodeID,
		}
		err = db.Create(ctx, nodeStatus)
		require.NoError(t, err)

		res := c.GET("/v1/nodes/" + nodeID).Expect().Status(http.StatusOK)
		body := res.JSON().Path("$.item").Object()
		body.ValueEqual("id", nodeID)
		body.ValueEqual("hostname", "secure-server")
		body.ValueEqual("version", "42.1.0")
		body.ValueEqual("type", resource.NodeTypeKongProxy)
		body.ValueEqual("last_ping", 42)
		body.ContainsKey("compatibility_status")
		compatibilityStatus := body.Path("$.compatibility_status").Object()
		compatibilityStatus.ValueEqual("state",
			"COMPATIBILITY_STATE_UNKNOWN")
	})
}

func TestNodeList(t *testing.T) {
	p, err := util.GetPersister(t)
	require.Nil(t, err)
	objectStore := store.New(p, log.Logger)
	db := objectStore.ForCluster(store.DefaultCluster)

	ctx := context.Background()

	// create node 1
	node1 := resource.NewNode()
	node1.Node = goodNode()
	id1 := node1.ID()
	err = db.Create(ctx, node1)
	require.NoError(t, err)

	// create node 2
	node2 := resource.NewNode()
	node2.Node = goodNode()
	id2 := node2.ID()
	err = db.Create(ctx, node2)
	require.NoError(t, err)

	handler, err := NewHandler(HandlerOpts{
		Logger: log.Logger,
		StoreLoader: serverUtil.DefaultStoreLoader{
			Store: objectStore.ForCluster(store.DefaultCluster),
		},
		Validator: validator,
	})
	require.Nil(t, err)
	handler = serverUtil.HandlerWithRecovery(serverUtil.HandlerWithLogger(handler, log.Logger), log.Logger)

	s := httptest.NewServer(handler)
	defer s.Close()
	c := httpexpect.New(t, s.URL)

	t.Run("list returns multiple nodes", func(t *testing.T) {
		body := c.GET("/v1/nodes").Expect().Status(http.StatusOK).JSON().Object()
		items := body.Value("items").Array()
		items.Length().Equal(2)
		require.ElementsMatch(t, []string{id1, id2}, items.Path("$..id").Raw())
	})
	t.Run("list returns multiple nodes with paging", func(t *testing.T) {
		// Get First Page
		body := c.GET("/v1/nodes").
			WithQuery("page.size", "1").
			WithQuery("page.number", "1").
			Expect().Status(http.StatusOK).JSON().Object()
		items := body.Value("items").Array()
		items.Length().Equal(1)
		id1Got := items.Element(0).Object().Value("id").String().Raw()
		body.Value("page").Object().Value("total_count").Number().Equal(2)
		body.Value("page").Object().Value("next_page_num").Number().Equal(2)
		body = c.GET("/v1/nodes").
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
}

func TestNodeListWithStatus(t *testing.T) {
	p, err := util.GetPersister(t)
	require.Nil(t, err)
	objectStore := store.New(p, log.Logger)
	db := objectStore.ForCluster(store.DefaultCluster)

	ctx := context.Background()

	// node1 = fully compatible
	// node2 = incompatible
	// node3 = unknown

	node1 := resource.NewNode()
	node1.Node = goodNode()
	id1 := node1.ID()
	err = db.Create(ctx, node1)
	require.NoError(t, err)

	nodeStatus1 := resource.NewNodeStatus()
	nodeStatus1.NodeStatus = &nonPublic.NodeStatus{
		Id: id1,
	}
	require.NoError(t, db.Create(ctx, nodeStatus1))

	node2 := resource.NewNode()
	node2.Node = goodNode()
	id2 := node2.ID()
	err = db.Create(ctx, node2)
	require.NoError(t, err)

	nodeStatus2 := resource.NewNodeStatus()
	nodeStatus2.NodeStatus = &nonPublic.NodeStatus{
		Id: id2,
		Issues: []*model.CompatibilityIssue{
			{
				Code: "P101",
				AffectedResources: []*model.Resource{
					{
						Type: "plugin",
						Id:   "f43976aa-9342-46ef-a36c-8a154b29ed21",
					},
				},
			},
		},
	}
	require.NoError(t, db.Create(ctx, nodeStatus2))

	node3 := resource.NewNode()
	node3.Node = goodNode()
	id3 := node3.ID()
	err = db.Create(ctx, node3)
	require.NoError(t, err)

	handler, err := NewHandler(HandlerOpts{
		Logger: log.Logger,
		StoreLoader: serverUtil.DefaultStoreLoader{
			Store: objectStore.ForCluster(store.DefaultCluster),
		},
		Validator: validator,
	})
	require.Nil(t, err)
	handler = serverUtil.HandlerWithRecovery(serverUtil.HandlerWithLogger(handler, log.Logger), log.Logger)

	s := httptest.NewServer(handler)
	defer s.Close()
	c := httpexpect.New(t, s.URL)

	body := c.GET("/v1/nodes").Expect().Status(http.StatusOK).JSON().Object()
	items := body.Value("items").Array()
	items.Length().Equal(3)
	var gotIDs []string
	for _, item := range items.Iter() {
		gotIDs = append(gotIDs, item.Object().Value("id").String().Raw())
	}
	require.ElementsMatch(t, []string{id1, id2, id3}, gotIDs)

	for _, node := range items.Iter() {
		nodeID := node.Object().Value("id").String().Raw()
		switch nodeID {
		case id1:
			// node1 = fully compatible
			node.Object().Path("$.compatibility_status.state").String().
				Equal("COMPATIBILITY_STATE_FULLY_COMPATIBLE")
			node.Object().NotContainsKey("$.compatibility_status.issues")

		case id2:
			// node2 = incompatible
			node.Object().Path("$.compatibility_status.state").String().
				Equal("COMPATIBILITY_STATE_INCOMPATIBLE")
			compatibilityStatus := node.Path("$.compatibility_status").Object()
			compatibilityStatus.ValueEqual("state",
				"COMPATIBILITY_STATE_INCOMPATIBLE")
			issues := compatibilityStatus.Value("issues").Array()
			issues.Length().Equal(1)
			issue := issues.First().Object()
			issue.ValueEqual("code", "P101")
			issue.ValueEqual("severity", "error")

			description := strings.ReplaceAll(`For the 'acme' plugin,
 one or more of the following 'config' fields are set: 'preferred_chain',
 'storage_config.vault.auth_method', 'storage_config.vault.auth_path',
 'storage_config.vault.auth_role', 'storage_config.vault.jwt_path' but
 Kong Gateway versions < 2.6 do not support these fields.
 Plugin features that rely on these fields are not working as intended.`, "\n", "")
			gotDescription := issue.Value("description").String().Raw()
			require.Equal(t, description, gotDescription)
			issue.ValueEqual("resolution", `Please upgrade Kong Gateway to version '2.6' or above.`)

			affectedResources := issue.Value("affected_resources").Array()
			affectedResources.Length().Equal(1)
			affectedResources.First().Object().Equal(map[string]interface{}{
				"type": "plugin",
				"id":   "f43976aa-9342-46ef-a36c-8a154b29ed21",
			})

		case id3:
			// node3 = unknown
			node.Object().Path("$.compatibility_status.state").String().
				Equal("COMPATIBILITY_STATE_UNKNOWN")
			node.Object().NotContainsKey("$.compatibility_status.issues")

		default:
			t.Fail()
		}
	}
}

func TestNodeService_listAllNodeStatus(t *testing.T) {
	p, err := util.GetPersister(t)
	require.Nil(t, err)
	objectStore := store.New(p, log.Logger)

	storeLoader := serverUtil.DefaultStoreLoader{
		Store: objectStore.ForCluster(store.DefaultCluster),
	}
	db := objectStore.ForCluster(store.DefaultCluster)
	nodeService := &NodeService{
		CommonOpts: CommonOpts{
			loggerFields: []zapcore.Field{zap.String("admin-service", "node")},
			storeLoader:  storeLoader,
		},
	}
	ctx := context.Background()

	t.Run("lists no node statuses", func(t *testing.T) {
		nodeStatuses, err := nodeService.listAllNodeStatus(ctx, db)
		require.NoError(t, err)
		require.Empty(t, nodeStatuses)
	})
	t.Run("lists 1001 node-statuses with pagination", func(t *testing.T) {
		for i := 0; i <= 1000; i++ {
			nodeStatus := resource.NewNodeStatus()
			nodeStatus.NodeStatus = &nonPublic.NodeStatus{
				Id: uuid.NewString(),
				Issues: []*model.CompatibilityIssue{
					{
						Code: "F424",
					},
				},
			}
			err = db.Create(ctx, nodeStatus)
			require.NoError(t, err)
		}

		nodeStatuses, err := nodeService.listAllNodeStatus(ctx, db)
		require.NoError(t, err)
		require.Len(t, nodeStatuses, 1001)
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
