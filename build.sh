#!/bin/bash
if [ ! "$DNSDOMAIN" ] 
then
if [ ! -f "certificat.crt" ] 
then
openssl genrsa -out certificat.key 4096
openssl rsa -in certificat.key -out certificat.key.unlock
openssl req -new -key certificat.key -out certificat.csr -subj "/C=US/ST=NRW/L=Houston/O=Jon Doe/OU=DevOps/CN=www.example.com/emailAddress=dev@www.example.com"
openssl x509 -req -days 365 -in certificat.csr -signkey certificat.key -out certificat.crt
fi
fi
mypath=`realpath $1`
local=`realpath .`
export GOPATH=$local:$mypath
echo $GOPATH
go get golang.org/x/crypto/acme/autocert
go get -v github.com/go-session/session
go build $1/gateway/server.go
go build $1/ctrl/ctrl1.go
go build $1/gateway/user.go
go build $1/gateway/backend/storage.go

export TLS_KEY_PATH=./certificat.key.unlock
export TLS_CERT_PATH=./certificat.crt
export STATIC_ASSETS_DIR=$1/gateway/
export CREDENTIALS_TCPPORT=:9100
export CREDENTIALS_URI=127.0.0.1
\mkdir cert
export CERT_STORAGE=./cert

export SMTP_SERVER=
export SMTP_ACCOUNT=
export SMTP_PASSWORD=

export STORAGE_ROOT=./backend
export STORAGE_TCPPORT=:9200
export STATIC_ASSETS_DIR=../osfci/gateway/
export STORAGE_URI=127.0.0.1

./user &
./storage &
./server



