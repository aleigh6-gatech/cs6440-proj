package web

import (
	"strings"
	"github.com/martini-contrib/render"
	"github.com/go-martini/martini"
	"log"
	conf "coordinator/config"
	"coordinator/proxy"
	"coordinator/data_sync"
)

var m martini.ClassicMartini

var config *conf.Config

type healthRow struct {
	Cluster string
	Endpoint string
	Health bool
}

type dataSyncRow struct {
	Cluster string
	Endpoint string
	TransactionSeq int
}

func splitEndpointFullname(fullname string) (string, string) {
	tokens := strings.Split(fullname, "#")
	return tokens[0], tokens[1]
}


// StartWeb starts web app
func StartWeb(_config *conf.Config) {
	config = _config

	m := martini.Classic()
	m.Use(render.Renderer(render.Options{
		Extensions: []string{".html"},
	}))

	m.Get("/admin", func(r render.Render){

		// prepare servers health
		healthRows := []healthRow{}
		for endpointFullname, health := range proxy.HealthStatus {
			cluster, endpoint := splitEndpointFullname(endpointFullname)
			row := healthRow{
				Cluster: cluster,
				Endpoint: endpoint,
				Health: health,
			}
			healthRows = append(healthRows, row)
		}

		// prepare sync data
		dataSyncRows := []dataSyncRow{}
		for endpointFullname, seq := range data_sync.Cursors {
			cluster, endpoint := splitEndpointFullname(endpointFullname)
			row := dataSyncRow{
				Cluster: cluster,
				Endpoint: endpoint,
				TransactionSeq: seq,
			}
			dataSyncRows = append(dataSyncRows, row)
		}

		inst := struct {
			HealthcheckInterval int
			ServersHealth []healthRow
			DataSync []dataSyncRow
		}{
			HealthcheckInterval: config.HealthCheckInterval,
			ServersHealth: healthRows,
			DataSync: dataSyncRows,
		}

		r.HTML(200, "index", inst)
	})

	m.Get("/", func() string {
		return "ok"
	})

	// m.RunOnAddr(":8080")
	m.Run()
	log.Println("Martini started")
}