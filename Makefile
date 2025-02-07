.PHONY: build build-windows clean

build:
	go build -o bin/yaga-bhop ./cmd/main.go

build-windows:
    set GOOS=windows && set GOARCH=amd64 && go build -o bin/yaga-bhop.exe ./cmd/main.go

clean:
	rm -rf bin/