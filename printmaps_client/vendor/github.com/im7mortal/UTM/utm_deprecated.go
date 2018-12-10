package UTM

// Don't use it. It panic and return no error.
// FromLatLonF convert a latitude and longitude to Universal Transverse Mercator coordinates.
//
// Deprecated: Use FromLatLon functions to converse instead.
func FromLatLonF(lat, lon float64) (easting, northing float64) {
	var err error
	// Northing always false in this implementation.
	easting, northing, _, _, err = FromLatLon(lat, lon, false)
	if err != nil {
		panic(err)
	}
	return
}

//Coordinate contains coordinates in the Universal Transverse Mercator coordinate system
//
// Deprecated: Use ToLatLon functions to convert LatLon instead.
type Coordinate struct {
	Easting    float64
	Northing   float64
	ZoneNumber int
	ZoneLetter string
}

// FromLatLon convert a latitude and longitude to Universal Transverse Mercator coordinates
//
// Deprecated: Use FromLatLon functions to convert LatLon instead.
func (point *LatLon) FromLatLon() (coord Coordinate, err error) {
	// Northing always false in this implementation.
	coord.Easting, coord.Northing, coord.ZoneNumber, coord.ZoneLetter, err = FromLatLon(point.Latitude, point.Longitude, false)
	return
}

//LatLon contains a latitude and longitude
//
// Deprecated: Use FromLatLon functions to convert LatLon instead.
type LatLon struct {
	Latitude  float64
	Longitude float64
}

// ToLatLon convert Universal Transverse Mercator coordinates to a latitude and longitude
// Since the zone letter is not strictly needed for the conversion you may also
// the ``northern`` parameter instead, which is a named parameter and can be set
// to either true or false. In this case you should define fields clearly
// You can't set ZoneLetter or northern both.
//
// Deprecated: Use ToLatLon functions to convert LatLon instead.
func (coordinate *Coordinate) ToLatLon(northern ...bool) (LatLon, error) {

	latitude, longitude, err := ToLatLon(coordinate.Easting, coordinate.Northing, coordinate.ZoneNumber, coordinate.ZoneLetter, northern...)
	return LatLon{latitude, longitude}, err

}
