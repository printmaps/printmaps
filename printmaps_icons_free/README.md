# Icons

Die von Printmaps verwendeten Map-Marker-Icons stammen von "Icons-Land" (http://www.icons-land.com/flat-mapmarkers-svg-icons.php).
Das vollumfängliche Icon-Set unterliegt Lizenzbeschränkungen und darf nicht veröffentlicht werden.
Für den produktiven Einsatz von Printmaps empfiehlt sich die Lizensierung des Icon-Sets.

Es steht jedoch auch ein eingeschränktes Icon-Set (Ball_Right, Flag4_Left, PushPin_Left, Flag1_Right) für die freie Nutzung zur Verfügung.
Auch hier gilt es die Lizenzbedingungen (siehe IconsLandDemoLicenseAgreement.txt) zu beachten.

![](sample-needle.png)

Das Programm "printmaps_icons" modifiziert die Icons weder in Form noch Farbe (was gemäß Lizenzbedingungen
auch unzulässig ist), sondern zentriert lediglich die Nadelspitze aller Objekte in die Iconmitte.
Hierdurch zeigt die Nadelspitze (Iconmitte) immer exakt auf die geografischen Koordinaten des zu markierenden 
Punktes (Node, POI).

Desweiteren werden die Dateien umbenannt, sodass sie mit "Printmaps_" beginnen.
Icons und Patterns die diesem Namensschema entsprechen, werden als interne Objekte behandelt
und stehen für die Nutzung auf einer Karte direkt zur Verfügung.

Die Icons sollten, zusammen mit allen Printmaps_-Patterns, in ein Verzeichnis "markers" kopiert werden.
Dieses Verzeichnis ist dem Buildservice bekannt zu machen (printmaps_buildservice.yaml).

---

to be done - english translation