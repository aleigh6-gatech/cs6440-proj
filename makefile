.SILENT:
setup:
	echo "========= Checking prerequisites ..."
	CMDS="go docker-compose"
	for i in go docker-compose; do \
		command -v $i >/dev/null && continue || { echo "$i is not installed. Exit."; exit 1; } \
	done

	echo "========= Fetching sample EHR application"
	git clone https://github.com/hapifhir/hapi-fhir-jpaserver-starter.git sample-ehr/
	cp demo/docker-compose* sample-ehr/

	echo "========= Starting app1 on localhost:8080"
	cd sample-ehr && docker-compose up -d
	echo "========= App1 started"

	echo "========= Starting app2 on localhost:8081"
	cd sample-ehr && docker-compose -f docker-compose2.yml up -d
	echo "========= App2 started"

	echo "========= Waiting for app1 to be live, please wait patiently"
	sleep 5
	curl -f -m 10 --retry 30 --retry-delay 5 -S -I --retry-connrefused 127.0.0.1:8080

	echo "========= Waiting for app2 to be live"
	curl -f -m 10 --retry 30 --retry-delay 5 -S -I --retry-connrefused 127.0.0.1:8081

	echo "========= Setting up initial data"
	cd demo && ./setup_db.sh

start:
	echo "========= Starting app1 on localhost:8080"
	cd sample-ehr && docker-compose up -d
	echo "========= App1 started"

	echo "========= Starting app2 on localhost:8081"
	cd sample-ehr && docker-compose -f docker-compose2.yml up -d
	echo "========= App2 started"

	echo "========= Waiting for app1 to start"
	sleep 5
	curl -f -m 10 --retry 30 --retry-delay 5 -S -I --retry-connrefused 127.0.0.1:8080

	echo "========= Waiting for app2 to start"
	curl -f -m 10 --retry 30 --retry-delay 5 -S -I --retry-connrefused 127.0.0.1:8081

	echo "========= Starting coordinator"
	cd coordinator && go run coordinator.go

stop:
	cd ./sample-ehr && docker-compose down --remove-orphans && echo "app1 stopped"
	cd ./sample-ehr && docker-compose -f docker-compose2.yml down --remove-orphans && echo "app2 stopped"

clean: stop
	cd ./sample-ehr && docker-compose rm
	cd ./sample-ehr && docker-compose -f docker-compose2.yml rm

watch:
	tail -f coordinator/*.log