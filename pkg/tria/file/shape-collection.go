package file

import (
	"encoding/gob"

	"github.com/eesrc/geo/pkg/model"
	"github.com/eesrc/geo/pkg/tria/geometry"
)

// ShapeCollectionFile is a file representation of a complete ShapeCollection
// including its shapes
type ShapeCollectionFile struct {
	ShapeCollection model.ShapeCollection
	ShapeStore      geometry.ShapeStore
}

func registerGobShapeCollectionInterfaces() {
	gob.Register([]interface{}{})
	gob.Register(map[string]interface{}{})
}
