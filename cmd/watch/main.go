package main

import (
	log "github.com/sirupsen/logrus"

	"github.com/eesrc/geo/pkg/tria/file"
)

func main() {
	watchEventChan, err := file.WatchIndexFolder("test")

	if err != nil {
		log.Error("Something went wrong when trying to watch folder:", err)
	}

	for watchEvent := range watchEventChan {
		shapeCollectionFile := watchEvent.ShapeCollectionFile

		log.Info(watchEvent.Event, watchEvent.Path)

		if file.Updated == watchEvent.Event {
			shapeCollection := shapeCollectionFile.ShapeCollection
			shapeStore := shapeCollectionFile.ShapeStore
			log.Printf("name: %s, desc: %s, ID: %d, TeamID: %d", shapeCollection.Name, shapeCollection.Description, shapeCollection.ID, shapeCollection.TeamID)
			log.Printf("Circles: %d, Polygons: %d", len(shapeStore.Circles), len(shapeStore.Polygons))
		}
	}

}
