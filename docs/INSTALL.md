# Introduction

This resource will feature playbooks and installation steps for the CI.

### Running in standalone mode

This is an ongoing effort to debug and allow the platform to be developed without having access to real hardware. We are working on getting the system working in a standalone way by simulating I/O from a real "working" machine. The build scripts present into the root tree of the repo can be used to test such behaviour. They must be executed outside the github repo to avoid problems.

The proper way to execute is by issuing one of the following command(s) when you have created an out of tree build directory. If your requirement is building the whole infrastructure, choose build_all.sh . If it is just the go modules, use build_go.sh .

Edit the build_all.sh or build_go.sh scripts first and adapt the environment variable to your environment

./build_all.sh <PATH to the OSFCI Tree> 

OR

./build_go.sh <PATH to the OSFCI Tree>
