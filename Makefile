logsetd: *.go
	go build

test:
	go test
.PHONY: test

.data:
	mkdir .data

develop: logsetd .data
	LOGSET_STORE=.data ./logsetd
