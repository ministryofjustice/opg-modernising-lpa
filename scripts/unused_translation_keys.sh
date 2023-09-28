#!/usr/bin/env sh

for k in `cat lang/en.json | jq -r 'keys|join("\n")'`;
do
    COUNT=`git grep $k | grep -v '.json:' | wc -l`
    if [ $COUNT = 0 ]; then
        echo $k
    fi
done
