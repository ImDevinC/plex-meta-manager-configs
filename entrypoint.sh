#!/bin/bash

git clone https://github.com/ImDevinC/plex-meta-manager-configs /source
cp /source/movies.yml /source/movies-backup.yml
/tini -s python3 plex_meta_manager.py -- --config=/source/config/config.yaml -ro --run
/config-diff -source /source/movies-backup.yml