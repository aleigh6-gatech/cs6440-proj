package proxy

import (
	"net/http"
	"fmt"
	"time"
	 conf "coordinator/config"
	 "log"
)

var config *conf.Config
var healthStatus = make(map[string]bool)

var httpClient = http.Client{
	Timeout: 5 * time.Second,
}

func checkEndpoint(address string, path string) bool {
	return true

	// client.get
	// resp, _ := http.Get(address)

	// return resp.StatusCode < 400
}

// EndpointHealthKey returns the key value for check endpoint health status.
func EndpointHealthKey(clusterName string, endpoint string) string {
	return fmt.Sprintf("%v#%v", clusterName, endpoint)
}

func startHealthCheck() {
	go func() {
		for {
			for _, cluster := range config.Clusters {
				for _, endpoint := range cluster.Endpoints {
					if checkEndpoint(endpoint, "") {
						healthStatus[EndpointHealthKey(cluster.Name, endpoint)] = true
					} else {
						healthStatus[EndpointHealthKey(cluster.Name, endpoint)] = false
					}
				}
			}

			// sleep
			time.Sleep( time.Duration(config.HealthCheckInterval) * time.Second )
		}
	}()
	log.Println("Proxy starts health check backends")
}

// BestEndpointInCluster finds the most preferred healthy endpoint in a cluster
func BestEndpointInCluster(clusterName string) string {
	for _, cluster := range config.Clusters {
		if cluster.Name == clusterName {
			for _, endpoint := range cluster.Endpoints {
				healthKey := EndpointHealthKey(clusterName, endpoint)
				if healthStatus[healthKey] {
					return endpoint
				}
			}
		}
	}
	return ""
}

// UpdateConfig updates config object
func UpdateConfig(_config *conf.Config) {
	config = _config
}

func startListening() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request){
        fmt.Fprintf(w, "Hello!")
	})

    fmt.Printf("Starting server at port %v\n", config.Port)
    if err := http.ListenAndServe(fmt.Sprintf("localhost:%v", config.Port), nil); err != nil {
        log.Fatal(err)
    }
}


// StartProxy starts a proxy with config
func StartProxy(newConfig *conf.Config) {
	config = newConfig

	startHealthCheck()

	startListening()
}