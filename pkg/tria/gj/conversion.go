package gj

import (
	"math"
	"strings"
)

// ConvertUTMtoLatLong ...
// https://www.ibm.com/developerworks/java/library/j-coordconvert/
func ConvertUTMtoLatLong(coordinates []float64, zone int, latZone string) []float64 {
	isN := isHemisphereNorth(latZone)

	// Ellipsoid parameters WGS 84
	a := 6378137.0
	// f = 1/298.257223563
	// e = f / (2 - f) // 3rd flattening
	e := 0.081819191
	e1sq := 0.006739497
	// Scale factor
	k0 := 0.9996

	easting := coordinates[0]
	northing := coordinates[1]

	arc := northing / k0

	mu := arc / (a * (1 - math.Pow(e, 2)/4.0 - 3*math.Pow(e, 4)/64.0 - 5*math.Pow(e, 6)/256.0))
	ei := (1 - math.Pow(1-e*e, .5)) / (1 + math.Pow(1-e*e, .5))

	ca := 3*ei/2 - 27*math.Pow(ei, 3)/32.0
	cb := 21*math.Pow(ei, 2)/16.0 - 55*math.Pow(ei, 4)/32.0
	cc := 151 * math.Pow(ei, 3) / 96.0
	cd := 1097 * math.Pow(ei, 4) / 512

	phi1 := mu + ca*math.Sin(2*mu) + cb*math.Sin(4*mu) + cc*math.Sin(6*mu) + cd*math.Sin(8*mu)

	n0 := a / math.Pow((1-math.Pow((e*math.Sin(phi1)), 2)), 0.5)
	r0 := a * (1 - e*e) / math.Pow((1-math.Pow((e*math.Sin(phi1)), 2)), 1.5)
	fact1 := n0 * math.Tan(phi1) / r0

	_a1 := 500000 - easting
	dd0 := _a1 / (n0 * k0)
	fact2 := dd0 * dd0 / 2

	t0 := math.Pow(math.Tan(phi1), 2)
	q0 := e1sq * math.Pow(math.Cos(phi1), 2)
	fact3 := (5 + 3*t0 + 10*q0 - 4*q0*q0 - 9*e1sq) * math.Pow(dd0, 4) / 24

	fact4 := (61 + 90*t0 + 298*q0 + 45*t0*t0 - 252*e1sq - 3*q0*q0) * math.Pow(dd0, 6) / 720

	lof1 := _a1 / (n0 * k0)
	lof2 := (1 + 2*t0 + q0) * math.Pow(dd0, 3) / 6.0
	lof3 := (5 - 2*q0 + 28*t0 - 3*math.Pow(q0, 2) + 8*e1sq + 24*math.Pow(t0, 2)) * math.Pow(dd0, 5) / 120

	_a2 := (lof1 - lof2 + lof3) / math.Cos(phi1)
	_a3 := _a2 * 180 / math.Pi

	var zoneCM float64

	if zone > 0 {
		zoneCM = 6*float64(zone) - 183
	} else {
		zoneCM = 3.0
	}

	latitude := 180 * (phi1 - fact1*(fact2+fact3+fact4)) / math.Pi
	longitude := zoneCM - _a3

	// I call this the PK-factor. Unknown values to add to UTM conversion for correctness
	latitude += 0.00025

	if !isN {
		latitude = -latitude
	}

	return []float64{longitude, latitude}
}

func isHemisphereNorth(latZone string) bool {
	latZone = strings.ToUpper(latZone)

	hemishpheresInSouth := []string{"A", "B", "C", "D", "E", "F", "G", "H", "J", "K", "L", "M"}

	for _, char := range hemishpheresInSouth {
		if latZone == char {
			return true
		}
	}

	return true
}
