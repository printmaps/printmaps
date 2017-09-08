# Kartenstile

Standardmäßig stehen folgende Kartenstile zur Verfügung:
* osm-carto
* raster10
* schwarzplan
* schwarzplan+

Der wichtigste Kartenstil ist sicherlich "osm-carto", das offizielle Design der internationen OSM-Webseite.
In regelmäßigen Abständen wird der Kartenstil aktualisiert. Aus den Daten im entsprechenden Github-Repository
ist für Printmaps die XML-Datei "mapnik.xml" zu erzeugen.

## carto-Compiler installieren

    sudo npm install -g carto

## carto-Version prüfen

    carto --version
    carto 0.18.1 (Carto map stylesheet compiler)

## OSM-Carto-Style erzeugen
Das nachfolgende Beispiel verdeutlicht wie die Version 4.2.0 des OSM-Carto-Stils erzeugt wird.

    git clone https://github.com/gravitystorm/openstreetmap-carto.git osm-carto-4.2.0
    cd osm-carto-4.2.0
    git checkout v4.2.0

## Bootstrapping

    scripts/get-shapefiles.py
    cd data
    rm *.zip
    rm *.tgz

## Anpassung in project.mml
Die erforderliche Anpassung ist abhängig vom gewählten OSM-Datenbanknamen.
Wurde zum Beispiel als DB-Name "osmcarto4" gewählt, ist der Standard-DB-Name "gis" entsprechend anzupassen.

    dbname: "gis" -> dbname: "osmcarto4"

## mapnik.xml erzeugen

    carto -a 3.0.9 project.mml > mapnik.xml

## schwarzplan+-Style
Der Style benötigt zusätzlich die Daten "land-polygons-split-3857".
Es können die entsprechenden Daten des OSM-Carto-Style verwendet werden (siehe oben).

    cd schwarzplan+-1.2.0
    cd data
    cp -r ~/printstyles/osm-carto-4.2.0/data/land-polygons-split-3857 .

## Printmaps-Konfigurationen anpassen

    printmaps_buildservice.yaml
        - Pfade
    printmaps_webservice_capabilities.json
        - Release
        - Date
        - Layers

Bei einem neuen Release können neue Layer hinzukommen oder wegfallen. Um die jeweils aktuellen
Layer zu ermitteln, empfiehlt sich folgende Vorgehensweise:
- printmaps_buildservice.yaml: "testmode: true"
- map.yaml: keine Layer ausblenden (HideLayers auskommentieren)
- Testkarte erzeugen
- die Layer des Kartenstils finden sich im Logfile (printmaps_buildservice.log)
- Layer nach printmaps_webservice_capabilities.json übernehmen
- abschließend: Testmode deaktivieren ("testmode: false")

---

to be done - english translation
