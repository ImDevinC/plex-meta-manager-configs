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
envsubst < /source/config/config.yaml > /config/config.yaml
/tini -s python3 plex_meta_manager.py -- --config=/config/config.yaml -ro --run
if [ $? -gt 0 ]; then
    echo "failed to run pmm"
    exit 1
fi
/config-diff -source /source/movies-backup.yml
if [ $? -gt 0 ]; then
    echo "failed to diff configs"
    exit 1
fi