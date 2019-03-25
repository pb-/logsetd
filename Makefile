logsetd: *.go
	go build

test:
	go test
.PHONY: test

.data:
	mkdir .data

develop: logsetd .data
	LOGSETD_STORE=.data ./logsetd
