#!/bin/sh

cd /usr/share/github/nh-cafe-bi/backend
docker build -t nhcafe-bo .

cd /usr/share/github/nh-cafe-bi
docker-compose -f docker-compose.prod.yml up -d
