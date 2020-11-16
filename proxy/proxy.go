package proxy

import (
	"net/http/httptest"
	conf "coordinator/config"
	"coordinator/util"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"
)

var config *conf.Config
var healthStatus = make(map[string]bool)

var httpClient = http.Client{
	Timeout: 5 * time.Second,
}

func checkEndpoint(address string, path string) bool {
	fullURL := fmt.Sprintf("%v/%v", address, path)
	resp, err := http.Get(fullURL)
	if err != nil {
		return false
	}
	return resp.StatusCode < 400
}

func startHealthCheck() {
	go func() {
		for {
			for _, cluster := range config.Clusters {
				for _, endpoint := range cluster.Endpoints {
					if checkEndpoint(endpoint, "") {
						healthStatus[util.EndpointFullname(cluster.Name, endpoint)] = true
					} else {
						healthStatus[util.EndpointFullname(cluster.Name, endpoint)] = false
					}
				}
			}

			// sleep
			time.Sleep(time.Duration(config.HealthCheckInterval) * time.Second)
		}
	}()
	log.Println("Proxy starts health check backends")
}

// BestEndpointInCluster finds the most preferred healthy endpoint in a cluster
func BestEndpointInCluster(clusterName string) string {
	for _, cluster := range config.Clusters {
		if cluster.Name == clusterName {
			for _, endpoint := range cluster.Endpoints {
				healthKey := util.EndpointFullname(clusterName, endpoint)
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

func getSitePath(url string) string {
	if pos := strings.Index(url, "//"); pos >= 0 { // url does not contains protocol
		url = url[pos+2:]
	}

	tokens := strings.Split(url, "/")
	requestPath := "/" + strings.Join(tokens[1:], "/")
	return requestPath
}

func forwardRequest(endpoint string, req *http.Request, resp http.ResponseWriter) {
	// get request path
	sitePath := getSitePath(req.RequestURI)

	log.Printf("forwarding request to %s, path %s\n", endpoint, sitePath)
	endpointURL, _ := url.Parse(endpoint)

	// create the reverse proxy
	proxy := httputil.NewSingleHostReverseProxy(endpointURL)

	// Update the headers to allow for SSL redirection
	req.URL.Host = endpointURL.Host
	req.URL.Scheme = endpointURL.Scheme
	req.Header.Set("X-Forwarded-Host", req.Header.Get("Host"))
	req.Host = endpointURL.Host

	// Note that ServeHttp is non blocking and uses a go routine under the hood
	proxy.ServeHTTP(resp, req)
	log.Printf("request forwarded to %s, path %s\n", endpointURL, sitePath)
}

func urlMatch(pattern string, url string) bool {
	requestPath := getSitePath(url)
	log.Printf("path: %v, pattern %v, matches %v\n", url, pattern, strings.HasPrefix(requestPath, pattern))

	return strings.HasPrefix(requestPath, pattern)
}

func getClusterFromConfig(clusterName string) conf.Cluster {
	for _, cluster := range(config.Clusters) {
		if cluster.Name == clusterName {
			return cluster
		}
	}
	return conf.Cluster{}
}

func routeRequest(req *http.Request, resp http.ResponseWriter) {
	for _, route := range config.Routes {
		if urlMatch(route.Path, req.RequestURI) {
			// forward request to all the clusters
			for _, clusterName := range route.Clusters {
				bestEndpoint := BestEndpointInCluster(clusterName)

				// forward request to each endpoint in the cluster
				cluster := getClusterFromConfig(clusterName)

				for _, endpoint := range cluster.Endpoints {
					if endpoint != bestEndpoint {
						log.Printf("%v is not the best endpoint. Skip", endpoint)
						forwardRequest(bestEndpoint, req, httptest.NewRecorder())
					} else {
						forwardRequest(bestEndpoint, req, resp)
					}
				}
			}
		}
	}
}

func startListening() {
	http.HandleFunc("/", func(resp http.ResponseWriter, req *http.Request) {
		routeRequest(req, resp)
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
