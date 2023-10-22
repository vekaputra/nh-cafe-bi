#!/bin/sh

if [ ! -d "/usr/share/github" ]; then
    mkdir -p /usr/share/github
fi

cd /usr/share/github
git clone git@github.com:vekaputra/nh-cafe-bi.git

# cd /usr/share/github/nh-cafe-bi/backend
# docker build -t nhcafe-bo .

if [ ! -d "/usr/share/github/nh-cafe-bi/backend/data" ]; then
    mkdir -p /usr/share/github/nh-cafe-bi/backend/data
fi
