package storage

import (
	"container/list"
	"fmt"
	"sync"

	"github.com/pkg/errors"
	pb "github.com/samirettali/go-pastebin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	_ "github.com/jinzhu/gorm/dialects/postgres"
)

type PgConfig struct {
	Host     string `required:"true"`
	Port     string `required:"true"`
	User     string `required:"true"`
	DBName   string `required:"true"`
	Password string `required:"true"`
}

// PgStorage is an implementation of the Storage interface
type PgStorage struct {
	Config *PgConfig

	db    *gorm.DB
	mutex sync.Mutex
	cache *list.List
}

// Init initializes the collection pointer
func (s *PgStorage) Init() error {
	var err error
	dsn := fmt.Sprintf("host=%s port=%s user=%s dbname=%s sslmode=disable password=%s", s.Config.Host, s.Config.Port, s.Config.User, s.Config.DBName, s.Config.Password)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{Logger: logger.Default.LogMode(logger.Error)})
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
