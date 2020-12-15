package storage

import (
	"container/list"
	"fmt"
	"sync"

	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"
	pb "github.com/samirettali/go-pastebin"

	_ "github.com/jinzhu/gorm/dialects/postgres"
)

// PgStorage is an implementation of the Storage interface
type PgStorage struct {
	// URI      string
	// Database string
	// Table    string

	db    *gorm.DB
	mutex sync.Mutex
	cache *list.List
}

// Init initializes the collection pointer
func (s *PgStorage) Init() error {
	var err error
	db, err := gorm.Open("postgres", "host=localhost port=5432 user=postgres dbname=pastebin sslmode=disable password=postgres")
	if err != nil {
		return err
	}
	db.AutoMigrate(&pb.Paste{})
	s.db = db
	s.cache = list.New()
	return nil
}

// IsSaved checks if the paste is already saved
func (s *PgStorage) IsSaved(key string) (bool, error) {
	if s.isInCache(key) {
		return true, nil
	}
	var paste pb.Paste
	s.db.Where("Key = ?", key).First(&paste)
	return paste.Key != "", nil
}

// Save saves a paste
func (s *PgStorage) Save(paste pb.Paste) error {
	result := s.db.Create(&paste)
	if result.Error != nil {
		msg := fmt.Sprintf("Can't insert this paste: %s", paste.Key)
		return errors.Wrap(result.Error, msg)
	}
	s.addToCache(paste.Key)
	return nil
}

func (s *PgStorage) addToCache(key string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.cache.PushBack(key)
	if s.cache.Len() > 250 {
		e := s.cache.Front()
		s.cache.Remove(e)
	}
}

func (s *PgStorage) isInCache(key string) bool {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	for c := s.cache.Front(); c != nil; c = c.Next() {
		if c.Value == key {
			return true
		}
	}
	return false
}

func (s *PgStorage) exit() error {
	return s.db.Close()
}
