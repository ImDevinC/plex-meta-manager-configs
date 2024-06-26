#!/bin/bash

git clone https://github.com/ImDevinC/plex-meta-manager-configs /source
if [ $? -gt 0 ]; then
    echo "failed to clone source repo"
    exit 1
fi

cp /source/config/movies.yml /source/movies-backup.yml
if [ $? -gt 0 ]; then
    echo "failed to create backup config"
    exit 1
fi

/tini -s python3 kometa.py -- --config=/config/config.yaml -ro --run
if [ $? -gt 0 ]; then
    echo "failed to run pmm"
    exit 1
fi
lines=$(wc -l /source/movies-backup.yml)
if [ $? -gt 0 ]; then
   echo "missing backup file"
   exit 1
fi
/config-diff -source /source/movies-backup.yml
if [ $? -gt 0 ]; then
    echo "failed to diff configs"
    exit 1
fi
