package gj

import (
	"fmt"

	"github.com/eesrc/geo/pkg/tria/geometry"

	geojson "github.com/paulmach/go.geojson"
)

// TransformFromGeoJSONCoordinatesToPoints Transforms a geoJSON geometry to points
func TransformFromGeoJSONCoordinatesToPoints(geoJSONGeometry *geojson.Geometry, projection MapProjection) [][]geometry.Point {
	var pointList [][]geometry.Point

	if geoJSONGeometry == nil {
		return pointList
	}

	if geoJSONGeometry.IsPoint() {
		pointList = append(pointList, []geometry.Point{getPointFromProjection(geoJSONGeometry.Point, projection)})
		return pointList
	}

	if geoJSONGeometry.IsPolygon() {
		for i := 0; i < len(geoJSONGeometry.Polygon); i++ {
			var polygonPoints []geometry.Point
			polygon := geoJSONGeometry.Polygon[i]

			for j := 0; j < len(polygon); j++ {
				polygonPoints = append(polygonPoints, getPointFromProjection(polygon[j], projection))
			}

			pointList = append(pointList, polygonPoints)
		}

		return pointList
	}

	if geoJSONGeometry.IsMultiPolygon() {
		for _, multipolygon := range geoJSONGeometry.MultiPolygon {
			for i := 0; i < len(multipolygon); i++ {
				var polygonPoints []geometry.Point
				polygon := multipolygon[i]

				for j := 0; j < len(polygon); j++ {
					polygonPoints = append(polygonPoints, getPointFromProjection(polygon[j], projection))
				}

				pointList = append(pointList, polygonPoints)
			}
		}

		return pointList
	}

	return pointList
}

func getPointFromProjection(point []float64, projection MapProjection) geometry.Point {
	if projection == UTMProjection {
		return getPointFromUTM(point, 33, "Z")
	}
	return getPointFromWGS84(point)
}

func getPointFromWGS84(point []float64) geometry.Point {
	return geometry.Point{X: point[0], Y: point[1]}
}

func getPointFromUTM(point []float64, zone int, latZone string) geometry.Point {
	convertedUTMPoints := ConvertUTMtoLatLong(point, zone, latZone)
	return geometry.Point{X: convertedUTMPoints[0], Y: convertedUTMPoints[1]}
}

// TransformToGeoJSONFeatureCollectionFromPoints Transforms a list of point2d to a geoJSON feature collection
func TransformToGeoJSONFeatureCollectionFromPoints(points [][]geometry.Point) geojson.FeatureCollection {
	featureCollection := geojson.NewFeatureCollection()

	if len(points) < 1 {
		return *featureCollection
	}

	for i, featurePoints := range points {
		featureCollection.AddFeature(GetFeatureFromPoints(featurePoints, map[string]interface{}{"name": fmt.Sprintf("Unnamed %d", i)}))
	}

	return *featureCollection
}

// NewGeoJSONFeatureCollectionFromShape returns a GeoJSON FeatureCollection based on a shape
func NewGeoJSONFeatureCollectionFromShape(shape geometry.Shape, includeTriangles bool) geojson.FeatureCollection {
	featureCollection := geojson.NewFeatureCollection()

	switch shape := shape.(type) {
	case *geometry.Polygon:
		polygon := shape

		geometries := NewPolygonFeaturesFromPolygon(polygon, includeTriangles)
		for _, geometry := range geometries {
			featureCollection.AddFeature(geometry)
		}

	case *geometry.Circle:
		circle := shape
		featureCollection.AddFeature(NewCircleFeatureFromCircle(*circle))

	default:
		return *featureCollection
	}

	return *featureCollection
}

// NewGeoJSONFeatureCollectionFromShapes returns a GeoJSON FeatureCollection based on a list of shapes
func NewGeoJSONFeatureCollectionFromShapes(shapes []geometry.Shape, includeTriangles bool) geojson.FeatureCollection {
	featureCollection := geojson.NewFeatureCollection()

	for _, shape := range shapes {
		features := GetFeaturesFromShape(shape, includeTriangles)

		for _, feature := range features {
			featureCollection.AddFeature(feature)
		}
	}

	return *featureCollection
}

