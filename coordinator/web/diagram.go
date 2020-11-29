package web

import (
	"fmt"
	"bytes"
	"coordinator/sync_proxy"
	"coordinator/util"
	b64 "encoding/base64"
	gv "github.com/goccy/go-graphviz"
	"log"
)

func generateDiagramB64() string {
	// generate diagram
	g := gv.New()
	graph, err := g.Graph()
	if err != nil {
		log.Fatal(err)
	  	return ""
	}
	defer func() {
		if err := graph.Close(); err != nil {
		  	log.Fatal(err)
		}
		g.Close()
	}()

	userNode, _ := graph.CreateNode("User")
	userNode.SetShape("plaintext")
	userNode.SetLabel("User")

	coordinatorNode, _ := graph.CreateNode("Coordinator")
	coordinatorNode.SetLabel(fmt.Sprintf("Coordinator@%v", config.Port))
	coordinatorNode.SetShape("box")

	edge, _ := graph.CreateEdge("", userNode, coordinatorNode)
	edge.SetArrowTail("normal")

	for _, cluster := range config.Clusters {
		clusterSub := graph.SubGraph(fmt.Sprintf("cluster_%s", cluster.Name), 1)
		clusterSub.SetStyle("solid")
		clusterSub.SetLabel(cluster.Name)

		for _, endpoint := range cluster.Endpoints {
			endpointNode, _ := clusterSub.CreateNode(endpoint)
			endpointNode.SetLabel(endpoint)
			endpointNode.SetStyle("filled")

			endpointFullname := util.EndpointFullname(cluster.Name, endpoint)

			if syncProxy.HealthStatus[endpointFullname] {
				endpointNode.SetFillColor("springgreen")
			} else {
				endpointNode.SetFillColor("crimson")
			}

			edge, _ = graph.CreateEdge("", coordinatorNode, endpointNode)
			edge.SetArrowTail("normal")
		}
	}

	var buf bytes.Buffer

	if err = g.Render(graph, "png", &buf); err != nil {
	  	log.Fatal(err)
	}

	return b64.StdEncoding.EncodeToString(buf.Bytes())
}
