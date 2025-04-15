#!/bin/bash

set -e

sudo docker stop $(sudo docker ps -q) || true && sudo docker rm -vf $(sudo docker ps -aq) || true && sudo docker rmi -f $(sudo docker images -q) || true && sudo docker volume prune -f || true && sudo docker network prune -f || true && sudo docker builder prune -a -f || true && sudo docker system prune -a --volumes -f || true

# Build the Node image
sudo docker build --target=node -t ixios-spark-node .
sudo docker save ixios-spark-node > ixiosSpark-docker-node.tar

# Build the Validator image
sudo docker build --target=validator -t ixios-spark-validator .
sudo docker save ixios-spark-validator > ixiosSpark-docker-validator.tar