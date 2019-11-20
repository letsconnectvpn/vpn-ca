.PHONY: fmt clean

_bin/vpn-ca: vpn-ca/main.go
	go build -o $@ vpn-ca/main.go

fmt:
	gofmt -s -w vpn-ca/main.go

clean:
	rm -f _bin/*
