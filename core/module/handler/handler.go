package handler

import (
	"fmt"

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
	Plugins     pluginCfg `yaml:"plugins"`
	Steps       []string
	Type        HandlerType
	RegistryURL string `yaml:"registryUrl"`
}

// Step represents a named step
type Step string

const (
	StepInitialize Step = "initialize"
	StepValidate   Step = "validate"
	StepProcess    Step = "process"
	StepFinalize   Step = "finalize"
)

// ValidSteps ensures only allowed values are accepted
var ValidSteps = map[Step]bool{
	StepInitialize: true,
	StepValidate:   true,
	StepProcess:    true,
	StepFinalize:   true,
}

// Custom YAML unmarshalling to validate step names
func (s *Step) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var stepName string
	if err := unmarshal(&stepName); err != nil {
		return err
	}

	step := Step(stepName)
	if !ValidSteps[step] {
		return fmt.Errorf("invalid step: %s", stepName)
	}
	*s = step
	return nil
}
