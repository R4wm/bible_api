#!/bin/bash

if [ $UID -ne 0 ]
then
    echo run as root
    exit 1
fi

script_path=$(dirname $(dirname $(realpath "$0")))

echo ">>> making dir /etc/bible_api/data"
mkdir -p /etc/bible_api/data
cp "$script_path/data/kjv.db" /etc/bible_api/data/kjv.db
echo ">>> building bible_api"
CGO_ENABLED=1 /usr/local/go/bin/go build -o /tmp/bible_api ../cmd/bible_api.go

echo ">>> installing systemctl service file"


echo ">>> installing bible_api to /usr/bin"
sudo cp /tmp/bible_api /usr/bin/

echo OK
