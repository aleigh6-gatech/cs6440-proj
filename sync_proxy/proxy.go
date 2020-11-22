package syncProxy

import (
	"sync"
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

// HealthStatus is from full endpoint name to boolean
var HealthStatus = make(map[string]bool)

var httpClient = http.Client{
	Timeout: 5 * time.Second,
}

func startHealthCheck() {
	ticker := time.NewTicker(time.Duration(config.HealthCheckInterval) * time.Second)

	go func() {
		for {
			select {
			case <- ticker.C:
				for _, cluster := range config.Clusters {
					for _, endpoint := range cluster.Endpoints {
						if util.CheckEndpoint(endpoint, "") {
							HealthStatus[util.EndpointFullname(cluster.Name, endpoint)] = true
						} else {
							HealthStatus[util.EndpointFullname(cluster.Name, endpoint)] = false
						}
					}
				}
			}
		}
	}()
	log.Println("Proxy starts health check backends")
}

// BestEndpointInCluster finds the most preferred healthy endpoint in a cluster
func BestEndpointInCluster(clusterName string) string {
	for _, cluster := range config.Clusters {
		if cluster.Name == clusterName {
			for _, endpoint := range cluster.Endpoints {
				endpointFullname := util.EndpointFullname(clusterName, endpoint)
				if HealthStatus[endpointFullname] {
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

// ForwardRequest forwards request to endpoint
func ForwardRequest(endpoint string, req *http.Request, resp http.ResponseWriter) bool {
	// get request path
	sitePath := getSitePath(req.RequestURI)

	log.Printf("forwarding request to %s, path %s, original req %v\n", endpoint, sitePath, req)
	endpointURL, _ := url.Parse(endpoint)

	// create the reverse proxy
	proxy := httputil.NewSingleHostReverseProxy(endpointURL)

	// Update the headers to allow for SSL redirection
	req.URL.Host = endpointURL.Host
	req.URL.Scheme = endpointURL.Scheme
	req.Header.Set("X-Forwarded-Host", req.Header.Get("Host"))
	req.Host = endpointURL.Host
	log.Printf("updated req %v\n", req)

	// Note that ServeHttp is non blocking and uses a go routine under the hood

	// health check before sending
	if !util.CheckEndpoint(endpoint, "") {
		log.Printf("forwarding request failed. Health check failed for %s\n", endpointURL)
		return false
	} else {
		proxy.ServeHTTP(resp, req)
		log.Printf("request forwarded to %s, path %s\n", endpointURL, sitePath)
		return true
	}
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

func routeRequest(req *http.Request, resp http.ResponseWriter, requestSeq int) {
	for _, route := range config.Routes {
		if urlMatch(route.Path, req.RequestURI) {
			// forward request to all the listed clusters
			for _, clusterName := range route.Clusters {
				bestEndpoint := BestEndpointInCluster(clusterName)

				// forward request to each endpoint in the cluster
				cluster := getClusterFromConfig(clusterName)

				for _, endpoint := range cluster.Endpoints {
					endpointFullname := util.EndpointFullname(clusterName, endpoint)

					// check endpoint health first
					if !HealthStatus[endpointFullname] {
						continue
					}

					if !util.CheckEndpoint(endpoint, "") {
						HealthStatus[endpointFullname] = false
						continue
					}

					// check best endpoint to determine which response to return to the user
					if endpoint != bestEndpoint {
						log.Printf("%v is not the best endpoint. Request will be forwarded. Skip writing the response from it", endpoint)

						ForwardRequest(bestEndpoint, req, httptest.NewRecorder())
					} else {
						ForwardRequest(bestEndpoint, req, resp)
					}

					if requestSeq != -1 { // POST request
						Cursors[endpointFullname] = requestSeq
					}
				}
			}
			break
		}
	}
}

func startListening() {
	http.HandleFunc("/", func(resp http.ResponseWriter, req *http.Request) {
		requestSeq := -1

		if req.Method == "POST" {
			requestSeq = AddTransaction(req)
		}

		routeRequest(req, resp, requestSeq)
	})

	fmt.Printf("Starting server at port %v\n", config.Port)
	if err := http.ListenAndServe(fmt.Sprintf("localhost:%v", config.Port), nil); err != nil {
		log.Fatal(err)
	}
}

// StartProxy starts a proxy with config
func StartProxy(newConfig *conf.Config) {
	config = newConfig

	var wg sync.WaitGroup

	wg.Add(1)
	go startHealthCheck()

	wg.Add(1)
	go startListening()

	wg.Add(1)
	go startDataSync()

	wg.Wait()
}
