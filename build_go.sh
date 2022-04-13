#!/bin/bash
# This script is to build the go modules from a compile server
# Mandatory input parameters are:
# Relative path to osfci github local fork (ex: ../osfci)
# Absolute path to an installation director (ex: /usr/local/production)
# Go must be installed

mypath=`realpath $1`
local=`realpath .`
#export GOPATH=$local:$mypath
#echo $GOPATH
if [ -f "go.mod" ] ; then
	echo "deleting go.mod in build dir\n"
        rm -f "go.mod"
fi
if [ -f "go.sum" ] ; then
	echo "deleting go.sum in build dir\n"
        rm -f "go.sum"
fi
cd $1/base/
pwd
go get -u go.uber.org/zap
go get gopkg.in/natefinch/lumberjack.v2
go mod init base/base
go mod tidy
cd /home/ciadmin/build/
pwd
go get golang.org/x/crypto/acme/autocert
go get golang.org/x/sys/unix
go get -v github.com/go-session/session
go mod init github.com/spf13
go mod edit -replace base/base=$1/base
go mod tidy
echo "Building Server.go ...\n"
go get base/base
go get github.com/spf13/viper
go get golang.org/x/crypto/acme/autocert
go get github.com/fsnotify/fsnotify@v1.4.9
go build -ldflags="-s" $1/gateway/server.go
echo "Building ctrl1.go ...\n"
go build -ldflags="-s" $1/ctrl/ctrl1.go
echo "Building user.go ...\n"
go build -ldflags="-s" $1/gateway/user.go
echo "Building storage.go ...\n"
go build -ldflags="-s" $1/gateway/backend/storage.go
tar cvf gateway.tar $1/gateway/html $1/gateway/css/ $1/gateway/images/ $1/gateway/js
go get github.com/docker/docker/api/types
go get github.com/docker/docker/client
echo "Building compile.go ...\n"
go build -ldflags="-s" $1/compile/compile.go
if [ -f "$1/base/go.mod" ] ; then
	echo "Deleting go.mod in base after build"
        rm -f "$1/base/go.mod"
fi
exit 1
