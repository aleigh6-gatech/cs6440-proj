package config

// Config of the coordinator
type Config struct {
	Clusters []cluster

	Routes []route

	HealthCheckInterval int `yaml:"health_check_interval" default:"3"`
}

type cluster struct {
	Name string
	Backends []string
}

type route struct {
	Path string
	Clusters []string
}