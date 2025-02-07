.PHONY: build build-windows clean

build:
	go build -o bin/yaga-bhop ./cmd/main.go

build-windows:
	GOOS=windows GOARCH=amd64 go build -o bin/yaga-bhop.exe ./cmd/main.go
	mt.exe -manifest cmd/admin.manifest -outputresource:bin/yaga-bhop.exe;#1

clean:
	rm -rf bin/