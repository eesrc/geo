package sqlitestore

import (
	"bytes"
	"database/sql/driver"
	"encoding/gob"

	"github.com/eesrc/geo/pkg/model"
	"github.com/eesrc/geo/pkg/serializing"
	"github.com/eesrc/geo/pkg/tria/geometry"
	log "github.com/sirupsen/logrus"
)

// shapeStorageModel represents a storable list of shapes
type shapeStorageModel struct {
	ShapeType  shapeType
	ShapeBytes []byte
}

type shapeType string

const (
	polygonShape shapeType = "polygon"
	circleShape  shapeType = "circle"
)

func init() {
	gob.Register([]interface{}{})
	gob.Register(map[string]interface{}{})
	gob.Register(&geometry.Polygon{})
	gob.Register(&geometry.Circle{})
}

func shapeStoragefromShapeModel(shapeModel *model.Shape) (shapeStorageModel, error) {
	storageModel := shapeStorageModel{}

	var buff bytes.Buffer
	encoder := gob.NewEncoder(&buff)
	err := encoder.Encode(&shapeModel.Shape)
	if err != nil {
		log.WithError(err).Error("Failed to serialize shape")
		return storageModel, err
	}

	storageModel.ShapeBytes = buff.Bytes()

	switch shapeModel.Shape.(type) {
	case *geometry.Polygon:
		storageModel.ShapeType = polygonShape
	case *geometry.Circle:
		storageModel.ShapeType = circleShape
	}

	return storageModel, nil
}

func shapeModelFromShapeStorage(storageModel *shapeStorageModel) (geometry.Shape, error) {
	var shape geometry.Shape
	decoder := gob.NewDecoder(bytes.NewBuffer(storageModel.ShapeBytes))
	err := decoder.Decode(&shape)
	if err != nil {
		log.WithError(err).Errorf("Failed to deserialize polygon. Length: %d", len(storageModel.ShapeBytes))
		return nil, err
	}
	return shape, nil
}

// Value implements SQL value driver
func (shapeStore shapeStorageModel) Value() (driver.Value, error) {
	return serializing.ValueJSON(shapeStore)
}

// Scan implements SQL scan driver
func (shapeStore *shapeStorageModel) Scan(src interface{}) error {
	return serializing.ScanJSON(shapeStore, src)
}
