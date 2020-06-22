package geometry

import (
	"database/sql/driver"
	"encoding/json"

	log "github.com/sirupsen/logrus"
)

// ShapeStore is a local struct to easier serialize/deserialize a list of shapes
type ShapeStore struct {
	Circles  []Circle
	Polygons []Polygon
}

// Value implements SQL value driver
func (shapeStore ShapeStore) Value() (driver.Value, error) {
	b, err := json.Marshal(shapeStore)
	if err != nil {
		return nil, err
	}
	return string(b), nil
}

// Scan implements SQL scan driver
func (shapeStore *ShapeStore) Scan(src interface{}) error {
	var data []byte
	if b, ok := src.([]byte); ok {
		data = b
	} else if s, ok := src.(string); ok {
		data = []byte(s)
	}
	return json.Unmarshal(data, shapeStore)
}

// ShapesToShapeStore creates a serializable ShapeStore from a list of shapes
func ShapesToShapeStore(shapes []Shape) ShapeStore {
	shapeStore := ShapeStore{}

	for _, shape := range shapes {
		if shape != nil {
			switch shape := shape.(type) {
			case *Polygon:
				polygon := shape
				shapeStore.Polygons = append(shapeStore.Polygons, *polygon)
			case *Circle:
				circle := shape
				shapeStore.Circles = append(shapeStore.Circles, *circle)
			default:
				log.Warn("I have no idea who I am", shape)
			}
		}
	}

	return shapeStore
}

// ShapeStoreToShapes deserializes a ShapeStore to a list of shapes
func ShapeStoreToShapes(store *ShapeStore) []Shape {
	var shapes []Shape

	for _, polygon := range store.Polygons {
		polygonCopy := polygon
		shapes = append(shapes, &polygonCopy)
	}

	for _, circle := range store.Circles {
		circleCopy := circle
		shapes = append(shapes, &circleCopy)
	}

	return shapes
}
