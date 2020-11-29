GaTech CS6440 individual project - High availability replication system
---

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