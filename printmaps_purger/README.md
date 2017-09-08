# Purgeservice

Der (optionale) Purgeservice löscht in regelmäßigen Zeitintervallen serverseitig veraltete Kartendaten und besteht aus:
* printmaps_purger
* printmaps_purger.yaml

Die Konfiguration (printmaps_buildservice.yaml) des Purgeservices ist an die gewünschten Einstellungen anzupassen.

Der Purgeservice erzeugt im laufenden Betrieb eine kompakte Logdatei (printmaps_purger.log).
Im Fehler- oder Problemfall sollte diese Datei eingesehen werden.

## Purgeservice (als Hintergrundprozess) starten

    nohup ./printmaps_purger 1>printmaps_purger.out 2>&1 &

## Purgeservice stoppen

    ps -Af | grep "printmaps_"
    kill pid

---

to be done - english translation
