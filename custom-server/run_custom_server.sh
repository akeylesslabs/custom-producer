#!/bin/bash

DOCKER_NAME=custom-server
IMAGE_NAME=akeyless/custom-server

GW_ACCESS_ID="" # Your Gateway's Admin Access ID

docker rm -f $DOCKER_NAME
docker pull $IMAGE_NAME
docker run -d -p 2608:2608 -v $PWD/custom_logic.sh:/custom_logic.sh \
	   -e GW_ACCESS_ID=${GW_ACCESS_ID}                          \
	   --restart unless-stopped --name $DOCKER_NAME $IMAGE_NAME

