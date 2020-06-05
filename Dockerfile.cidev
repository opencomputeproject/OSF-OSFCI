#set base image of ubuntu 18.04

FROM ubuntu:18.04

#author

MAINTAINER Gregg Shick

#metadata

LABEL version "1.0"
LABEL description="Docker container for development/test on CI system"


#install basic apps
RUN apt-get update
RUN apt install -qy apt-utils git
RUN apt install -qy wget ssh build-essential gdb libssl-dev libcurl4-gnutls-dev libexpat1-dev gettext unzip xvfb snapd squashfuse fuse snap-confine sudo fontconfig vim rand nano

#download and install go and set environment
RUN mkdir ~/go && mkdir ~/go/src
ENV GOROOT /usr/local/go
ENV GOPATH $HOME/go
ENV PATH $GOPATH/bin:$GOROOT/bin:$PATH
WORKDIR /tmp
RUN wget https://dl.google.com/go/go1.14.2.linux-amd64.tar.gz
RUN tar -C /usr/local -xvf go1.14.2.linux-amd64.tar.gz
RUN rm -f *.*
WORKDIR ../root
#Copy the osfci contents to the images
COPY . . 

#Execute start script for web server and proxy. 

RUN chmod u+x build.sh
ENTRYPOINT ["./build.sh"]
CMD ["/root"]



