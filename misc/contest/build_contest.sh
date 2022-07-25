#!/bin/bash
# This script is to build the contestcli module
# Mandatory:
# Go must be installed

go get github.com/linuxboot/contest/cmds/clients/contestcli/cli
go mod init misc/contest
go mod tidy
go build -ldflags="-s" contestcli.go

if [ -f "go.mod" ] ; then
        echo "deleting go.mod in build dir\n"
        rm -f "go.mod"
fi
if [ -f "go.sum" ] ; then
        echo "deleting go.sum in build dir\n"
        rm -f "go.sum"
fi
