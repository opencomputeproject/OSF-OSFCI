# osfci
Open Source Firmware Continuous Integration source code

# Disclaimer

This code is a proof of concept of what a continuous integration platform could look like in the near future with open source firmware and remote connectivity. It's intend is to support the development and validation of open source firmware stack on top of Proliant / Apollo server and demonstrate the technical behavior of the solution

# Who uses it ?

That proof of concept has been developped by the Advanced Technology team, and is currently running live on https://osfci.tech website as to validate its behavior.

# Running in standalone mode

As to debug and allow the platform to be developped without having access to real hardware, we are working on getting the system working in a standalone way by simulating command and output which might be originating from a real "working" machine. This is an ongoing effort, which is of high proriority currently. The build.sh script present into the root tree of the repo can be used to test such behavior. It must be executed outside the github repo as to avoid to polute it and ease your development. 

The proper way to execute it is by issuing the following command when you have created an out of tree build directory

edit the build.sh script first and adapt the environment variable to your environment

./build.sh <PATH to the OSFCI Tree>

# Authors

Jean-Marie Verdun

