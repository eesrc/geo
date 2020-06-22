package gj

import (
	"github.com/eesrc/geo/pkg/tria/geometry"
	geojson "github.com/paulmach/go.geojson"
)

// NewCircleFeatureFromPoint Returns a GeoJSON PointFeature with property radius set to current radius
func NewCircleFeatureFromPoint(point geometry.Point, radius float64) *geojson.Feature {
	pointFeature := geojson.NewPointFeature([]float64{point.X, point.Y})

	pointFeature.SetProperty("radius", radius)

	return pointFeature
}

// NewCircleFeatureFromCircle Returns a GeoJSON PointFeature with property "radius" set to current radius and "name" as the circle name
func NewCircleFeatureFromCircle(circle geometry.Circle) *geojson.Feature {
	pointFeature := NewCircleFeatureFromPoint(circle.Origo, circle.Radius)

	for k, v := range circle.GetProperties() {
		pointFeature.SetProperty(k, v)
	}

	pointFeature.SetProperty("name", circle.Name)
	pointFeature.SetProperty("id", circle.GetID())

	return pointFeature
}
