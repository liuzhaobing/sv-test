#!/bin/bash
git stash
git pull
docker stop interface_python
docker rm interface_python
docker rmi interface_python:1.0
docker build -t interface_python:1.0 .
docker run -d -p 27991:8090 -v /etc/localtime:/etc/localtime --name=interface_python --restart=always interface_python:1.0