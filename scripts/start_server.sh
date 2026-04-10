#!/bin/bash
cd /home/ec2-user
cp docker-compose-prod.yml docker-compose.yml
docker-compose up -d