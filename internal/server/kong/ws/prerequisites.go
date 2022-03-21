package ws

import (
	"strings"

	model "github.com/kong/koko/internal/gen/grpc/kong/admin/model/v1"
	grpcKongUtil "github.com/kong/koko/internal/gen/grpc/kong/util/v1"
	"github.com/kong/koko/internal/resource"
	"github.com/kong/koko/internal/status"
)

func checkPreReqs(attr nodeAttributes,
	checks []*grpcKongUtil.DataPlanePrerequisite,
) []*model.Condition {
	var res []*model.Condition
	for _, check := range checks {
		if check.GetRequiredPlugins() != nil {
			pluginCheck := check.GetRequiredPlugins()
			missingPlugins := checkMissingPlugins(pluginCheck.RequiredPlugins,
				attr.Plugins)
			if len(missingPlugins) > 0 {
				condition := conditionForMissingPlugins(missingPlugins)
				res = append(res, condition)
			}
		}
	}
	return res
}

func conditionForMissingPlugins(plugins []string) *model.Condition {
	return &model.Condition{
		Code: status.DPMissingPlugin,
		Message: status.MessageForCode(status.DPMissingPlugin,
			strings.Join(plugins, ", ")),
		Severity: resource.SeverityError,
	}
}

func checkMissingPlugins(requiredPlugins []string, nodePlugins []string) []string {
	nodePluginMap := map[string]struct{}{}
	for _, plugin := range nodePlugins {
		nodePluginMap[plugin] = struct{}{}
	}

	var res []string
	for _, requiredPlugin := range requiredPlugins {
		if _, ok := nodePluginMap[requiredPlugin]; !ok {
			res = append(res, requiredPlugin)
		}
	}
	return res
}
