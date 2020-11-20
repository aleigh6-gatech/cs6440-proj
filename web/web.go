package web

import (
	"strings"
	"github.com/martini-contrib/render"
	"github.com/go-martini/martini"
	"log"
	conf "coordinator/config"
	"coordinator/proxy"
	"coordinator/data_sync"
	"encoding/json"
)

var m martini.ClassicMartini

var config *conf.Config

type healthRow struct {
	Cluster string `json:"cluster"`
	Endpoint string `json:"endpoint"`
	Health bool  `json:"health"`
}

type dataSyncRow struct {
	Cluster string `json:"cluster"`
	Endpoint string `json:"endpoint"`
	TransactionSeq int `json:"transaction_seq"`
}

// StatusResponse endpoint status object
type StatusResponse struct {
	Healths []healthRow `json:"healths"`
	DataSyncs []dataSyncRow `json:"data_syncs"`
}

func splitEndpointFullname(fullname string) (string, string) {
	tokens := strings.Split(fullname, "#")
	return tokens[0], tokens[1]
}

func getStatusResponse() StatusResponse {

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
	for endpointFullname, seq := range dataSync.Cursors {
		cluster, endpoint := splitEndpointFullname(endpointFullname)
		row := dataSyncRow{
			Cluster: cluster,
			Endpoint: endpoint,
			TransactionSeq: seq,
		}
		dataSyncRows = append(dataSyncRows, row)
	}

	resp := StatusResponse {
		Healths: healthRows,
		DataSyncs: dataSyncRows,
	}

	return resp
}


// StartWeb starts web app
func StartWeb(_config *conf.Config) {
	config = _config

	m := martini.Classic()
	m.Use(render.Renderer(render.Options{
		Extensions: []string{".html"},
	}))

	m.Get("/admin", func(r render.Render){
		statusResp := getStatusResponse()

		inst := struct {
			HealthcheckInterval int
			ServersHealth []healthRow
			DataSync []dataSyncRow
		}{
			HealthcheckInterval: config.HealthCheckInterval,
			ServersHealth: statusResp.Healths,
			DataSync: statusResp.DataSyncs,
		}

		r.HTML(200, "index", inst)
	})

	m.Get("/", func() string {
		return "ok"
	})

	m.Get("/status", func() string {
		statusResp := getStatusResponse()

		b, err := json.Marshal(statusResp)
		if err != nil {
			return ""
		}
		return string(b)
	})

	// m.RunOnAddr(":8080")
	m.Run()
	log.Println("Martini started")
}