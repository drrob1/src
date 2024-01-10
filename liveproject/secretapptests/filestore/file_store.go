package filestore

import (
	"encoding/json"
	"io"
	"log"
	"os"
	"sync"

	"src/liveproject/secretapp/types"
	//"github.com/amitsaha/manning-go-advanced-beginners/project-series-2/mini-project-1/milestone1-code/types"
)

type fileStore struct {
	Mu    sync.Mutex
	Store map[string]string
}

var FileStoreConfig struct {
	DataFilePath string
	Fs           fileStore
}

func Init(dataFilePath string) error {
	_, err := os.Stat(dataFilePath)

	if err != nil {
		_, err := os.Create(dataFilePath)
		if err != nil {
			return err
		}
	}
	FileStoreConfig.Fs = fileStore{Mu: sync.Mutex{}, Store: make(map[string]string)}
	FileStoreConfig.DataFilePath = dataFilePath
	return nil
}

func (j *fileStore) Write(data types.SecretData) error {
	j.Mu.Lock()
	defer j.Mu.Unlock()

	err := j.ReadFromFile()
	if err != nil {
		return err
	}
	j.Store[data.Id] = data.Secret
	return j.WriteToFile()
}

func (j *fileStore) Read(id string) (string, error) {
	j.Mu.Lock()
	defer j.Mu.Unlock()

	err := j.ReadFromFile()
	if err != nil {
		return "", err
	}

	data := j.Store[id]
	delete(j.Store, id)
	j.WriteToFile()

	return data, nil
}

func (j *fileStore) WriteToFile() error {
	var f *os.File
	jsonData, err := json.Marshal(j.Store)
	if err != nil {
		return err
	}
	f, err = os.Create(FileStoreConfig.DataFilePath)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.Write(jsonData)
	return err
}

func (j *fileStore) ReadFromFile() error {

	f, err := os.Open(FileStoreConfig.DataFilePath)
	if err != nil {
		return err
	}
	jsonData, err := io.ReadAll(f)
	if err != nil {
		log.Fatal(err)
	}
	if len(jsonData) != 0 {
		return json.Unmarshal(jsonData, &j.Store)
	}
	return nil
}
