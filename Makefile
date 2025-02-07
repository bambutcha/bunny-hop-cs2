.PHONY: build build-windows clean

build:
	go build -o bin/yaga-bhop ./cmd/main.go

build-windows:
	GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -o bin/yaga-bhop.exe ./cmd/main.go

clean:
	rm -rf bin/