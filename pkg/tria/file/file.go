package file

import (
	"encoding/gob"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/eesrc/geo/pkg/model"
	"github.com/eesrc/geo/pkg/tria/geometry"

	geojson "github.com/paulmach/go.geojson"
)

// SaveShapes saves the given shapes to a gob-encoded file for later use
func SaveShapes(path, name string, shapes []geometry.Shape) error {
	shapeStore := geometry.ShapesToShapeStore(shapes)

	if _, err := os.Stat(path); os.IsNotExist(err) {
		err = os.Mkdir(path, 0755)
		if err != nil {
			return err
		}
	}

	if !strings.HasSuffix(path, "/") {
		path += "/"
	}

	return writeGob(path+name, shapeStore)
}

// SaveShapeCollection saves the given shape collection to a gob-encoded file for later use
func SaveShapeCollection(path, name string, shapeCollection model.ShapeCollection) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		err = os.Mkdir(path, 0755)
		if err != nil {
			return err
		}
	}

	if !strings.HasSuffix(path, "/") {
		path += "/"
	}

	return writeGob(path+name, shapeCollection)
}

// SaveShapeCollectionFile saves the given shape collection to a gob-encoded file for later use
func SaveShapeCollectionFile(path, name string, shapeCollectionFile ShapeCollectionFile) error {
	registerGobShapeCollectionInterfaces()

	if _, err := os.Stat(path); os.IsNotExist(err) {
		err = os.Mkdir(path, 0755)
		if err != nil {
			return err
		}
	}

	if !strings.HasSuffix(path, "/") {
		path += "/"
	}

	return writeGob(path+name, shapeCollectionFile)
}

// LoadShapes loads a prior saved file and deserializes it to a list of shapes
func LoadShapes(path, name string) ([]geometry.Shape, error) {
	shapeStore := &geometry.ShapeStore{}

	if !strings.HasSuffix(path, "/") {
		path += "/"
	}

	filePath := path + name

	err := loadGob(filePath, shapeStore)

	if err != nil {
		return []geometry.Shape{}, err
	}

	shapes := geometry.ShapeStoreToShapes(shapeStore)

	return shapes, nil
}

// LoadShapeCollection loads a prior saved file and deserializes it to a shape collection
func LoadShapeCollection(path, name string) (*model.ShapeCollection, error) {
	shapeCollection := &model.ShapeCollection{}

	if !strings.HasSuffix(path, "/") {
		path += "/"
	}

	filePath := path + name

	err := loadGob(filePath, shapeCollection)

	if err != nil {
		return shapeCollection, err
	}

	return shapeCollection, nil
}

// LoadShapeCollectionFile loads a prior saved file and deserializes it to a shape collection
// along with a corresponding shape store
func LoadShapeCollectionFile(path, name string) (*ShapeCollectionFile, error) {
	shapeCollectionFile := ShapeCollectionFile{}
	registerGobShapeCollectionInterfaces()

	if !strings.HasSuffix(path, "/") {
		path += "/"
	}

	filePath := path + name

	log.Info("Loading", filePath)

	err := loadGob(filePath, &shapeCollectionFile)

	if err != nil {
		return &shapeCollectionFile, err
	}

	return &shapeCollectionFile, nil
}

// ReadGeoJSONFeatureCollection reads a GeoJSON feature collection and returns as a geojson-struct
func ReadGeoJSONFeatureCollection(path, file string, inputType InputType) (geojson.FeatureCollection, error) {
	if !strings.HasSuffix(path, "/") {
		path += "/"
	}

	filePath, _ := filepath.Abs(path + file)
	jsonFile, err := os.Open(filePath)

	if err != nil {
		return *geojson.NewFeatureCollection(), err
	}

	jsonBytes, err := ioutil.ReadAll(jsonFile)
	jsonFile.Close()

	if err != nil {
		fmt.Println(err)
		log.Fatal("Could not read file")
	}

	var featureCollection geojson.FeatureCollection

	switch inputType {
	case GeoJSON:
		err = json.Unmarshal(jsonBytes, &featureCollection)
		if err != nil {
			return geojson.FeatureCollection{}, err
		}
	default:
		return geojson.FeatureCollection{}, errors.New("Non-supported input type for file")
	}

	return featureCollection, nil
}

func loadGob(file string, object interface{}) error {
	gobFile, err := os.Open(file)
	defer closeFile(gobFile)

	if err != nil {
		return err
	}

	decoder := gob.NewDecoder(gobFile)
	return decoder.Decode(object)
}

func writeGob(file string, object interface{}) error {
	gobFile, err := os.Create(file)
	defer closeFile(gobFile)

	if err != nil {
		return err
	}

	encoder := gob.NewEncoder(gobFile)
	return encoder.Encode(object)
}

func closeFile(file *os.File) {
	err := file.Close()

	if err != nil {
		log.WithError(err).Errorf("Failed to close file '%s'", file.Name())
	}
}
