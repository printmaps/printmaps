# Updateservice

Der (optionale) Updateservice aktualisiert in regelmäßigen Zeitintervallen die Datenbank mit den OSM-Daten und besteht aus:
* printmaps_updater
* printmaps_updater.yaml

Die Konfiguration (printmaps_buildservice.yaml) des Updateservices ist an die gewünschten Einstellungen anzupassen.

Der Updateservice erzeugt im laufenden Betrieb eine kompakte Logdatei (printmaps_updater.log).
Im Fehler- oder Problemfall sollte diese Datei eingesehen werden.

Der Updateservice ist ein Wrapper um die Tools:
* osmosis
* osm2pgsql

Das Verständnis der Funktionsweise der vorgenannten Tools ist von essentieller Bedeutung. 

## Updateservice (als Hintergrundprozess) starten

    nohup ./printmaps_updater 1>printmaps_updater.out 2>&1 &

## Updateservice stoppen

    ps -Af | grep "printmaps_"
    kill pid

## Initialisierung
Zunächst wird die OSM-Datenbank via "osm2pgsql" mit dem initialen Datenbestand geladen.
Anschließend ist der aktuelle Update-Status zu ermitteln (Beispiel):

    curl --location --url "http://download.geofabrik.de/europe/dach-updates/state.txt" --output "state.txt"

Das Tool "osmosis" für Updates vorbereiten:

    osmosis --rrii workingDirectory=./

Im aktuellen Verzeichnis werden angelegt:

    -rw-rw-r-- 1 printmaps printmaps  252 May 10 15:55 configuration.txt
    -rw-rw-r-- 1 printmaps printmaps    0 May 10 15:55 download.lock

Die Basis-URL in der Datei "configuration.txt" ist anpassen:

    baseUrl=http://download.geofabrik.de/europe/dach-updates/

---

to be done - english translation
