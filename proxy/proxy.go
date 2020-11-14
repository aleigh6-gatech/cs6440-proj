package proxy

import (
	"strings"
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

func forwardRequest(req *http.Request, endpoint string) {
	log.Printf("request forwarded to %s\n", endpoint)

}

func urlMatch(pattern string, url string) bool {

	log.Printf("path: %v\n", url)
	if pos := strings.Index(url, "//"); pos >= 0 { // url does not contains protocol
		url = url[pos+2:]
	}

	tokens := strings.Split(url, "/")
	requestPath := "/" + strings.Join(tokens[1:], "/")
	log.Printf("path: %v, pattern %v, matches %v\n", url, pattern, strings.HasPrefix(requestPath, pattern))

	return strings.HasPrefix(requestPath, pattern)
}


func routeRequest(req *http.Request) {
	for _, route := range config.Routes {
		if urlMatch(route.Path, req.RequestURI) {
			// forward request to all the clusters
			for _, clusterName := range route.Clusters {
				bestEndpoint := BestEndpointInCluster(clusterName)

				forwardRequest(req, bestEndpoint)
			}
		}
	}
}


func startListening() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request){

		routeRequest(r)
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