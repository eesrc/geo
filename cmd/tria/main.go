package main

import (
	"flag"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/eesrc/geo/pkg/model"
	"github.com/eesrc/geo/pkg/tria/file"
	"github.com/eesrc/geo/pkg/tria/geometry"
	"github.com/eesrc/geo/pkg/tria/gj"
	"github.com/eesrc/geo/pkg/tria/triangulation"
)

const (
	defaultMapProjection = "WGS84"
	defaultOutputType    = "gob"
	defaultOutputFolder  = "output"

	defaultName        = "Unnamed"
	defaultDescription = "No description"
	defaultID          = 0
	defaultTeamID      = 0
)

var (
	// command line flags
	inputFile     = flag.String("i", "", "Input file for triangulation")
	mapProjection = flag.String("p", defaultMapProjection, "Map projection for triangulation {WGS84,UTM}")
	outputType    = flag.String("t", defaultOutputType, "Type of output for triangulation {geojson,gob}")
	outputFolder  = flag.String("o", defaultOutputFolder, "Folder for output.")
	verbose       = flag.Bool("v", false, "Verbose output")

	name        = flag.String("n", defaultName, "Name of the data set")
	description = flag.String("d", defaultDescription, "Description of the data set")
	id          = flag.Int64("id", defaultID, "ID of the data set")
	teamID      = flag.Int64("tid", defaultTeamID, "Team ID of the data set")
)

func main() {
	flag.Parse()

	if *inputFile == "" {
		log.Print("You must at least specify -i parameter")
		flag.Usage()
		os.Exit(1)
	}

	start := time.Now()

	featureCollection, err := file.ReadGeoJSONFeatureCollection(".", *inputFile, file.GeoJSON)

	if *verbose {
		log.Printf("Loaded FeatureCollection with %d features", len(featureCollection.Features))
	}

	if err != nil {
		log.Fatalf("Error while reading '%s': %v", *inputFile, err)
	}

	mapProjection := gj.NewMapProjectionFromString(*mapProjection)
	shapes, err := triangulation.TriangulateGeoJSONFeatureCollectionToShapes(&featureCollection, mapProjection)

	if *verbose {
		log.Printf("Found %d shape(s)", len(shapes))
	}

	if err != nil {
		if err, ok := err.(triangulation.Error); ok {
			log.Printf("Failed on polygon %#v", err.PolygonWhenError)
		}
		log.Fatalf("Error while triangulating '%s': %v", *inputFile, err)
	}

	stop := time.Now()

	if *outputType == "geojson" {
		path := *outputFolder
		if !strings.HasSuffix(path, "/") {
			path += "/"
		}
		jsonBytes, err := gj.NewGeoJSONFeatureCollectionFromShapes(shapes, true).MarshalJSON()

		if err != nil {
			log.Fatal(err)
		}

		err = ioutil.WriteFile(path+filepath.Base(*inputFile), jsonBytes, 0644)
		if err != nil {
			log.Fatal(err)
		}

		if *verbose {
			log.Printf("Wrote '%s'", *inputFile)
		}
	} else {
		shapeCollectionFile := file.ShapeCollectionFile{
			ShapeCollection: model.ShapeCollection{
				Name:        *name,
				Description: *description,
				ID:          *id,
				TeamID:      *teamID,
			},
			ShapeStore: geometry.ShapesToShapeStore(shapes),
		}

		err = file.SaveShapeCollectionFile(*outputFolder, filepath.Base(*inputFile+".gob"), shapeCollectionFile)

		if err != nil {
			log.Fatalf("Error when trying to save shape collection file: %v", err)
		}
	}

	if !*verbose {
		return
	}

	for _, shape := range shapes {
		switch shape := shape.(type) {
		case *geometry.Polygon:
			polygon := shape
			log.Printf(" - Polygon '%s', %d triangles", polygon.GetName(), len(polygon.Triangles))

		case *geometry.Circle:
			circle := shape
			log.Printf(" - Circle '%s', r = %f", circle.GetName(), circle.Radius)

		default:
			log.Printf("Warning: unknown shape: %+v", shape)
		}
	}
	log.Printf("Triangulation took %s", stop.Sub(start))
}
