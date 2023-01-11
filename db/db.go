package db

import (
	"errors"
	"os"
	"strings"

	"github.com/fantasticake/fantasticoin/utils"
	"go.etcd.io/bbolt"
)

var (
	db            *bbolt.DB
	dbName        = "database.db"
	blocksBucket  = "blocksBucket"
	dataBucket    = "dataBucket"
	checkpointKey = "checkpointKey"
)

func DB() *bbolt.DB {
	if db == nil {
		port := strings.Split(os.Args[1], "=")[1]
		database, err := bbolt.Open(port+dbName, 0600, nil)
		db = database
		utils.HandleErr(err)
		err = db.Update(func(tx *bbolt.Tx) error {
			_, err = tx.CreateBucketIfNotExists([]byte(blocksBucket))
			if err != nil {
				return err
			}
			_, err = tx.CreateBucketIfNotExists([]byte(dataBucket))
			return err
		})
		utils.HandleErr(err)
	}
	return db
}

func Close() {
	DB().Close()
}

func SaveCheckpoint(data []byte) {
	err := DB().Update(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket([]byte(dataBucket))
		err := bucket.Put([]byte(checkpointKey), data)
		return err
	})
	utils.HandleErr(err)
}

func SaveBlock(key []byte, data []byte) {
	err := DB().Update(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket([]byte(blocksBucket))
		err := bucket.Put(key, data)
		return err
	})
	utils.HandleErr(err)
}

func GetCheckpoint() []byte {
	var data []byte
	DB().View(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket([]byte(dataBucket))
		data = bucket.Get([]byte(checkpointKey))
		return nil
	})
	return data
}

func FindBlock(key []byte) ([]byte, error) {
	var data []byte
	DB().View(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket([]byte(blocksBucket))
		data = bucket.Get(key)
		return nil
	})
	if data == nil {
		return nil, errors.New("Not found")
	}
	return data, nil
}
