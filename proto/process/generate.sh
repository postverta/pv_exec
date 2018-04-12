#!/bin/sh
protoc --go_out=plugins=grpc:. process.proto
