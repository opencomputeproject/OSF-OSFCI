#include <iostream>
#include "BrainStem2/BrainStem-all.h"

int main(int argc, const char * argv[]) {

    // Create an instance of the USBHub2x4
    aUSBHub2x4 hub;
    aErr err = aErrNone;

    err = hub.discoverAndConnect(USB);
    if (err != aErrNone) {
        std::cout << "Error "<< err <<" Module not found or didn't answered " << std::endl;
        return 1;

    } 
    // We just switch from to Port 1 as host
    err = hub.usb.setUpstreamMode(1);
    // Disconnect
    err = hub.disconnect();
    if (err == aErrNone) {
        std::cout << "Disconnected from BrainStem module." << std::endl;
    }
    return 0;
}

