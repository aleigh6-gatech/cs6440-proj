package data_sync

import (
	"log"
	"time"
	conf "coordinator/config"
	"net/http"
	"coordinator/util"
)

var config *conf.Config

type wrapRequest struct {
	seq int
	request http.Request
}


var transactions []wrapRequest

var cursors map[string]int

var counter int

// AddTransaction adds transaction into the transactions cache
func AddTransaction(req *http.Request) {
	// save a copy of the request to transacitons
	leng := len(transactions)
	latestSeq := transactions[leng-1].seq + 1
	newTransaction := wrapRequest{ seq: latestSeq, request: *req }
	transactions = append(transactions, newTransaction)
}


func backfillData(clusterName string, endpoint string) {
	// find start in transactions
	endpointFullname := util.EndpointFullname(clusterName, endpoint)
	cursor := cursors[endpointFullname]
	startIdx := -1
	// find the starting point for the endpoint to backfill
	for i, tx := range(transactions) {
		if tx.seq > cursor {
			startIdx = i
			break
		}
	}

	if startIdx != -1 { // needs backfill
		// check connection
		if !util.CheckEndpoint(endpoint, "") {
			log.Printf("Endpoint %v check failed during backfill. Postponed\n", endpoint)
			return
		}

		// backfill data
		for _, tx := range(transactions[startIdx:]) {
			// send the request again, and update cursors
			req := tx.request
			contentType := req.Header.Get("Content-type")
			_, err := http.Post(req.RequestURI, contentType, req.Body)
			if err != nil {
				break
			}

			cursors[endpointFullname]++
		}
	}
}

// swipe transactions
func swipeTxs() {
	log.Println("swiping txs")

	if len(cursors) == 0 {
		return
	}

	minCursor := -1
	for _, cursor := range(cursors) {
		if minCursor == -1 {
			minCursor = cursor
		} else if cursor < minCursor {
			minCursor = cursor
		}
	}

	swipeCount := 0
	for {
		tx := transactions[0]
		if tx.seq < minCursor {
			transactions = transactions[1:]
			swipeCount++
		} else{
			break
		}
	}
	log.Printf("Swiped data: %v\n", swipeCount)
}

// StartDataSync starts a thread to monitor and sync data
func StartDataSync(newConfig *conf.Config) {
	config = newConfig

	// init transacitons, cursors
	for _, cluster := range config.Clusters {
		for _, endpoint := range cluster.Endpoints {
			cursors[util.EndpointFullname(cluster.Name, endpoint)] = 0
		}
	}

	// backfill data loop
	go func() {
		for {
			// looping clusters
			for _, cluster := range config.Clusters {
				// looping endpoints
				for _, endpoint := range cluster.Endpoints {
					backfillData(cluster.Name, endpoint)
				}
			}


			// sleep
			time.Sleep( time.Duration(config.DataSyncInterval) * time.Second )
		}
	}()


	// swipe transactions
	go func() {
		for {
			swipeTxs()

			// sleep
			time.Sleep( time.Duration(5) * time.Second )
		}
	}()

}