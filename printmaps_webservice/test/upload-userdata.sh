#!/bin/bash
#
# upload user data file (e.g. gpx file with tracks)

set -o verbose

curl \
--silent \
--include \
--header "Accept: application/vnd.api+json; charset=utf-8" \
--request POST \
--form "file=@aasee.gpx" \
http://printmaps-osm.de:8282/api/beta/maps/upload/0ac04905-7c27-40cb-a667-e0f9dae61bd3

