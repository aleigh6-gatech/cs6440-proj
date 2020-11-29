package syncProxy

import (
	"fmt"
	"io/ioutil"
	"bytes"
	// "net/http/httptest"
	"log"
	"time"
	"net/http"
	"coordinator/util"
)


type WrapRequest struct {
	Seq int
	request http.Request
	Path string
	Body []byte
	Routed bool
}


var Transactions = []*WrapRequest{}

// Cursors stores the latest transaction seq number that was executed for each endpoint
var Cursors = make(map[string]int)

var counter int

var NumTxs = 0

func PrintTxs() {
	log.Printf("Print Txs\n")
	for _, tx := range Transactions {
		log.Printf("[%v %v len(%v) %v %p]\n", tx.Seq, tx.Path, len(tx.Body), tx.Routed, &tx)
	}
	log.Printf("\n")
}

// AddTransaction adds transaction into the transactions cache
func AddTransaction(req *http.Request) *WrapRequest {
	// save a copy of the request to transacitons

	var b bytes.Buffer
	b.ReadFrom(req.Body)
	req.Body = ioutil.NopCloser(&b)
	nb := make([]byte, len(b.Bytes()))
	copy(nb, b.Bytes())

	newTransaction := &WrapRequest{
		Seq: NumTxs,
		request: *util.CloneRequest(req),
		Path: req.URL.Path,
		Body: nb,
	}

	Transactions = append(Transactions, newTransaction)
	NumTxs ++

	return newTransaction
}

func backfillDataFor(clusterName string, endpoint string) {

	// find start in transactions
	endpointFullname := util.EndpointFullname(clusterName, endpoint)
	cursor := Cursors[endpointFullname]
	startIdx := -1

	// find the starting point for the endpoint to backfill
	for i, tx := range(Transactions) {
		if tx.Seq > cursor && tx.Routed {
			startIdx = i
			break
		}
	}

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
			newURL := fmt.Sprintf("%v%v", endpoint, tx.Path)
			req, err := http.NewRequest("POST", newURL, bytes.NewReader(tx.Body))
			req.Header.Add("Content-Type", "application/json")

			httpClient.Timeout = time.Duration(60 * time.Second)
			_, err = httpClient.Do(req)

			if err == nil {
				log.Printf("backfill request success: tx.seq %v\n", tx.Seq)
				Cursors[endpointFullname] = tx.Seq
			} else {
				log.Printf("backfill request failed : %v\n", err)
				break
			}
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
		if cursor < minCursor || minCursor == -1 {
			minCursor = cursor
		}
	}

	swipeCount := 0
	for (len(Transactions) > 0) {
		tx := Transactions[0]
		if tx.Seq <= minCursor {
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

		log.Printf("After backfill\nlen txs %v\nNumTxs: %v\nCursors: %v\n", len(Transactions), NumTxs, Cursors)
		PrintTxs()
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