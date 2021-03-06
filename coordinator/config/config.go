package config

// Config of the coordinator
type Config struct {
	Clusters []Cluster

	Routes []Route

	HealthCheckInterval int `yaml:"health_check_interval" default:"3"`
	DataSyncInterval int `yaml:"data_sync_interval" default:"1"`
	HostIP string `yaml:"host_ip" default:"127.0.0.1"`
	Port int
	ProxyControlPort int `yaml:"proxy_control_port"`
}

type Cluster struct {
	Name string
	Endpoints []string // address of endpoints
}

type Route struct {
	Path string
	Clusters []string
}