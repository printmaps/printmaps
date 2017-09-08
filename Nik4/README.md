# Nik4

Die eigenliche Kartenerzeugung erfolgt über das Python-Skript:
* nik4-printmaps.py

"nik4-printmaps.py" wird intern vom Buildservice "printmaps_buildservice" aufgerufen.

"nik4-printmaps.py" von "nik4.py" (https://github.com/Zverik/Nik4) abgeleitet
und an die Erfordernisse von Printmaps angepaßt.

"nik4-printmaps.py" kann auch unabhängig, zum Beispiel zur Erzeugung von (Test-) Karten, genutzt werden.
Darüberhinaus bietet es sich zur Überprüfung einer neu durchgeführten mapnik-Installation an.

    $ python ./Nik4/nik4-printmaps.py -h
    usage: nik4-printmaps.py [-h] [--version] --ppi PPI --scale SCALE --size W H
                             --center X Y [--add-layers ADD_LAYERS]
                             [--hide-layers HIDE_LAYERS] [--debug] [--info]
                             style output
    
    Nik4 1.6 (Printmaps): mapnik image renderer
    
    positional arguments:
      style                 Style file for mapnik
      output                Resulting image file
    
    optional arguments:
      -h, --help            show this help message and exit
      --version             show program's version number and exit
      --ppi PPI             Pixels per inch (alternative to scale)
      --scale SCALE         Scale as in 1:1000 (set ppi is recommended)
      --size W H            Target dimensions in mm
      --center X Y          Center of an image
      --add-layers ADD_LAYERS
                            Map layers to include, comma-separated
      --hide-layers HIDE_LAYERS
                            Map layers to hide, comma-separated
      --debug               Display calculated values
      --info                Quit after displaying calculated values

Beispiele:

    python ./Nik4/nik4-printmaps.py --scale 10000 --size 400 400 --ppi 300 --center 7.6279 51.9506 ./printstyles/osm-carto-4.2.0/mapnik.xml muenster-osm-carto.png

    python ./Nik4/nik4-printmaps.py --scale 2500 --size 400 400 --ppi 300 --center 7.6279 51.9506 ./printstyles/schwarzplan+-1.2.0/mapnik.xml muenster-schwarzplan+.png
