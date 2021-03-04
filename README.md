# osfci
Open Source Firmware Continuous Integration source code

# Disclaimer

This code is a proof of concept of what a continuous integration platform could look like soon with open source firmware and remote connectivity. It aims to support the development and validation of open source firmware stack on top of Proliant / Apollo server and serves as a fully functioning technical demonstration.

# Who uses it ?

That proof of concept has been developed by the Advanced Technology team and is currently running live on https://osfci.tech website as to validate its behaviour.

# Running in standalone mode

This is an ongoing effort to debug and allow the platform to be developed without having access to real hardware. We are working on getting the system working in a standalone way by simulating I/O from a real "working" machine. The build.sh script present into the root tree of the repo can be used to test such behaviour. It must be executed outside the github repo to avoid problems.

The proper way to execute it is by issuing the following command when you have created an out of tree build directory

edit the build.sh script first and adapt the environment variable to your environment

./build.sh <PATH to the OSFCI Tree>

# Authors

Jean-Marie Verdun

