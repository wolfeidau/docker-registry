FROM ubuntu:trusty

run apt-get install -y curl build-essential git-core golang

# Install Go (this is copied from the docker Dockerfile)
env GOPATH  /go
env CGO_ENABLED 0

run git clone https://github.com/wolfeidau/docker-registry.git /docker-registry.git
run cd /docker-registry.git && make && cp bin/docker-registry /usr/local/bin/

expose 80
cmd /usr/local/bin/docker-registry
