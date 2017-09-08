# Buildservice

Der Buildservice erzeugt eine Karte und besteht aus:
* printmaps_buildservice
* printmaps_buildservice.yaml

Die Konfiguration (printmaps_buildservice.yaml) des Buildservices ist an die gewählte Installation anzupassen.

Build- und Webservice sind voneinander entkoppelte Prozesse.
Die Kommunikation zwischen beiden Prozessen erfolgt über Dateien.

Die eigentliche Erzeugung einer großformatigen Karte erfolgt via Aufruf des Programmes "nik4-printmaps.py" (siehe Nik4).

Der Buildservice erzeugt im laufenden Betrieb eine kompakte Logdatei (printmaps_buildservice.log).
Im Fehler- oder Problemfall sollte diese Datei eingesehen werden.

## Buildservice (als Hintergrundprozess) starten

    nohup ./printmaps_buildservice 1>printmaps_buildservice.out 2>&1 &

## Buildservice stoppen

    ps -Af | grep "printmaps_"
    kill pid

---

to be done - english translation
