FROM ubuntu:trusty

run apt-get update && apt-get install -y curl build-essential git-core golang

# Install Go (this is copied from the docker Dockerfile)
env GOPATH  /go
env CGO_ENABLED 0

run go get github.com/wolfeidau/docker-registry

expose 5000
cmd /go/bin/docker-registry
