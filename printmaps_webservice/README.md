# Webservice

Der Webservice bedient die Kommunikation mit einem Client und besteht aus:
* printmaps_webservice
* printmaps_webservice.yaml
* printmaps_webservice_capabilities.json
* printmaps_webservice_maintenance.html
* printmaps_webservice_mapdata.poly

Die Konfiguration (printmaps_webservice.yaml) des Buildservices ist an die gewählte Installation anzupassen.

Build- und Webservice sind voneinander entkoppelte Prozesse.
Die Kommunikation zwischen beiden Prozessen erfolgt über Dateien.

Der Webservice erzeugt im laufenden Betrieb eine kompakte Logdatei (printmaps_webservice.log).
Im Fehler- oder Problemfall sollte diese Datei eingesehen werden.

## Webservice (als Hintergrundprozess) starten

    nohup ./printmaps_webservice 1>printmaps_webservice.out 2>&1 &

## Webservice stoppen

    ps -Af | grep "printmaps_"
    kill pid

---

to be done - english translation
