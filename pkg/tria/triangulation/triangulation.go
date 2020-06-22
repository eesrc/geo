package triangulation

import (
	"github.com/eesrc/geo/pkg/tria/file"
	"github.com/eesrc/geo/pkg/tria/geometry"
	"github.com/eesrc/geo/pkg/tria/gj"
	geojson "github.com/paulmach/go.geojson"
)

// TriangulatePoints initiates a triangulation based on a single dimensioned polygon
// Returns a double dimensioned array consisting of triangles
func TriangulatePoints(points []geometry.Point) ([][]geometry.Point, error) {
	return TriangulateByEarCut(points)
}

// NewTriangulatedPolygonFromPoints creates a fully triangulated Polygon from a list of points
func NewTriangulatedPolygonFromPoints(points []geometry.Point) (geometry.Polygon, error) {
	triangles, err := TriangulatePoints(points)

	if err != nil {
		return geometry.Polygon{}, err
	}

	return geometry.Polygon{
		Name:        "Unnamed",
		Description: "No description",
		AlphaShape:  points,
		Triangles:   triangles,
		BoundingBox: geometry.CalculateBoundingBox(points),
	}, nil
}

// TriangulateShapesFromFile tries to fetch shapes from a GeoJSON file.
// The shapes will be triangulated if necesssary, cached, and saved besides the file in a gob-format.
// The next load will use the cached entries.
func TriangulateShapesFromFile(filePath, fileName string, fileType file.InputType, mapProjection gj.MapProjection) ([]geometry.Shape, error) {
	// If we found shapes in a cached file, we return the cached result
	shapes, err := file.LoadShapes(filePath, fileName)
	if err == nil {
		return shapes, nil
	}

	// No cached file found, read file and go through with triangulation
	geoJSONData, err := file.ReadGeoJSONFeatureCollection(filePath, fileName, fileType)

	if err != nil {
		return shapes, err
	}

	shapes, err = TriangulateGeoJSONFeatureCollectionToShapes(&geoJSONData, mapProjection)

	if err != nil {
		return shapes, err
	}

	err = file.SaveShapes(filePath, fileName, shapes)

	if err != nil {
		return shapes, err
	}

	return shapes, nil
}

// TriangulateGeoJSONFeatureToShape takes a GeoJSON Feature as input and triangulates it to a shape.
func TriangulateGeoJSONFeatureToShape(feature *geojson.Feature, mapProjection gj.MapProjection) (geometry.Shape, error) {
	var shape geometry.Shape

	name := gj.GetFeatureName(feature)
	pointList := gj.TransformFromGeoJSONCoordinatesToPoints(feature.Geometry, mapProjection)

	for _, points := range pointList {
		// If the length is 1, it's a single point. Single points are not really supported, but we make a circle.
		if len(points) == 1 {
			// We set the radius to a default 30m, with the option of overloading with a property in the geometry property
			circle := geometry.Circle{
				Origo:  points[0],
				Radius: feature.PropertyMustFloat64("radius", 30.0/111111.0),
				Name:   name,
			}

			shape = &circle
		} else {
			polygon, err := NewTriangulatedPolygonFromPoints(points)
			polygon.Name = name

			if err != nil {
				return shape, err
			}

			shape = &polygon
		}

		shape.SetProperties(feature.Properties)
	}

	return shape, nil
}

// TriangulateGeoJSONFeatureCollectionToShapes takes a GeoJSON FeatureCollection and returns it as a list of shapes.
// You need to provide a MapProjection so we know which MapProjection the provided dataset has.
func TriangulateGeoJSONFeatureCollectionToShapes(featureCollection *geojson.FeatureCollection, mapProjection gj.MapProjection) ([]geometry.Shape, error) {
	var shapes []geometry.Shape

	for i := 0; i < len(featureCollection.Features); i++ {
		shape, err := TriangulateGeoJSONFeatureToShape(featureCollection.Features[i], mapProjection)

		if err != nil {
			return shapes, err
		}

		shapes = append(shapes, shape)
	}

	return shapes, nil
}
