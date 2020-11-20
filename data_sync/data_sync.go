package dataSync

import (
	"log"
	"time"
	conf "coordinator/config"
	"net/http"
	"coordinator/util"
)

var config *conf.Config

type WrapRequest struct {
	Seq int
	request http.Request
}


var Transactions = []WrapRequest{}

// Cursors stores the latest transaction seq number for each endpoint
var Cursors = make(map[string]int)

var counter int

// AddTransaction adds transaction into the transactions cache
func AddTransaction(req *http.Request) int {
	// save a copy of the request to transacitons
	leng := len(Transactions)
	var latestSeq int
	if leng == 0 {
		latestSeq = 0
	} else {
		latestSeq = Transactions[leng-1].Seq + 1
	}
	newTransaction := WrapRequest{ Seq: latestSeq, request: *req }
	Transactions = append(Transactions, newTransaction)

	return latestSeq
}

func backfillDataFor(clusterName string, endpoint string) {
	// find start in transactions
	endpointFullname := util.EndpointFullname(clusterName, endpoint)
	cursor := Cursors[endpointFullname]
	startIdx := -1
	// find the starting point for the endpoint to backfill
	for i, tx := range(Transactions) {
		if tx.Seq > cursor {
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
		for _, tx := range(Transactions[startIdx:]) {
			// send the request again, and update Cursors
			req := tx.request
			contentType := req.Header.Get("Content-type")
			_, err := http.Post(req.RequestURI, contentType, req.Body)
			if err != nil {
				break
			}

			Cursors[endpointFullname]++
		}
	}
}

// swipe transactions
func swipeTxs() {
	log.Println("swiping txs")

	if len(Cursors) == 0 {
		return
	}


	minCursor := -1
	for _, cursor := range(Cursors) {
		if minCursor == -1 {
			minCursor = cursor
		} else if cursor < minCursor {
			minCursor = cursor
		}
	}

	swipeCount := 0
	for (len(Transactions) > 0) {
		tx := Transactions[0]
		if tx.Seq < minCursor {
			Transactions = Transactions[1:]
			swipeCount++
		} else{
			break
		}
	}
	log.Printf("Swiped data: %v\n", swipeCount)
}

func backfillData() {
	for _, cluster := range config.Clusters {
		// looping endpoints
		for _, endpoint := range cluster.Endpoints {
			backfillDataFor(cluster.Name, endpoint)
		}
	}
}

// StartDataSync starts a thread to monitor and sync data
func StartDataSync(newConfig *conf.Config) {
	config = newConfig

	backfillTicker := time.NewTicker(time.Duration(config.DataSyncInterval) * time.Second)
	swipeTicker := time.NewTicker(5 * time.Second)


	// init transacitons, Cursors
	for _, cluster := range config.Clusters {
		for _, endpoint := range cluster.Endpoints {
			Cursors[util.EndpointFullname(cluster.Name, endpoint)] = 0
		}
	}

	for {
		select {
			case <- backfillTicker.C:
				backfillData()
			case <- swipeTicker.C:
				swipeTxs()
		}
	}
}