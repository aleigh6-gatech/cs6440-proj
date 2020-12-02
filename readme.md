GaTech CS6440 individual project - High availability replication system
---

Live site demo: http://138.91.241.95:3000/admin


### Introduction

This project develops a proxy (called coordinator) stands in between of user clients and multiple clusters of servers. The purpose is to improve system availability, by routing request to the healthy endpoints (replicas) in the designated cluster.

In the demo, we assume that each endpoint has their own data storage. The coordinator also ensure the data are synchronized among all replicas, by replaying POST request to them. In practical cases, data synchronzization can also be handled by distributed database system, rather than replaying requests.

For each request, the coordinator will send back the response from a healthy endpoint that has the highest priority in the cluster.

### Structure

The structure of this project:

```
├── client          # client script
├── coordinator     # coordinator system
├── demo            # demonstration related files and configurations
├── makefile        # make tasks
└── readme.md
```

### Prerequisites

In order to set up the project on local, the listed tools are required to be installed:

* Git
* Docker-compose
* Golang 1.14+

The project setup instruction is tested on Mac OS X.

To set up a demo project on local:

```
git clone https://github.com/aleigh6-gatech/cs6440-proj.git
cd cs6440-proj
make setup
make start
```
