#!/bin/sh
protoc --go_out=plugins=grpc:. worktree.proto
