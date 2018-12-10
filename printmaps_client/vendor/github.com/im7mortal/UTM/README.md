[![Build Status](https://travis-ci.org/im7mortal/UTM.svg)](https://travis-ci.org/im7mortal/UTM)
[![Coverage Status](https://coveralls.io/repos/im7mortal/UTM/badge.svg?branch=master)](https://coveralls.io/r/im7mortal/UTM?branch=master)
[![GoDoc](https://godoc.org/github.com/im7mortal/UTM?status.svg)](https://godoc.org/github.com/im7mortal/UTM)

UTM
===

Bidirectional UTM-WGS84 converter for golang. It use logic from [UTM python version](https://pypi.python.org/pypi/utm) by Tobias Bieniek

Usage
-----

	go get github.com/im7mortal/UTM

Convert a latitude, longitude into an UTM coordinate

```go
	easting, northing, zoneNumber, zoneLetter, err := UTM.FromLatLon(40.71435, -74.00597, false)
```

Convert an UTM coordinate into a latitude, longitude.

```go
	latitude, longitude, err := UTM.ToLatLon(377486, 6296562, 30, "V")
```

Since the zone letter is not strictly needed for the conversion you may also
the ``northern`` parameter instead, which is a named parameter and can be set
to either ``true`` or ``false``. In this case you should define fields clearly(!).
You can't set ZoneLetter or northern both.

```go
	latitude, longitude, err := UTM.ToLatLon(377486, 6296562, 30, "", false)
```

The UTM coordinate system is explained on this [Wikipedia page](https://en.wikipedia.org/wiki/Universal_Transverse_Mercator_coordinate_system)

Speed
-----

Benchmark             | Amount of iterations | Average speed
--------------------- | -------------------- | -------------
ToLatLon              | 10000000             | 123 ns/op
ToLatLonWithNorthern  | 10000000             | 121 ns/op
FromLatLon            | 20000000             | 80.6 ns/op

> go test -bench=.

Full example
-----------
```go
package main

import (
	"github.com/im7mortal/UTM"
	"fmt"
)

func main() {

	easting, northing, zoneNumber, zoneLetter, err := UTM.FromLatLon(40.71435, -74.00597, false)
	if err != nil {
		panic(err.Error())
	}
	fmt.Println(
		fmt.Sprintf(
			"Easting: %d; Northing: %d; ZoneNumber: %d; ZoneLetter: %s;",
			easting,
			northing,
			zoneNumber,
			zoneLetter,
		))

	easting, northing, zoneNumber, zoneLetter, err = UTM.FromLatLon(40.71435, -74.00597, true)
	if err != nil {
		panic(err.Error())
	}
	fmt.Println(
		fmt.Sprintf(
			"Easting: %d; Northing: %d; ZoneNumber: %d; ZoneLetter: %s;",
			easting,
			northing,
			zoneNumber,
			zoneLetter,
		))

	latitude, longitude, err := UTM.ToLatLon(377486, 6296562, 30, "", true)
	fmt.Println(fmt.Sprintf("Latitude: %.5f; Longitude: %.5f;", latitude, longitude))

	latitude, longitude, err = UTM.ToLatLon(377486, 6296562, 30, "V")
	fmt.Println(fmt.Sprintf("Latitude: %.5f; Longitude: %.5f;", latitude, longitude))

}

```

Authors
-------

* Petr Lozhkin <im7mortal@gmail.com>

License
-------

Permission is hereby granted, free of charge, to any person obtaining a copy of this software and associated documentation files (the "Software"), to deal in the Software without restriction, including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
