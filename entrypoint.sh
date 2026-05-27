#!/bin/bash

SERVER_TYPE=${SERVER_TYPE:-github}

if [ "$SERVER_TYPE" = "forgejo" ]; then
    if [ -z "$FORGEJO_URL" ] || [ -z "$FORGEJO_OWNER" ] || [ -z "$FORGEJO_REPO" ]; then
        echo "missing required forgejo environment variables (FORGEJO_URL, FORGEJO_OWNER, FORGEJO_REPO)"
        exit 1
    fi
    git clone "${FORGEJO_URL}/${FORGEJO_OWNER}/${FORGEJO_REPO}" /source
else
    if [ -z "$GITHUB_OWNER" ] || [ -z "$GITHUB_REPO" ]; then
        echo "missing required github environment variables (GITHUB_OWNER, GITHUB_REPO)"
        exit 1
    fi
    git clone "https://github.com/${GITHUB_OWNER}/${GITHUB_REPO}" /source
fi

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
cp /source/movies-backup.yml /config/movies-backup-backup.yml
