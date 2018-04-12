#!/bin/sh
protoc --go_out=plugins=grpc:. exec.proto
