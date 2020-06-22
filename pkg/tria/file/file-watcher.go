package file

import (
	"sync"
	"time"
)

const debounceTimeMS = 750

// Simple helper to help watch when files has been written completely
type fileWatcher struct {
	mutex        *sync.Mutex
	fileChannels map[string]chan string
	fileChan     chan string
}

func (handler *fileWatcher) FileUpdated(key string) {
	handler.mutex.Lock()
	defer handler.mutex.Unlock()

	// Publish to file channel if exists
	if handler.fileChannels[key] != nil {
		go func() { handler.fileChannels[key] <- key }()
		return
	}

	// No file channel found, create and subscribe for changes
	// We do this as files are written sequentially and we are
	// not sure wheter a file is done writing.
	handler.fileChannels[key] = make(chan string)

	go func() {
		debounceTime := debounceTimeMS * time.Millisecond
		timer := time.NewTimer(debounceTime)

		for {
			select {
			case <-handler.fileChannels[key]:
				timer.Reset(debounceTime)
			case <-timer.C:
				handler.fileChan <- key

				close(handler.fileChannels[key])
				delete(handler.fileChannels, key)
			}
		}
	}()
}
