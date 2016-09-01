#!/bin/sh

GIT_VERSION=`git rev-parse --short HEAD || echo "GitNotFound"`
APP_VERSION=1.0
go build -ldflags  "-X main.buildTime=`date  +%Y%m%d-%H%M%S` -X main.binaryVersion=$APP_VERSION -X main.gitRevision=${GIT_VERSION}"
