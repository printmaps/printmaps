#!/bin/bash
#
# delete map (all artifacts)

set -o verbose

curl \
--silent \
--include \
--header "Accept: application/vnd.api+json; charset=utf-8" \
--request DELETE \
http://printmaps-osm.de:8282/api/beta/maps/d0273d49-43db-4447-b9a8-259a1cf1b211

