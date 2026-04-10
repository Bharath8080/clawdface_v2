#!/bin/bash
cd /home/ec2-user
docker-compose down
docker rm $(docker ps --filter status=exited -q)
docker image rm 605982976738.dkr.ecr.us-east-2.amazonaws.com/trugen:clawdface-be -f
docker image rm 605982976738.dkr.ecr.us-east-2.amazonaws.com/trugen:webapp -f