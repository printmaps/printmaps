// Package UTM is bidirectional UTM-WGS84 converter for golang
package UTM

import (
	"errors"
	"math"
	"unicode"
)

const (
	k0 = 0.9996
	e  = 0.00669438
	r  = 6378137
)

var e2 = e * e
var e3 = e2 * e
var eP2 = e / (1.0 - e)

var sqrtE = math.Sqrt(1 - e)

var fe = (1 - sqrtE) / (1 + sqrtE)
var fe2 = fe * fe
var fe3 = fe2 * fe
var fe4 = fe3 * fe
var fe5 = fe4 * fe

var m1 = 1 - e/4 - 3*e2/64 - 5*e3/256
var m2 = 3*e/8 + 3*e2/32 + 45*e3/1024
var m3 = 15*e2/256 + 45*e3/1024
var m4 = 35 * e3 / 3072

var p2 = 3./2*fe - 27./32*fe3 + 269./512*fe5
var p3 = 21./16*fe2 - 55./32*fe4
var p4 = 151./96*fe3 - 417./128*fe5
var p5 = 1097. / 512 * fe4

type zoneLetter struct {
	zone   int
	letter string
}

const x = math.Pi / 180

func rad(d float64) float64 { return d * x }
func deg(r float64) float64 { return r / x }

var zoneLetters = []zoneLetter{
	{84, " "},
	{72, "X"},
	{64, "W"},
	{56, "V"},
	{48, "U"},
	{40, "T"},
	{32, "S"},
	{24, "R"},
	{16, "Q"},
	{8, "P"},
	{0, "N"},
	{-8, "M"},
	{-16, "L"},
	{-24, "K"},
	{-32, "J"},
	{-40, "H"},
	{-48, "G"},
	{-56, "F"},
	{-64, "E"},
	{-72, "D"},
	{-80, "C"},
}

// ToLatLon convert Universal Transverse Mercator coordinates to a latitude and longitude
// Since the zone letter is not strictly needed for the conversion you may also
// the ``northern`` parameter instead, which is a named parameter and can be set
// to either true or false. In this case you should define fields clearly
// You can't set ZoneLetter or northern both.
func ToLatLon(easting, northing float64, zoneNumber int, zoneLetter string, northern ...bool) (latitude, longitude float64, err error) {

	northernExist := len(northern) > 0
	zoneLetterExist := !(zoneLetter == "")

	if !zoneLetterExist && !northernExist {
		err = inputError("either ZoneLetter or northern needs to be set")
		return
	} else if zoneLetterExist && northernExist {
		err = inputError("set either ZoneLetter or northern, but not both")
		return
	}
	if !(100000 <= easting && easting < 1000000) {
		err = inputError("easting out of range (must be between 100.000 m and 999.999 m")
		return
	}
	if !(0 <= northing && northing <= 10000000) {
		err = inputError("northing out of range (must be between 0 m and 10.000.000 m)")
		return
	}
	if !(1 <= zoneNumber && zoneNumber <= 60) {
		err = inputError("zone number out of range (must be between 1 and 60)")
		return
	}

	var northernValue bool

	if zoneLetterExist {
		zoneLetter := unicode.ToUpper(rune(zoneLetter[0]))
		if !('C' <= zoneLetter && zoneLetter <= 'X') || zoneLetter == 'I' || zoneLetter == 'O' {
			err = inputError("zone letter out of range (must be between C and X)")
			return
		}
		northernValue = zoneLetter >= 'N'
	} else {
		northernValue = northern[0]
	}

	x := easting - 500000
	y := northing

	if !northernValue {
		y -= 10000000
	}

	m := y / k0
	mu := m / (r * m1)

	pRad := mu +
		p2*math.Sin(2*mu) +
		p3*math.Sin(4*mu) +
		p4*math.Sin(6*mu) +
		p5*math.Sin(8*mu)

	pSin := math.Sin(pRad)
	pSin2 := pSin * pSin

	pCos := math.Cos(pRad)

	pTan := pSin / pCos
	pTan2 := pTan * pTan
	pTan4 := pTan2 * pTan2

	epSin := 1 - e*pSin2
	epSinSqrt := math.Sqrt(1 - e*pSin2)

	n := r / epSinSqrt
	rad := (1 - e) / epSin

	c := fe * pCos * pCos
	c2 := c * c

	d := x / (n * k0)
	d2 := d * d
	d3 := d2 * d
	d4 := d3 * d
	d5 := d4 * d
	d6 := d5 * d

	latitude = pRad - (pTan / rad) *
		(d2/2 -
			d4/24*(5+3*pTan2+10*c-4*c2-9*eP2)) +
		d6/720*(61+90*pTan2+298*c+45*pTan4-252*eP2-3*c2)

	longitude = (d -
		d3/6*(1+2*pTan2+c) +
		d5/120*(5-2*c+28*pTan2-3*c2+8*eP2+24*pTan4)) / pCos

	latitude = deg(latitude)
	longitude = deg(longitude) + float64(zoneNumberToCentralLongitude(zoneNumber))

	return

}

