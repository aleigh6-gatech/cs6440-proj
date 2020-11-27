#!/bin/bash
for file in ../client/common-data/fhir/*
do
	curl --request POST --location "http://localhost:8080/fhir" --header "Content-Type:application/json" -d "@$file"
done

for file in ../client/common-data/fhir/*
do
	curl --request POST --location "http://localhost:8081/fhir" --header "Content-Type:application/json" -d "@$file"
done