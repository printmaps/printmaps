#!/bin/bash

# Kartenstil osm-carto-grey ableiten
# Autor   : Klaus Tockloth
# Version : 2.0.0 - 07.04.2017

# Debugging
# set +x
set -o xtrace

# Kartenstil ableiten
perl grayscale-css.pl mapnik.xml >mapnik-mono.xml

# PNG-Symbole umwandeln
cd symbols
mogrify -type Grayscale -alpha Activate *.png

# SVG-Shields umwandeln
cd shields
for filename in *.svg; do
    mv "$filename" "1-$filename"
    perl ../../grayscale-css.pl "1-$filename" >"$filename"
done