// GetFeaturesFromShape returns a geoJSON feature from given shape
func GetFeaturesFromShape(shape geometry.Shape, includeTriangles bool) []*geojson.Feature {
	var features = make([]*geojson.Feature, 0)

	switch shape := shape.(type) {
	case *geometry.Polygon:
		polygon := shape
		features = append(features, NewPolygonFeaturesFromPolygon(polygon, includeTriangles)...)
	case *geometry.Circle:
		circle := shape
		features = append(features, NewCircleFeatureFromCircle(*circle))
	default:
		return features
	}

	return features
}

// GetFeatureFromPoints creates a GeoJSON feature from a list of points and a map of properties.
func GetFeatureFromPoints(points []geometry.Point, properties map[string]interface{}) *geojson.Feature {
	numberOfPoints := len(points)

	// GeoJSON needs the start and end point to be the same to be valid. Check if it already satisfies this criteria.
	matchingEnds := points[0].X == points[numberOfPoints-1].X && points[0].Y == points[numberOfPoints-1].Y

	numberOfGeoJSONPoints := numberOfPoints
	if !matchingEnds {
		numberOfGeoJSONPoints += 1
	}

	// Initiate three dimensional Feature point array for the GeoJSON Feature
	polyPoints := make([][][]float64, 1)

	polyPoints[0] = make([][]float64, numberOfGeoJSONPoints)
	for i := 0; i < numberOfGeoJSONPoints; i++ {
		polyPoints[0][i] = make([]float64, 2)
	}

	for i, point := range points {
		polyPoints[0][i][0] = point.X
		polyPoints[0][i][1] = point.Y
	}

	// If no matching ends, populate the last point with the first
	if !matchingEnds {
		polyPoints[0][len(points)][0] = points[0].X
		polyPoints[0][len(points)][1] = points[0].Y
	}

	feature := geojson.NewPolygonFeature(polyPoints)

	for key, value := range properties {
		feature.SetProperty(key, value)
	}

	return feature
}

// GetFeatureFromPoint creates a GeoJSON feature from a single point and a map of properties.
func GetFeatureFromPoint(point geometry.Point, name string) *geojson.Feature {
	feature := geojson.NewPointFeature([]float64{point.X, point.Y})

	feature.SetProperty("name", name)

	return feature
}

// GetFeatureName Returns a best guess for name of the feature. If no clue, returns "Unknown name"
func GetFeatureName(feature *geojson.Feature) string {
	// Basic case
	if feature.Properties["name"] != nil {
		return feature.Properties["name"].(string)
	}
	// Basic case
	if feature.Properties["NAME"] != nil {
		return feature.Properties["NAME"].(string)
	}

	// GeoNorge special case. They sometimes have the raw name as a string,
	// sometimes as a list of names, and sometimes as a list of map of strings.
	if feature.Properties["navn"] != nil {
		name, err := feature.Properties["navn"].(string)
		if err {
			return name
		}
		rawMunicipalNames := feature.Properties["navn"].([]interface{})
		municipalNames := make([]map[string]interface{}, len(rawMunicipalNames))
		for i := range rawMunicipalNames {
			municipalNames[i] = rawMunicipalNames[i].(map[string]interface{})
		}

		return municipalNames[0]["navn"].(string)
	}

	// A3 is a property used in representing world countries abbreviations
	if feature.Properties["A3"] != nil {
		return feature.Properties["A3"].(string)
	}

	// Adressetekst is the whole address for a local point
	if feature.Properties["adressetekst"] != nil {
		return feature.Properties["adressetekst"].(string)
	}

	// Postnummer is used for getting postal codes and postal place of norwegian postal polygons
	if feature.Properties["postnummer"] != nil && feature.Properties["poststed"] != nil {
		return fmt.Sprintf("Postnummer: %d - Poststed %s", feature.PropertyMustInt("postnummer", 0000), feature.Properties["poststed"])
	}

	// English names of some multilingual country region datasets
	if feature.Properties["name:en"] != nil {
		return feature.Properties["name:en"].(string)
	}

	return "Unknown name"
}
