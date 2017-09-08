#!/bin/bash
#
# update map meta data

postdata=$(cat <<EOF
{
    "Data": {
        "Type": "maps",
        "ID": "0ac04905-7c27-40cb-a667-e0f9dae61bd3",
        "Attributes": {
            "Fileformat": "png",
            "Scale": 10000,
            "PrintWidth": 600,
            "PrintHeight": 600,
            "Latitude": 51.9505,
            "Longitude": 7.6049,
            "Style": "osm-carto",
            "HideLayers": "admin-low-zoom,admin-mid-zoom,admin-high-zoom,admin-text",
            "UserData": [
                {
                    "Style": "<LineSymbolizer stroke='crimson' stroke-width='10' stroke-opacity='0.75' stroke-linecap='round' />",
                    "SRS": "+init=epsg:4326",
                    "Type": "ogr",
                    "File": "aasee.gpx",
                    "Layer": "tracks"
                }
            ],
            "UserItems": [
                {
                    "Style": "<PolygonSymbolizer fill='white' fill-opacity='0.75' />",
                    "WellKnownText": "POLYGON((0.0 0.0, 0.0 600.0, 600.0 600.0, 600.0 0.0, 0.0 0.0), (15.0 15.0, 15.0 585.0, 585.0 585.0, 585.0 15.0, 15.0 15.0))"
                },
                {
                    "Style": "<LineSymbolizer stroke='black' stroke-width='3' stroke-linecap='square' />",
                    "WellKnownText": "LINESTRING(15.0 15.0, 15.0 585.0, 585.0 585.0, 585.0 15.0, 15.0 15.0)"
                },
                {
                    "Style": "<TextSymbolizer fontset-name='fontset-2' size='40' fill='firebrick' allow-overlap='true'>'Aasee, Münster\\nTour and Sculptures\\nScale 1:10000'</TextSymbolizer>",
                    "WellKnownText": "POINT(300.0 560.0)"
                },
                {
                    "Style": "<TextSymbolizer fontset-name='fontset-0' size='12' fill='firebrick' orientation='90' halo-radius='1' halo-fill='white' allow-overlap='true'>'© OpenStreetMap contributors'</TextSymbolizer>",
                    "WellKnownText": "POINT(7.5 300.0)"
                },
                {
                    "Style": "<TextSymbolizer fontset-name='fontset-2' size='360' fill='firebrick' opacity='0.2' orientation='315' allow-overlap='true'>'S A M P L E'</TextSymbolizer>",
                    "WellKnownText": "POINT(300.0 300.0)"
                },
                {
                    "Style": "<TextSymbolizer fontset-name='fontset-1' size='16' fill='firebrick' halo-radius='1' halo-fill='white' allow-overlap='true'>'500 Meter'</TextSymbolizer>",
                    "WellKnownText": "POINT(40.0 36.0)"
                }
            ],
            "UserScalebar": {
                "Style": "<LineSymbolizer stroke='firebrick' stroke-width='8' stroke-linecap='butt' />",
                "NatureLength": 500,
                "XPos": 30,
                "YPos": 30
            }
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
--request PATCH \
http://printmaps-osm.de:8282/api/beta/maps/metadata

