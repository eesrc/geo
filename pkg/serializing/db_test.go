package serializing

import (
	"encoding/json"
	"testing"

	log "github.com/sirupsen/logrus"

	geojson "github.com/paulmach/go.geojson"
)

func BenchmarkScanJSONGeoJSON(b *testing.B) {
	jsonBytes := getJSONBytes(getTestGeoJSONFeatureCollection())

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var featureCollection *geojson.FeatureCollection
		err := ScanJSON(&featureCollection, jsonBytes)

		if err != nil {
			log.Fatal(err)
		}
	}
}

func BenchmarkScanJSONShapeProperties(b *testing.B) {
	jsonBytes := getJSONBytes(getShapeProperties())

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var shapeProperties map[string]string
		err := ScanJSON(&shapeProperties, jsonBytes)

		if err != nil {
			log.Fatal(err)
		}
	}
}

func BenchmarkValueJSONGeoJSON(b *testing.B) {
	featureCollection := getTestGeoJSONFeatureCollection()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := ValueJSON(featureCollection)
		if err != nil {
			log.Fatal(err)
		}
	}
}

func BenchmarkValueJSONShapeProperties(b *testing.B) {
	shapeProperties := getShapeProperties()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := ValueJSON(shapeProperties)
		if err != nil {
			log.Fatal(err)
		}
	}
}

func getJSONBytes(obj interface{}) []byte {
	bytes, err := json.Marshal(obj)

	if err != nil {
		log.Fatal(err)
	}

	return bytes
}

func getTestGeoJSONFeatureCollection() *geojson.FeatureCollection {
	rawGeoJSON := `{
		"type": "FeatureCollection",
		"features": [
		  {
			"type": "Feature",
			"properties": {
			  "stroke": "#555555",
			  "stroke-width": 2,
			  "stroke-opacity": 1,
			  "fill": "#555555",
			  "fill-opacity": 0.5,
			  "name": "Trondheim"
			},
			"geometry": {
			  "type": "Polygon",
			  "coordinates": [
				[
				  [10.37109375, 63.47259824039697],
				  [10.0909423828125, 63.36444523158069],
				  [10.3656005859375, 63.208878159899314],
				  [10.72265625, 63.27318217465046],
				  [10.755615234375, 63.38167869302983],
				  [10.61279296875, 63.44805382948824],
				  [10.37109375, 63.47259824039697]
				]
			  ]
			}
		  },
		  {
			"type": "Feature",
			"properties": {
			  "stroke": "#555555",
			  "stroke-width": 2,
			  "stroke-opacity": 1,
			  "fill": "#555555",
			  "fill-opacity": 0.5,
			  "name": "Stjørdal"
			},
			"geometry": {
			  "type": "Polygon",
			  "coordinates": [
				[
				  [10.8270263671875, 63.54855223203644],
				  [10.8050537109375, 63.45296439593958],
				  [10.9918212890625, 63.376755901872734],
				  [11.3214111328125, 63.484862553039946],
				  [11.18408203125, 63.58523167747513],
				  [10.997314453125, 63.61209992940288],
				  [10.8270263671875, 63.54855223203644]
				]
			  ]
			}
		  },
		  {
			"type": "Feature",
			"properties": {
			  "stroke": "#555555",
			  "stroke-width": 2,
			  "stroke-opacity": 1,
			  "fill": "#555555",
			  "fill-opacity": 0.5,
			  "name": "Trøndelag"
			},
			"geometry": {
			  "type": "Polygon",
			  "coordinates": [
				[
				  [11.05224609375, 63.968729627090354],
				  [8.399047851562498, 62.97020521809478],
				  [10.008544921875, 62.49710157662214],
				  [12.2552490234375, 62.36235277748268],
				  [12.689208984375, 63.42594585479083],
				  [11.458740234375, 63.20144925272788],
				  [11.79931640625, 63.553445554178374],
				  [11.722412109375, 63.84551155612745],
				  [11.05224609375, 63.968729627090354]
				]
			  ]
			}
		  }
		]
	  }
	  `

	var featureCollection *geojson.FeatureCollection
	err := json.Unmarshal([]byte(rawGeoJSON), &featureCollection)

	if err != nil {
		log.Panic("JSON Unmarshall", err)
	}

	return featureCollection
}

func getShapeProperties() map[string]string {
	return map[string]string{
		"test1": "value",
		"test2": "value",
		"test3": "value",
		"test4": "value",
		"test5": "value",
		"test6": "value",
		"test7": "value",
		"test8": "value",
	}
}
