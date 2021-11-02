# osfci
Open Source Firmware Continuous Integration source code

# Disclaimer

This code is a proof of concept of what a continuous integration platform could look like soon with open source firmware and remote connectivity. It aims to support the development and validation of open source firmware stack on top of Proliant / Apollo server and serves as a fully functioning technical demonstration.

# Who uses it ?

That proof of concept has been developed by the Advanced Technology team and is currently running on https://osfci.tech website as to validate its behaviour (08-17-21 update: the public facing server is being relocated).

# Running in standalone mode

This is an ongoing effort to debug and allow the platform to be developed without having access to real hardware. We are working on getting the system working in a standalone way by simulating I/O from a real "working" machine. The build scripts present into the root tree of the repo can be used to test such behaviour. They must be executed outside the github repo to avoid problems.

The proper way to execute is by issuing one of the following command(s) when you have created an out of tree build directory. If your requirement is building the whole infrastructure, choose build_all.sh . If it is just the go modules, use build_go.sh .

Edit the build_all.sh or build_go.sh scripts first and adapt the environment variable to your environment

./build_all.sh <PATH to the OSFCI Tree> 

OR

./build_go.sh <PATH to the OSFCI Tree>
