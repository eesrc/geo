package file

import (
	"os"
	"path/filepath"
	"strings"
	"sync"

	log "github.com/sirupsen/logrus"

	"github.com/fsnotify/fsnotify"
)

func folderFunc(pathX string, infoX os.FileInfo, errX error) error {
	log.Info(pathX)
	return nil
}

// WatchEventType is the type of event received when a file has changed
type WatchEventType string

const (
	// Updated means a file has been created or updated
	Updated WatchEventType = "updated"
	// Removed means a file has been removed
	Removed WatchEventType = "removed"
)

// WatchEvent is the event you receive when a file has changed
type WatchEvent struct {
	// Event is which kind of watch event type
	Event WatchEventType
	// Path is the path of the file
	Path string
	// ShapeCollection is the ShapeCollection for the event
	ShapeCollectionFile *ShapeCollectionFile
}

// WatchIndexFolder will watch given folder and return a channel which will publish
// a WatchEvent upon file changes for .gob files changed and if they contain a ShapeCollection
func WatchIndexFolder(path string) (chan WatchEvent, error) {
	watchChan := make(chan WatchEvent)
	fileChan := make(chan string)

	var fileHandler = fileWatcher{
		mutex:        &sync.Mutex{},
		fileChannels: map[string]chan string{},
		fileChan:     fileChan,
	}

	err := filepath.Walk(path, folderFunc)
	if err != nil {
		return watchChan, err
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return watchChan, err
	}

	// Listen to fsnotify events on given folder
	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}

				if event.Op&fsnotify.Create == fsnotify.Create {
					fileHandler.FileUpdated(event.Name)
				}

				if event.Op&fsnotify.Write == fsnotify.Write {
					fileHandler.FileUpdated(event.Name)
					break
				}

				if event.Op&fsnotify.Remove == fsnotify.Remove {
					ext := filepath.Ext(event.Name)

					if strings.ToLower(ext) == ".gob" {
						watchChan <- WatchEvent{
							Event: Removed,
							Path:  event.Name,
						}
					}
					break
				}

			case _, ok := <-watcher.Errors:
				if !ok {
					return
				}
			}
		}
	}()

	// Listen to files who has changed and published. Load shape collection if gob, otherwise ignore
	go func() {
		for filePath := range fileChan {
			ext := filepath.Ext(filePath)

			if strings.ToLower(ext) == ".gob" {
				shapeCollectionFile, err := LoadShapeCollectionFile(".", filePath)

				if err != nil {
					log.Error("Error on loading shape", err)
				} else {
					watchChan <- WatchEvent{
						Event:               Updated,
						Path:                filePath,
						ShapeCollectionFile: shapeCollectionFile,
					}
				}
			}
		}
	}()

	err = watcher.Add(path)

	if err != nil {
		return watchChan, err
	}

	return watchChan, nil
}
