#!/bin/bash
if [ $# -eq 0 ]; then
	echo "Usage: ./compiler_container.sh [code path] [compiler count]" 
	echo "Usage: ./compiler_container.sh /usr/local/production 3" 
	exit 1
fi
if [ ! -d $1/bin ]
then
	echo "Unable to find $1/bin directory. Exiting..."
	exit 2
fi
if [ ! -d $1/logs ]
then
	mkdir $1/logs
fi
if [ ! -d /tmp/volume ]
then
	mkdir /tmp/volume
	chmod 777 /tmp/volume
fi
if [ ! -d /datas ]
then
	sudo mkdir /datas
fi
sudo chown ciadmin:ciadmin /datas
# Creating log directory for compiler containers. 
count=$2
while [ $count -gt 0 ];

do
	if [ ! -d $1/logs/compiler$count ]
	then
		echo "$1/logs/compiler$count doesn't exists"
		mkdir -p $1/logs/compiler$count
	else
		echo "$1/logs/compiler$count exists"
		
	fi
	((count--))
done
cp Dockerfile.compiler $1/docker
# We need to take care of potential PROXY and put them into 
# the Docker files
if [ "$http_proxy" != "" ]
then
cat $1/docker/Dockerfile.compiler | sed 's,# lets build,RUN echo \"Acquire::http::Proxy \\\"'"$http_proxy"'\\\"\;" > \/etc\/apt\/apt.conf.d\/proxy.conf\nENV http_proxy '"$http_proxy"'\n# https\n,' > $1/docker/Docker.temp
cp $1/docker/Docker.temp $1/docker/Dockerfile.compiler
rm $1/docker/Docker.temp
fi
if [ "$https_proxy" != "" ]
then
cat $1/docker/Dockerfile.compiler | sed "s,# https,ENV https_proxy $https_proxy\n," > $1/docker/Docker.temp
cp $1/docker/Docker.temp $1/docker/Dockerfile.compiler
rm $1/docker/Docker.temp
fi
# Remove existing containers and image
if [ "$(docker ps -qa -f ancestor=compilernode)" ]; then
        if [ "$(docker ps -qa -f ancestor=compilernode -f status=running)" ]; then
                echo $(docker ps -qa -f ancestor=compilernode -f status=running)
                docker stop $(docker ps -qa -f ancestor=compilernode)
        fi
        docker rm -f $(docker ps -qa -f ancestor=compilernode)
fi
if [ "$(docker image inspect compilernode --format=\"ignore\")" ]; then
	docker image rm compilernode
fi
docker build -t compilernode -f $1/docker/Dockerfile.compiler $1
