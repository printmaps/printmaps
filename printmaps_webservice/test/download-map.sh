#!/bin/bash
#
# download map (zip)

set -o verbose  #echo on

curl \
--header "Accept: application/vnd.api+json; charset=utf-8" \
--dump-header "response-header.txt" \
--output "printmap.zip" \
http://printmaps-osm.de:8282/api/beta/maps/mapfile/0ac04905-7c27-40cb-a667-e0f9dae61bd3

set +o verbose # echo off from bash script
echo

cat 'response-header.txt'

found=`grep "HTTP/1.1 200 OK" < "response-header.txt"`
if  [ ! "$found" ]; then
    cat "printmap.zip"
    echo
fi
