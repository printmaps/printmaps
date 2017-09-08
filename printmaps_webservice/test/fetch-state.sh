#!/bin/bash
#
# fetch map state

set -o verbose

curl \
--silent \
--include \
--header "Accept: application/vnd.api+json; charset=utf-8" \
http://printmaps-osm.de:8282/api/beta/maps/mapstate/0ac04905-7c27-40cb-a667-e0f9dae61bd3

