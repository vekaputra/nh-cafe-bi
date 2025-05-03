#!/bin/sh

cd /home/ubuntu/github/nh-cafe-bi/backend
docker build -t nhcafe-bo .

cd /home/ubuntu/github/nh-cafe-bi
docker-compose -f docker-compose.prod.yml up -d
