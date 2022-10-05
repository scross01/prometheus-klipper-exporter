VERSIONFILE=version.txt
VERSION=`cat $(VERSIONFILE)`

build:
	go build .

release: build-rpi build-linux build-macos build-windows

build-rpi:
	mkdir -p build/release-$(VERSION)
	env GOOS=linux GOARCH=arm GOARM=7 go build -o build/release-$(VERSION)/prometheus-klipper-exporter-rpi-armv7-$(VERSION) .
	env GOOS=linux GOARCH=arm64 go build -o build/release-$(VERSION)/prometheus-klipper-exporter-rpi-arm64-$(VERSION) .

build-linux:
	mkdir -p build/release-$(VERSION)
	env GOOS=linux GOARCH=amd64 go build -o build/release-$(VERSION)/prometheus-klipper-exporter-linux-amd64-$(VERSION) .

build-macos:
	mkdir -p build/release-$(VERSION)
	env GOOS=darwin GOARCH=amd64 go build -o build/release-$(VERSION)/prometheus-klipper-exporter-macos-amd64-$(VERSION) .
	env GOOS=darwin GOARCH=arm64 go build -o build/release-$(VERSION)/prometheus-klipper-exporter-macos-arm64-$(VERSION) .

build-windows:
	mkdir -p build/release-$(VERSION)
	env GOOS=windows GOARCH=amd64 go build -o build/release-$(VERSION)/prometheus-klipper-exporter-windows-amd64-$(VERSION).exe .

build-docker:
	docker build -t klipper-exporter .

clean:
	rm -rf build/release-$(VERSION)/*

fmt:
	(cd collector && go fmt)
	go fmt

run:
	go run .	

.PHONY: build