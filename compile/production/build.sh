#!/bin/bash
# This script is building the whole infrastructure from a compile server
# Mandatory input parameters are:
# Relative path to osfci github local fork (ex: ../osfci)
# Absolute path to an installation director (ex: /usr/local/production)
# Go must be installed

mypath=`realpath $1`
local=`realpath .`
export GOPATH=$local:$mypath
echo $GOPATH
go get golang.org/x/crypto/acme/autocert
go get golang.org/x/sys/unix
go get -v github.com/go-session/session
go build $1/gateway/server.go
go build $1/ctrl/ctrl1.go
go build $1/gateway/user.go
go build $1/gateway/backend/storage.go
tar cvf gateway.tar $1/gateway/html $1/gateway/css/ $1/gateway/images/ $1/gateway/js
\rm -rf tmp
\rm -rf /usr/local/old/*
mkdir tmp
cd tmp
# let's build em100
git clone http://review.coreboot.org/em100
sudo apt -y install libusb-1.0-0-dev libusb-dev libcurl4-openssl-dev
cd em100
make
cd ..
# We ned to build the acroname
cp -Rf $mypath/ctrl/iUSB .
pwd
cd iUSB
#let's build uhubctl first
cd uhubctl
./build.sh
cd ..
if [ ! -d compileiUSB ]
then
	# We must install the development environment
mkdir compileiUSB
cd compileiUSB
wget https://acroname.com/system/files/software/brainstem_dev_kit_ubuntu_lts_18.04_x86_64_7.tgz
gunzip brainstem_dev_kit_ubuntu_lts_18.04_x86_64_7.tgz
tar xf brainstem_dev_kit_ubuntu_lts_18.04_x86_64_7.tar
export ACROSDK=`realpath .`
cd development/reflex_examples
cp ../../bin/reflex/* .
cp -r ../../bin/aInclude ..
cp -rf ../../development/lib/* .
cp $mypath/ctrl/iUSB/Acroname/Swap_Ports.reflex .
cp $mypath/ctrl/iUSB/Acroname/switch.cpp .
./arc Swap_Ports.reflex
g++ -o switch switch.cpp -I$ACROSDK/development/lib $ACROSDK/development/lib/libBrainStem2.a -ludev -pthread
cd $ACROSDK
fi
cd ../../..
pwd
cp -Rf $mypath/ctrl/iPDU tmp
cd tmp
cd iPDU
cd HPEiPDU
make all
tar cvf HPEiPDU.tar /usr/local/old
cd ../..
cp -Rf $mypath/ctrl/ttyd .
cd ttyd
gunzip ttyd.tar.gz
tar xf ttyd.tar
cd ttyd
mkdir build
cd build
cmake ..
make
cd ../../../..
mkdir ctrl
mkdir ctrl/bin
cp $1/ctrl/* ctrl/bin
rm ctrl/bin/ctrl1.go
cp tmp/iUSB/uhubctl/uhubctl/uhubctl ctrl/bin
cp tmp/iUSB/compileiUSB/development/reflex_examples/switch ctrl/bin
cp tmp/iUSB/compileiUSB/bin/Updater ctrl/bin
cp tmp/iUSB/compileiUSB/development/reflex_examples/Swap_Ports.map ctrl/bin
# cp -Rf tmp/iUSB/compileiUSB/development/bin/* ctrl/bin
cp tmp/em100/em100 ctrl/bin
cp tmp/iPDU/HPEiPDU/iPDU_HPE ctrl/bin
cp tmp/ttyd/ttyd/build/ttyd ctrl/bin
cp tmp/iPDU/HPEiPDU/HPEiPDU.tar ctrl
cp ctrl1 ctrl/bin/ctrl1
tar cvf ctrl1.tar ctrl
cd $1/compile
go build shadow.go
cp shadow /usr/local/production/bin
chmod 755 ./deploy
\rm -rf tmp
\rm -rf /tmp/ttyd
./deploy $2
chmod 755 $2/bin/readBiosFifo
