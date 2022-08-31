//go:build integration

package e2e

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/gavv/httpexpect/v2"
	kongClient "github.com/kong/go-kong/kong"
	v1 "github.com/kong/koko/internal/gen/grpc/kong/admin/model/v1"
	admin "github.com/kong/koko/internal/gen/grpc/kong/admin/service/v1"
	"github.com/kong/koko/internal/json"
	"github.com/kong/koko/internal/test/kong"
	"github.com/kong/koko/internal/test/run"
	"github.com/kong/koko/internal/test/util"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/structpb"
)

// TestCompatibilityIssueAPI ensures that a tracked change is surfaced to the
// API.
func TestCompatibilityIssueAPI(t *testing.T) {
	cleanup := run.Koko(t)
	defer cleanup()

	dpCleanup := run.KongDP(kong.GetKongConfForShared())
	defer dpCleanup()
	util.WaitForKong(t)
	util.WaitForKongAdminAPI(t)
	// comparing against 2.9 because 3.0.0-alpha.1 is < 3.0.0 as per the current
	// semver implementation
	kongClient.RunWhenKong(t, "< 2.9.0")

	config, err := structpb.NewStruct(map[string]interface{}{
		"endpoint": "http://exmaple.com",
	})
	require.NoError(t, err)
	plugin := &v1.Plugin{
		Name:   "opentelemetry",
		Config: config,
	}
	pluginBytes, err := json.ProtoJSONMarshal(plugin)
	require.NoError(t, err)

	adminClient := httpexpect.New(t, "http://localhost:3000")
	res := adminClient.POST("/v1/plugins").WithBytes(pluginBytes).Expect()
	res.Status(http.StatusCreated)

	plugin = &v1.Plugin{
		Name: "key-auth",
	}
	pluginBytes, err = json.ProtoJSONMarshal(plugin)
	require.NoError(t, err)

	res = adminClient.POST("/v1/plugins").WithBytes(pluginBytes).Expect()
	res.Status(http.StatusCreated)

	expectedConfig := &v1.TestingConfig{
		Plugins: []*v1.Plugin{
			plugin,
		},
	}
	// Validate the configurations
	util.WaitFunc(t, func() error {
		err := util.EnsureConfig(expectedConfig)
		if err != nil {
			t.Log("config validation failed", err)
		}
		return err
	})

	util.WaitFunc(t, func() error {
		cc, err := grpc.Dial("localhost:3001",
			grpc.WithTransportCredentials(insecure.NewCredentials()),
		)
		require.Nil(t, err)
		nodeClient := admin.NewNodeServiceClient(cc)
		resp, err := nodeClient.ListNodes(context.Background(), &admin.ListNodesRequest{})
		require.Nil(t, err)
		if len(resp.Items) != 1 {
			return fmt.Errorf("expected one node")
		}
		node := resp.Items[0]
		if node.CompatibilityStatus.State != v1.CompatibilityState_COMPATIBILITY_STATE_INCOMPATIBLE {
			return fmt.Errorf("unexpected compatibility state")
		}
		require.Len(t, node.CompatibilityStatus.Issues, 1)
		require.Equal(t, "P115", node.CompatibilityStatus.Issues[0].Code)
		expectedDescription := "Plugin 'opentelemetry' is not available in Kong" +
			" gateway versions < 3.0."
		expectedResolution := "Please upgrade Kong Gateway to version '3.0' or above."
		require.Equal(t, expectedDescription, node.CompatibilityStatus.Issues[0].Description)
		require.Equal(t, expectedResolution,
			node.CompatibilityStatus.Issues[0].Resolution)
		return nil
	})
}
