Compoents of coordinator
---

* Proxy
    * application (1) -> (N) endpoints
    * application is unordered
    * endpoints are ordered
    * data is replicated amoung all endpoinds in an application
    * defines routes
    * Proxy sends POST request to endpoint1
    * Control plane:
        * monitors health of data plane
        * tells other proxy to enable egress or disable egress
    * data plane
        * accepts request to enable egress or disable egress
* Coordinator
    * web UI service
    * keep data in sync,
        * knows the last completed request to each endpoint. Small Redis to store it
        * If endpoint is recovered, it will spin up a thread, to follow up all write request to the endpoint
        * we don't consider locking issue here, so complicated.




