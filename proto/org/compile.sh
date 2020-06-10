#!/bin/bash

protoc -I$GOPATH/src/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis \
	-I$GOPATH/src/github.com/grpc-ecosystem/grpc-gateway \
	-I$GOPATH/src/github.com/google/protobuf \
	-I$GOPATH/src/github.com/golang/protobuf \
	-I$GOPATH/src \
	-I. \
	--govalidators_out=. \
	--grpc-gateway_out=logtostderr=true:. \
	--go_out=plugins=grpc:. \
	--swagger_out=. *.proto
