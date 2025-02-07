.PHONY: build clean

build:
	go generate ./...
	go build -ldflags="-H windowsgui" -o bin/bhop.exe ./cmd/
	mt.exe -manifest admin.manifest -outputresource:bin/bhop.exe;#1

clean:
	rm -rf bin/