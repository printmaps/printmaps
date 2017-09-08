#!/bin/bash
#
# create map meta data (new map) 

postdata=$(cat <<EOF
{
    "data": {
        "type": "maps",
        "attributes": {
            "fileformat": "jpeg",
            "scale": 5000,
            "printWidth": 6600.0,
            "printHeight": 600.0,
            "latitude": 151.9505,
            "longitude": 7.6049,
            "style": "osm-cartago"
        }
    }
}
EOF
)

echo "postdata =\n$postdata"

set -o verbose

curl \
--silent \
--include \
--header "Content-Type: application/vnd.api+json; charset=utf-8" \
--header "Accept: application/vnd.api+json; charset=utf-8" \
--data "$postdata" \
http://printmaps-osm.de:8282/api/beta/maps/metadata

