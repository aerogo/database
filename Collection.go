package database

import (
	"bufio"
	"encoding/json"
	"os"
	"path"
	"runtime"
	"sort"
	"sync"
	"time"
)

// Collection ...
type Collection struct {
	data  sync.Map
	db    *Database
	name  string
	dirty chan bool
}

// NewCollection ...
func NewCollection(db *Database, name string) *Collection {
	collection := &Collection{
		db:    db,
		name:  name,
		dirty: make(chan bool, runtime.NumCPU()),
	}

	go func() {
		for {
			<-collection.dirty

			for len(collection.dirty) > 0 {
				<-collection.dirty
			}

			collection.flush()
			time.Sleep(db.ioSleepTime)
		}
	}()

	return collection
}

// Get ...
func (collection *Collection) Get(key string) interface{} {
	val, _ := collection.data.Load(key)
	return val
}

// Set ...
func (collection *Collection) Set(key string, value interface{}) {
	collection.data.Store(key, value)

	// The potential data race here does not matter at all.
	if len(collection.dirty) == 0 {
		collection.dirty <- true
	}
}

// Delete ...
func (collection *Collection) Delete(key string) {
	collection.data.Delete(key)

	// The potential data race here does not matter at all.
	if len(collection.dirty) == 0 {
		collection.dirty <- true
	}
}

// Clear deletes all objects from the collection.
func (collection *Collection) Clear() {
	collection.data = sync.Map{}
	runtime.GC()

	// The potential data race here does not matter at all.
	if len(collection.dirty) == 0 {
		collection.dirty <- true
	}
}

// Exists ...
func (collection *Collection) Exists(key string) bool {
	_, exists := collection.data.Load(key)
	return exists
}

// All ...
func (collection *Collection) All() chan interface{} {
	channel := make(chan interface{}, 128)

	go allValues(&collection.data, channel)

	return channel
}

// flush writes all data to the file system.
func (collection *Collection) flush() {
	file, err := os.OpenFile(path.Join(collection.db.root, collection.name+".dat"), os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0666)

	if err != nil {
		panic(err)
	}

	file.Seek(0, 0)
	bufferedWriter := bufio.NewWriter(file)

	records := []keyValue{}

	collection.data.Range(func(key, value interface{}) bool {
		records = append(records, keyValue{
			key:   key.(string),
			value: value,
		})
		return true
	})

	sort.Slice(records, func(i, j int) bool {
		return records[i].key < records[j].key
	})

	for _, record := range records {
		valueBytes, err := json.Marshal(record.value)

		if err != nil {
			panic(err)
		}

		bufferedWriter.WriteString(record.key)
		bufferedWriter.WriteByte('\n')

		bufferedWriter.Write(valueBytes)
		bufferedWriter.WriteByte('\n')
	}

	bufferedWriter.Flush()
	file.Close()
}

// allValues iterates over all values in a sync.Map and sends them to the given channel.
func allValues(data *sync.Map, channel chan interface{}) {
	data.Range(func(key, value interface{}) bool {
		channel <- value
		return true
	})

	close(channel)
}
