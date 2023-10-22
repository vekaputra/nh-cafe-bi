#!/bin/sh

export NH_MYSQL_ROOT_PASSWORD=
export NH_MYSQL_USER=
export NH_MYSQL_PASSWORD=

cd /usr/share/github/nh-cafe-bi/backend
docker-compose up -d
