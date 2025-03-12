package config

import (
	"fmt"
	"onix/shared/plugin"
)

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

// Config struct for loading YAML
type Config struct {
	Steps []Step `yaml:"steps"`
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

type ModuleCfg struct {
	Name       string `yaml:"name"`
	Type       string
	Path       string    `yaml:"path"`
	Plugins    PluginCfg `yaml:"plugin"`
	TargetType string    `yaml:"targetType"`
	Steps      []string
}

type PluginCfg struct {
	SchemaValidator *plugin.Config  `yaml:"schemaValidator,omitempty"`
	SignValidator   *plugin.Config  `yaml:"signValidator,omitempty"`
	Publisher       *plugin.Config  `yaml:"publisher,omitempty"`
	Signer          *plugin.Config  `yaml:"signer,omitempty"`
	Router          *plugin.Config  `yaml:"router,omitempty"`
	PreProcessors   []plugin.Config `yaml:"preProcessors,omitempty"`
	PostProcessors  []plugin.Config `yaml:"postProcessors,omitempty"`
	Steps           []plugin.Config
}
