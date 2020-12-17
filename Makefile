
OS_ARCH=darwin_amd64
#OS_ARCH=linux_amd64
PATH_DIR=/usr/local/bin

default: build

build:
	# MACOS Image
	GOOS=darwin GOARCH=amd64 go build -o bin/rezolvr_darwin_amd64
	GOOS=darwin GOARCH=amd64 go build -o ./bin/plugindocker_darwin_amd64.so -buildmode=plugin plugins/docker/plugindocker.go
	GOOS=darwin GOARCH=amd64 go build -o ./bin/pluginkube_darwin_amd64.so -buildmode=plugin plugins/kube/pluginkube.go
	# Build a Linux image too
	GOOS=linux GOARCH=amd64 go build -o bin/rezolvr_linux_amd64
	#GOOS=linux GOARCH=amd64 go build -o ./bin/plugindocker_linux_amd64.so -buildmode=plugin plugins/docker/plugindocker.go
	#GOOS=linux GOARCH=amd64 go build -o ./bin/pluginkube_linux_amd64.so -buildmode=plugin plugins/kube/pluginkube.go
	# Copy templates
	mkdir -p ./bin/plugins/docker/templates
	mkdir -p ./bin/plugins/kube/templates
	cp ./plugins/docker/*.template ./bin/plugins/docker/templates
	cp ./plugins/kube/*.template ./bin/plugins/kube/templates

test:
	go test -v ./...

install: build
	mkdir -p ~/.rezolvr/plugins/docker/templates
	mkdir -p ~/.rezolvr/plugins/kube/templates
	cp ./bin/rezolvr_${OS_ARCH} ${PATH_DIR}/rezolvr
	cp ./bin/plugindocker_${OS_ARCH}.so ~/.rezolvr/plugins/docker/plugindocker.so
	cp ./bin/pluginkube_${OS_ARCH}.so ~/.rezolvr/plugins/kube/pluginkube.so
	cp ./plugins/docker/*.template ~/.rezolvr/plugins/docker/templates
	cp ./plugins/kube/*.template ~/.rezolvr/plugins/kube/templates
