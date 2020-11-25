package syncProxy

import (
	"context"
	"net/http/httptest"
	"log"
	"time"
	"net/http"
	"coordinator/util"
)


type WrapRequest struct {
	Seq int
	request http.Request
}


var Transactions = []WrapRequest{}

// Cursors stores the latest transaction seq number that was executed for each endpoint
var Cursors = make(map[string]int)

var counter int

var NumTxs = 0

// AddTransaction adds transaction into the transactions cache
func AddTransaction(req *http.Request) int {
	// save a copy of the request to transacitons
	newTransaction := WrapRequest{ Seq: NumTxs, request: *req.Clone(context.TODO()) }
	Transactions = append(Transactions, newTransaction)
	NumTxs ++

	return NumTxs-1
}

func backfillDataFor(clusterName string, endpoint string) {

	// find start in transactions
	endpointFullname := util.EndpointFullname(clusterName, endpoint)
	cursor := Cursors[endpointFullname]
	startIdx := -1

	log.Printf("DEBUG cursor %v for %v\n", cursor, endpoint)

	// find the starting point for the endpoint to backfill
	for i, tx := range(Transactions) {
		if tx.Seq > cursor {
			startIdx = i
			break
		}
	}

	log.Printf("DEBUG: startIndex %v, transactionns length %v\n", startIdx, len(Transactions))

	if startIdx != -1 && startIdx < len(Transactions) { // needs backfill
		// check connection
		if !util.CheckEndpoint(Enabled[endpoint], endpoint, "") {
			log.Printf("Endpoint %v check failed during backfill. Postponed\n", endpoint)
			return
		}
		HealthStatus[endpointFullname] = true

		// backfill data
		log.Printf("Backfill for %v %v, started from %v\n", clusterName, endpoint, startIdx)

		for _, tx := range(Transactions[startIdx:]) {
			// send the request again, and update Cursors
			req := tx.request

			// duplicate request
			replay, err := http.NewRequest(req.Method, req.RequestURI, req.Body)
			if err != nil {
				log.Printf("Error when creating replay request %s %s, %v\n", clusterName, endpoint, err)
				break
			}
			for header, values := range req.Header {
				for _, value := range values {
					replay.Header.Add(header, value)
				}
			}
			log.Printf("DEBUG Request duplication, %v\nDup: %v\n", req, replay)
			resp := httptest.NewRecorder()

			// check endpoint health
			log.Printf("DEBUG health status %v\n", HealthStatus)
			if !HealthStatus[endpointFullname] {
				log.Printf("Endpoint %v not healthy. Backfill postponed.", endpointFullname)
				return
			}
			ForwardRequest(endpoint, replay, resp)

			Cursors[endpointFullname] = tx.Seq
		}
	}
}

// swipe transactions
func swipeTxs() {
	log.Println("swiping txs")

	log.Printf("DEBUG len %v\n %v %v\n",  len(Transactions), Transactions, Cursors)

	if len(Cursors) == 0 {
		return
	}

	minCursor := -1
	for _, cursor := range(Cursors) {
		log.Printf("cursor value: %v\n", cursor)
		if cursor < minCursor {
			minCursor = cursor
		}
	}

	log.Printf("DEBUG min cursor %v\n", minCursor)

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

		log.Printf("After backfill\nlen txs %v\ntxs: %v\nNumTxs: %v\nCursors: %v\n", len(Transactions), Transactions, NumTxs, Cursors)
	}
}

func fetchDbCount() {

}

// StartDataSync starts a thread to monitor and sync data
func startDataSync() {
	backfillTicker := time.NewTicker(time.Duration(config.DataSyncInterval) * time.Second)

	// init transacitons, Cursors
	for _, cluster := range config.Clusters {
		for _, endpoint := range cluster.Endpoints {
			Cursors[util.EndpointFullname(cluster.Name, endpoint)] = -1
		}
	}

	for {
		select {
			case <- backfillTicker.C:
				backfillData()
				fetchDbCount()
				swipeTxs()
		}
	}
}