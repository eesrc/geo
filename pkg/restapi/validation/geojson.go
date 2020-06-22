package validation

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/eesrc/geo/pkg/tria/geometry"
	"github.com/eesrc/geo/pkg/tria/gj"
	"github.com/eesrc/geo/pkg/tria/triangulation"
	geojson "github.com/paulmach/go.geojson"
)

// GetShapesFromBody retrieves shapes from a FeatureCollection in given body,
// decodes it and performs a triangulation on the given geojson Features
// Returns a validation error containing an ErrorResponse if something went wrong
func GetShapesFromBody(body io.ReadCloser) ([]geometry.Shape, error) {
	bodyBytes, _ := ioutil.ReadAll(body)

	featureCollection, err := geojson.UnmarshalFeatureCollection(bodyBytes)

	if err != nil {
		if err, ok := err.(*json.UnmarshalTypeError); ok {
			return []geometry.Shape{}, getUnmarshalError(err)
		}

		return []geometry.Shape{}, newError(
			NewErrorResponse(
				http.StatusBadRequest,
				NewParameterErrorDetail("featureCollection", "You need to provide a valid GeoJSON FeatureCollection object"),
			),
		)
	}

	// Check for correct type
	if featureCollection.Type != "FeatureCollection" {
		return []geometry.Shape{}, newError(
			NewErrorResponse(
				http.StatusBadRequest,
				NewParameterErrorDetail("featureCollection.type", fmt.Sprintf("You need to provide a GeoJSON FeatureCollection object. You provided a '%s'", featureCollection.Type)),
			),
		)
	}

	// Try to triangulate input GeoJSON.
	shapes, err := triangulation.TriangulateGeoJSONFeatureCollectionToShapes(featureCollection, gj.WGS84Projection)

	if err != nil {
		return []geometry.Shape{}, newError(
			NewErrorResponse(
				http.StatusBadRequest,
				NewParameterErrorDetail("featureCollection.features", "One or more polygons are too complex. Make sure the polygon provided is non-complex and non-overlapping."),
			),
		)
	}

	return shapes, nil
}

// ValidateAndGetShapeFromBody retrieves shapes from a GeoJSON Feature in given body,
// decodes it and performs a triangulation on the given Feature
// Returns a validation error containing an ErrorResponse if something went wrong
func GetShapeFromBody(body io.ReadCloser) (geometry.Shape, error) {
	jsonDecoder := json.NewDecoder(body)
	var feature *geojson.Feature

	// Try decoding Feature input
	err := jsonDecoder.Decode(&feature)

	if feature.Type != "Feature" {
		return &geometry.Polygon{}, newError(
			NewErrorResponse(
				http.StatusBadRequest,
				NewParameterErrorDetail(
					"feature",
					fmt.Sprintf("You need to provide a valid GeoJSON Feature object. You provided a '%s'", feature.Type),
				),
			),
		)
	}

	if err != nil {
		if err, ok := err.(*json.UnmarshalTypeError); ok {
			return &geometry.Polygon{}, getUnmarshalError(err)
		}

		return &geometry.Polygon{}, newError(
			NewErrorResponse(
				http.StatusBadRequest,
				NewParameterErrorDetail("feature", "You need to provide a valid GeoJSON Feature object"),
			),
		)
	}

	// Try to triangulate input GeoJSON.
	shape, err := triangulation.TriangulateGeoJSONFeatureToShape(feature, gj.WGS84Projection)

	if err != nil {
		return &geometry.Polygon{}, newError(
			NewErrorResponse(
				http.StatusBadRequest,
				NewParameterErrorDetail("feature", "The polygon provided are too complex. Make sure the polygon is non-complex and non-overlapping."),
			),
		)
	}

	return shape, nil
}
