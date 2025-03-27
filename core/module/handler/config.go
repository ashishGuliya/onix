package handler

import (
	"github.com/ashishGuliya/onix/pkg/model"
	"github.com/ashishGuliya/onix/pkg/plugin"
)

type HandlerType string

const (
	HandlerTypeStd    HandlerType = "std"
	HandlerTypeRegSub HandlerType = "regSub"
	HandlerTypeNPSub  HandlerType = "npSub"
	HandlerTypeLookup HandlerType = "lookUp"
)

type pluginCfg struct {
	SchemaValidator *plugin.Config  `yaml:"schemaValidator,omitempty"`
	SignValidator   *plugin.Config  `yaml:"signValidator,omitempty"`
	Publisher       *plugin.Config  `yaml:"publisher,omitempty"`
	Signer          *plugin.Config  `yaml:"signer,omitempty"`
	Router          *plugin.Config  `yaml:"router,omitempty"`
	Cache           *plugin.Config  `yaml:"cache,omitempty"`
	KeyManager      *plugin.Config  `yaml:"keyManager,omitempty"`
	Middleware      []plugin.Config `yaml:"middleware,omitempty"`
	Steps           []plugin.Config
}

type Config struct {
	Plugins      pluginCfg `yaml:"plugins"`
	Steps        []string
	Type         HandlerType
	RegistryURL  string `yaml:"registryUrl"`
	Role         model.Role
	SubscriberID string `yaml:"subscriberId"`
	Trace        map[string]bool
}
