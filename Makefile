
.PHONY: compile
compile:
	go get github.com/karalabe/xgo
	xgo -ldflags="-w" -out bin/pensum-v0.1.0  --targets=windows/amd64,linux/amd64 ./