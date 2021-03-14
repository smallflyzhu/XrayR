package panel

import (
	"github.com/RManLuo/XrayR/api"
	"github.com/RManLuo/XrayR/service/controller"
)

type Config struct {
	LogConfig   *LogConfig     `mapstructure:"Log"`
	NodesConfig []*NodesConfig `mapstructure:"Nodes"`
}

type NodesConfig struct {
	PanelType        string             `mapstructure:"PanelType"`
	ApiConfig        *api.Config        `mapstructure:"ApiConfig"`
	ControllerConfig *controller.Config `mapstructure:"ControllerConfig"`
}

type LogConfig struct {
	Level      string `mapstructure:"Level"`
	AccessPath string `mapstructure:"AccessPath"`
	ErrorPath  string `mapstructure:"ErrorPath"`
}
