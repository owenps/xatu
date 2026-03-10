package config

type Context struct {
	Name      string   `yaml:"name"`
	LogGroups []string `yaml:"log_groups"`
	Colour    string   `yaml:"colour"`
}
