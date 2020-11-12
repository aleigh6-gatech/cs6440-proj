package config

// Config of the coordinator
type Config struct {
	Clusters []Cluster

	Routes []Route
}

// Cluster that behind the proxy
type Cluster struct {
	Name string
	Backends []string
}

// Route defines path and target clusters
type Route struct {
	Path string
	Clusters []string
}