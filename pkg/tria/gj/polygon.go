package gj

import (
	"strconv"

	"github.com/eesrc/geo/pkg/tria/geometry"
	geojson "github.com/paulmach/go.geojson"
)

// NewPolygonFeaturesFromPolygon returns a list of GeoJSON Features from a polygon. If includeTriangles
// is true, the triangulated shapes will be added as a separate feature for each triangle.
func NewPolygonFeaturesFromPolygon(polygon *geometry.Polygon, includeTriangles bool) []*geojson.Feature {
	var polygonFeatures []*geojson.Feature

	polygonFeature := GetFeatureFromPoints(polygon.AlphaShape, map[string]interface{}{
		"name": polygon.GetName(),
	})

	for k, v := range polygon.GetProperties() {
		polygonFeature.SetProperty(k, v)

	}

	polygonFeatures = append(polygonFeatures, polygonFeature)

	if includeTriangles {
		// Each triangle is it's own feature
		for i, triangle := range polygon.Triangles {
			polygonFeatures = append(polygonFeatures, GetFeatureFromPoints(triangle, map[string]interface{}{"name": polygon.GetName() + " - Triangle " + strconv.Itoa(i)}))
		}
	}

	return polygonFeatures
}
