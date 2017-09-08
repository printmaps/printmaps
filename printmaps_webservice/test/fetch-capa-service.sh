#!/bin/bash
#
# fetch service capabilities

set -o verbose

curl \
--silent \
--include \
--header "Accept: application/vnd.api+json; charset=utf-8" \
http://printmaps-osm.de:8282/api/beta/maps/capabilities/service

