#!/bin/bash
git stash
git pull
docker stop flask_python_task
docker rm flask_python_task
docker rmi flask_python_task:1.0
docker build -t flask_python_task:1.0 .
docker run -d -p 27990:8090 -v /etc/localtime:/etc/localtime --name=flask_python_task --restart=always flask_python_task:1.0
docker exec -it flask_python_task bash /app/run.sh