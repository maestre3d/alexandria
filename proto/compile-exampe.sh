#!/usr/bin/env sh

# Install proto3 from source
#  brew install autoconf automake libtool
#  git clone https://github.com/google/protobuf
#  ./autogen.sh ; ./configure ; make ; make install
#
# Update protoc Go bindings via
#  go get -u github.com/golang/protobuf/{proto,protoc-gen-go}
#
# See also
#  https://github.com/grpc/grpc-go/tree/master/examples

# This is just an example to compile the proto lib to any
# programming language required by alexandria
# This is meant to be ran inside project root folder

# Golang compile
# Requires GOROOT, GOPATH and  GOBIN env variables
protoc -I proto/ proto/alexandria.proto --go_out=plugins=grpc:foo-service/pb/
