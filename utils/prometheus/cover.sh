#! /usr/bin/env bash

go test -c -covermode=count -coverpkg ./  -o prometheus.test
./prometheus.test -test.coverprofile coverage.cov
go tool cover -html=./coverage.cov -o coverage.html
rm prometheus.test coverage.cov
