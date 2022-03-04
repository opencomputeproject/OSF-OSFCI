# Introduction

The osfci/api directory contains bash shell script which could be used as a demo of the current CLI API available to OSFCI. The API is into an unstable state, as it is currently under active development. The current workflow is based on opening a session to the service, execute related commands which could be

-   Build a firmware from a github repo (either OpenBMC or Linuxboot)
-   Retrieve build logs
-   Load firmware on a target system
-   Control power operation of the target system
-   Remote login and locally execute command
-   Test automation through  [contest](https://github.com/linuxboot/contest)  framework

and closing the session. Each session has an expiring time which is currently set to 30 minutes.

The following steps are implemented

-   Session initialization
-   Build a firmware from a github repo (either OpenBMC or Linuxboot)
-   Session termination

We are focused on log retrieval process which requires some backend modification to store them at the proper place.  
The remote login requires some modification into linuxboot and openbmc source trees as to inject an ssh key within built images. openbmc patches have been upstreamed. An init script needs to be written for linuxboot, but a proof of concept demoed at latest OSF conference could be reused.  
The control power calls are easy to implement, as they do exist for the interactive session.

All API calls are performed through HTTPS, the API is not available without proper encryption.

# API calls

### Work in progress please comment

Current API calls include (please propose alternate naming, or convention etc ...):

### /user/$username/get_token

_Method_  **POST**  
_Parameters_: password="string"  
_Signed_: No  
_Return values:_  { "accessKey" : "_string_", "secretKey" : "_string_" }, plus an authentication cookie

_Description_: If successful this call is allocating a test node within the CI infrastructure to the username specified through $username parameter. The returned cookie must be pass to any upcoming API calls. The accessKey and secretKey parameters are used to sign specific API call

### /ci/get_server

_Method_:  **GET**  
_Parameters_:  
_Signed_: Yes  
_Return values:_  {"Servername":"_string_","Waittime":"_integer_","Queue":"_integer_","RemainingTime":"_integer_"}

_Description_: This call is retrieving an available server from the CI and allocate it to the username specified by the connection cookie. It is secured through AWS SHA1 header signature process. When successful, the RemainingTime value is higher than 0. A servername is provided. If no server is available a Waittime in second is provided as well as a queue position. Waittime and queue are estimated numbers, and it is up to the end user to re-initialize an API call to ask for a new session. When called multiple times while a server is soon allocated to an end user, the API entry could be used to know session RemainingTime.

### /ci/stop_server/$serverName

_Method_:  **GET**  
_Parameters_:  
_Signed_: No  
_Return values:_

_Description_: This call is closing a session by releasing the test server and cleaning the related infrastructure

### /ci/build_bmc_firmware/$username

_Method_:  **PUT**  
_Parameters_: "$git $branch $machine 0"  
_Signed_: Yes  
_Return values:_

_Description_: This call is launching a background OpenBMC build from the specified github url ($git), branch ($branch) and machine type (currently  **must be set to dl360poc**). The last parameter must be 0. It does specify that the session is non interactive. Only direct repository fork with patches from  [OpenBMC](https://github.com/openbmc/openbmc)  are supported

### /ci/build_bios_firmware/$username

_Method_:  **PUT**  
_Parameters_: "$git $branch $machine 0"  
_Signed_: Yes  
_Return values:_

_Description_: This call is launching a background linuxboot build from the specified github url ($git), branch ($branch) and machine type (currently  **must be set to hpe/dl360gen10"**). The last parameter must be 0. It does specify that the session is non interactive. Only direct repository fork with patches from  [linuxboot mainboards](https://github.com/linuxboot/mainboards)  are supported

### /ci/is_running/openbmc

_Method_:  **GET**  
_Parameters_:  
_Signed_: No  
_Return values:_  {"status" : "integer"}

_Description_: This call is returning 0 if a background build of openbmc is finished. It returns 1 if it is still running.

### /ci/is_running/linuxboot

_Method_:  **GET**  
_Parameters_:  
_Signed_: No  
_Return values:_  {"status":"integer"}

_Description_: This call is returning 0 if a background build of linuxboot is finished. It returns 1 if it is still running.

### /user/$username/getLinuxBootLog

_Method_:  **GET**  
_Parameters_:  
_Signed_: Yes  
_Return values:_  None

_Description_: This call is retrieving the latest linuxboot built log from the CI. Request must be sign to confirm identity, and a valid cookie must be passed.

### /user/$username/getOpenBMCLog

_Method_:  **GET**  
_Parameters_:  
_Signed_: Yes  
_Return values:_  None  
_Description_: This call is retrieving the latest OpenBMC built log from the CI. Request must be sign to confirm identity, and a valid cookie must be passed.

### /user/$username/getLinuxBoot

_Method_:  **GET**  
_Parameters_:  
_Signed_: Yes  
_Return values:_  None

_Description_: This call is retrieving the latest linuxboot image built from the CI from $username user. Request must be sign to confirm identity

### /user/$username/getOpenBMC

_Method_:  **GET**  
_Parameters_:  
_Signed_: Yes  
_Return values:_  None  
_Description_: This call is retrieving the latest OpenBMC image built from the CI from $username user. Request must be sign to confirm identity, and a valid cookie must be passed.

