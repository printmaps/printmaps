#!/bin/bash
#
# order map (request map build)

postdata=$(cat <<EOF
{
    "data": {
        "type": "maps",
        "id": "0ac04905-7c27-40cb-a667-e0f9dae61bd3"
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
http://printmaps-osm.de:8282/api/beta/maps/mapfile

