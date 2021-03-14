package controller

import (
	"encoding/json"
	"fmt"

	"github.com/RManLuo/XrayR/api"
	"github.com/xtls/xray-core/core"
	"github.com/xtls/xray-core/infra/conf"
)

//OutboundBuilder build freedom outbund config for addoutbound
func OutboundBuilder(nodeInfo *api.NodeInfo) (*core.OutboundHandlerConfig, error) {
	outboundDetourConfig := &conf.OutboundDetourConfig{}
	outboundDetourConfig.Protocol = "freedom"
	outboundDetourConfig.Tag = fmt.Sprintf("%s_%d", nodeInfo.NodeType, nodeInfo.Port)
	// Protocol setting
	proxySetting := &conf.FreedomConfig{
		DomainStrategy: "Asis",
	}
	var setting json.RawMessage
	setting, err := json.Marshal(proxySetting)
	if err != nil {
		return nil, fmt.Errorf("Marshal proxy %s config fialed: %s", nodeInfo.NodeType, err)
	}
	outboundDetourConfig.Settings = &setting
	return outboundDetourConfig.Build()
}
