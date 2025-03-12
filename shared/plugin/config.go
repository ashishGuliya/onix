package plugin

type PublisherCfg struct {
	ID     string            `yaml:"id"`
	Config map[string]string `yaml:"config"`
}

type ValidatorCfg struct {
	ID     string            `yaml:"id"`
	Config map[string]string `yaml:"config"`
}

type Config struct {
	ID     string            `yaml:"id"`
	Config map[string]string `yaml:"config"`
}

type ManagerConfig struct {
	Root          string   `yaml:"root"`
	PluginZipPath string   `yaml:"pluginZipPath"`
	Plugins       []string `yaml:"plugins"`
}
