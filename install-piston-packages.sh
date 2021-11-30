#!/bin/bash

cd piston/cli
if [ -f "package.json" ]; then
    npm install
else
    echo "missing package.json"
    cd ../..
    exit 1
fi

BASE="node index.js -u http://localhost:2000"

LIST="$($BASE ppman list)"

LANGUAGES=($(echo "$LIST" | cut -d ' ' -f2))
VERSIONS=($(echo "$LIST" | cut -d ' ' -f3))
LENGTH=$(echo "$LIST" | wc -l)

for i in $(seq 1 $LENGTH); do
	$BASE ppman install ${LANGUAGES[$i]}=${VERSIONS[$i]}
done

#$BASE ppman list
cd ../..

