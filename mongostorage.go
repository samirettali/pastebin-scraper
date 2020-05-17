package main

import (
	"container/list"
	"context"
	"sync"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// MongoStorage is an implementation of the Storage interface
type MongoStorage struct {
	URI        string
	Database   string
	Collection string

	col   *mongo.Collection
	mutex sync.Mutex
	cache *list.List
}

// Init initializes the collection pointer
func (s *MongoStorage) Init() error {
	var err error
	client, err := mongo.NewClient(options.Client().ApplyURI(s.URI))
	if err != nil {
		return err
	}
	if err = client.Connect(context.Background()); err != nil {
		return err
	}
	db := client.Database(s.Database)
	s.col = db.Collection(s.Collection)
	s.cache = list.New()
	return nil
}

// IsSaved checks if the paste is already saved
func (s *MongoStorage) IsSaved(key string) (bool, error) {
	if s.isInCache(key) {
		return true, nil
	}
	paste := Paste{}
	filter := &bson.M{
		"key": key,
	}
	err := s.col.FindOne(context.Background(), filter).Decode(&paste)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// Save saves a paste
func (s *MongoStorage) Save(paste Paste) error {
	s.addToCache(paste.Key)
	_, err := s.col.InsertOne(context.Background(), paste)
	return err
}

func (s *MongoStorage) addToCache(key string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.cache.PushBack(key)
	if s.cache.Len() > 250 {
		e := s.cache.Front()
		s.cache.Remove(e)
	}
}

func (s *MongoStorage) isInCache(key string) bool {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	for c := s.cache.Front(); c != nil; c = c.Next() {
		if c.Value == key {
			return true
		}
	}
	return false
}