// ValidateLatLone check that latitude and longitude are valid.
func ValidateLatLone(latitude, longitude float64) error {
	if !(-80.0 <= latitude && latitude <= 84.0) {
		return inputError("latitude out of range (must be between 80 deg S and 84 deg N)")
	}
	if !(-180.0 <= longitude && longitude <= 180.0) {
		return inputError("longitude out of range (must be between 180 deg W and 180 deg E)")
	}
	return nil
}

// FromLatLon convert a latitude and longitude to Universal Transverse Mercator coordinates
func FromLatLon(latitude, longitude float64, northern bool) (easting, northing float64, zoneNumber int, zoneLetter string, err error) {
	// check that latitude and longitude are valid
	err = ValidateLatLone(latitude, longitude)
	if err != nil {
		return
	}

	latRad := rad(latitude)
	latSin := math.Sin(latRad)
	latCos := math.Cos(latRad)

	latTan := latSin / latCos
	latTan2 := latTan * latTan
	latTan4 := latTan2 * latTan2

	zoneNumber = latLonToZoneNumber(latitude, longitude)

	zoneLetter = latitudeToZoneLetter(latitude)

	if northern {
		// N north, S south
		if latitude > 0 {
			zoneLetter = "N"
		} else {
			zoneLetter = "S"
		}
	}

	lonRad := rad(longitude)
	centralLon := zoneNumberToCentralLongitude(zoneNumber)
	centralLonRad := rad(float64(centralLon))

	n := r / math.Sqrt(1-e*latSin*latSin)
	c := eP2 * latCos * latCos

	a := latCos * (lonRad - centralLonRad)
	a2 := a * a
	a3 := a2 * a
	a4 := a3 * a
	a5 := a4 * a
	a6 := a5 * a
	m := r * (m1*latRad -
		m2*math.Sin(2*latRad) +
		m3*math.Sin(4*latRad) -
		m4*math.Sin(6*latRad))
	easting = k0*n * (a +
		a3/6*(1-latTan2+c) +
		a5/120*(5-18*latTan2+latTan4+72*c-58*eP2)) + 500000
	northing = k0 * (m + n*latTan * (a2/2 +
		a4/24*(5-latTan2+9*c+4*c*c) +
		a6/720*(61-58*latTan2+latTan4+600*c-330*eP2)))

	if latitude < 0 {
		northing += 10000000
	}

	return
}

func latitudeToZoneLetter(latitude float64) string {
	for _, zoneLetter := range zoneLetters {
		if latitude >= float64(zoneLetter.zone) {
			return zoneLetter.letter
		}
	}
	return " "
}

func latLonToZoneNumber(latitude float64, longitude float64) int {
	if 56 <= latitude && latitude <= 64 && 3 <= longitude && longitude <= 12 {
		return 32
	}

	if 72 <= latitude && latitude <= 84 && longitude >= 0 {
		if longitude <= 9 {
			return 31
		} else if longitude <= 21 {
			return 33
		} else if longitude <= 33 {
			return 35
		} else if longitude <= 42 {
			return 37
		}
	}

	return int((longitude+180)/6) + 1
}

func zoneNumberToCentralLongitude(zoneNumber int) int {
	return (zoneNumber-1)*6 - 180 + 3
}

// InputError allow to distinguish if an error is from UTM conversion functions.
type InputError error

func inputError(text string) InputError {
	return InputError(errors.New(text))
}
