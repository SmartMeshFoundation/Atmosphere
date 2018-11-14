#!/bin/sh

export GIT_COMMIT=`git rev-list -1 HEAD`
export GO_VERSION=`go version|sed 's/ //g'`
export BUILD_DATE=`date|sed 's/ //g'`
export VERSION=0.91
echo $GIT_COMMIT

go  build -ldflags " -X github.com/SmartMeshFoundation/Atmosphere/cmd/atmosphere/mainimpl.GitCommit=$GIT_COMMIT -X github.com/SmartMeshFoundation/Atmosphere/cmd/atmosphere/mainimpl.GoVersion=$GO_VERSION -X github.com/SmartMeshFoundation/Atmosphere/cmd/atmosphere/mainimpl.BuildDate=$BUILD_DATE -X github.com/SmartMeshFoundation/Atmosphere/cmd/atmosphere/mainimpl.Version=$VERSION "

cp atmosphere $GOPATH/bin