package gj

// MapProjection is the type of projection for given coordinates in a GEOJSON format
type MapProjection int

const (
	// WGS84Projection Use latitude longitude to determine position on a map
	WGS84Projection = 0
	// UTMProjection Use Universal Transverse Mercator basing calculations on a zone and latitude zone (hardcoded to "33" and "Z")
	UTMProjection = 1
)

// NewMapProjectionFromString Returns a MapProjection based on string input. Defaults to WGS84.
func NewMapProjectionFromString(projection string) MapProjection {
	if projection == "UTM" {
		return UTMProjection
	}

	return WGS84Projection
}
