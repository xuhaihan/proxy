package storage

import (
	"encoding/json"
	"errors"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"

	"github.com/boltdb/bolt"
)

type BoltDbStorage struct {
	Db         *bolt.DB
	bucketName string
	contents   sync.Map
	count      int32
}

// NewBoltDbStorage will return a boltdb object and error.
func NewBoltDbStorage(fileName string, bucketName string) (*BoltDbStorage, error) {

	if fileName == "" {
		return nil, errors.New("open boltdb whose fileName is empty")
	}

	if bucketName == "" {
		return nil, errors.New("create a bucket whose name is empty")
	}

	db, err := bolt.Open(fileName, 0600, nil)
	if err != nil {
		return nil, err
	}

	err = db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(bucketName))
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	storage := &BoltDbStorage{
		Db:         db,
		bucketName: bucketName,
	}

	// Sync data from database to memory.
	storage.sync()
	return storage, nil
}

// Exist will check the given key is existed in DB or not.
func (s *BoltDbStorage) Exist(key string) bool {
	return s.Get(key) != nil
}

// Get will get the json byte value of key.
func (s *BoltDbStorage) Get(key string) []byte {
	var value []byte

	if temp, ok := s.contents.Load(key); ok {
		if content, ok := temp.([]byte); ok {
			value = append(value, content...)
		}
	}

	return value
}

// Delete the value by the given key.
func (s *BoltDbStorage) Delete(key string) bool {
	isSucceed := false
	err := s.Db.Update(func(tx *bolt.Tx) error {
		return tx.Bucket([]byte(s.bucketName)).Delete([]byte(key))
	})

	if err == nil {
		isSucceed = true
		if _, ok := s.contents.Load(key); ok {
			s.contents.Delete(key)
			atomic.AddInt32(&s.count, -1)
		}
	}

	return isSucceed
}

// AddOrUpdate will add the value into DB if key is not existed, otherwise update the existing value.
// Null value will be ignored and the value will be marshal as json format.
func (s *BoltDbStorage) AddOrUpdate(key string, value interface{}) error {
	if value == nil {
		return errors.New("value is null")
	}

	content, err := json.Marshal(value)
	if err != nil {
		return err
	}

	err = s.Db.Update(func(tx *bolt.Tx) error {
		return tx.Bucket([]byte(s.bucketName)).Put([]byte(key), content)
	})

	if err == nil {
		if _, loaded := s.contents.LoadOrStore(key, content); !loaded {
			atomic.AddInt32(&s.count, 1)
		}
	}

	return err
}

// GetAll will return all key-value in DB.
func (s *BoltDbStorage) GetAll() map[string][]byte {
	result := make(map[string][]byte)

	s.contents.Range(func(key, value interface{}) bool {
		if k, ok := key.(string); ok {
			if v, ok := value.([]byte); ok {
				result[k] = v
			}
		}

		return true
	})

	return result
}

// Close will close the DB.
func (s *BoltDbStorage) Close() {
	s.Db.Close()
}

// Sync will sync the DB's data to memory.
func (s *BoltDbStorage) sync() {
	s.Db.View(func(tx *bolt.Tx) error {
		tx.Bucket([]byte(s.bucketName)).ForEach(func(k, v []byte) error {
			key, value := make([]byte, len(k)), make([]byte, len(v))
			copy(key, k)
			copy(value, v)
			s.contents.Store(string(key), value)
			atomic.AddInt32(&s.count, 1)
			return nil
		})

		return nil
	})

	// seelog.Debugf("content:%v,count:%d", s.contents, s.count)
}

// Get one random record.
func (s *BoltDbStorage) GetRandomOne() []byte {
	if s.count == 0 {
		return nil
	}

	var randomKey string
	var defaultKey string
	index := rand.New(rand.NewSource(time.Now().Unix())).Intn(int(atomic.LoadInt32(&s.count)))

	s.contents.Range(func(key, value interface{}) bool {
		// Set a default key to avoid that other goroutine is deleting content at the same time.
		if defaultKey == "" {
			defaultKey, _ = key.(string)
		}

		if index == 0 {
			randomKey, _ = key.(string)
			return false
		}

		index--
		return true
	})

	if randomKey == "" {
		randomKey = defaultKey
	}

	return s.Get(randomKey)
}
