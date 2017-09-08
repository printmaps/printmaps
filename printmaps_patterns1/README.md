# Patterns 1

Das Utilityprogramm "printmaps_patterns1" erzeugt ein Set an SVG-Pattern (Schraffuren, Raster, Punkte, ...), die für kartografische Zwecke hilfreich sind. Beispiel:

    <?xml version="1.0" encoding="UTF-8"?>
    <svg xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink" style="isolation:isolate" viewBox="0 0   200 200" width="200" height="200">
       <defs>
          <pattern id="hatch" width="20" height="20" patternTransform="rotate(45 0 0)" patternUnits="userSpaceOnUse">
            <line x1="0" y1="0" x2="0" y2="20" style="stroke:#81BDE3; stroke-width:6"/>
          </pattern>
       </defs>
       <rect x="0" y="0" width="200" height="200" fill="url(#hatch)"/>
    </svg>

![](sample-pattern.png)

Es hat sich jedoch herausgestellt, dass das vorstehende SVG-Pattern-Schema (<pattern> ... </pattern>) derzeit **nicht** von mapnik (3.0.13) verwendet werden kann. Ab der Version mapnik 3.1.0 soll dies jedoch möglich sein.

Siehe auch "Patterns 2".

---

to be done - english translation
