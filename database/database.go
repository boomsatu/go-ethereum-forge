
package database

import (
	"github.com/ethereum/go-ethereum/ethdb"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/errors"
	"github.com/syndtr/goleveldb/leveldb/filter"
	"github.com/syndtr/goleveldb/leveldb/opt"
)

type Database interface {
	Get(key []byte) ([]byte, error)
	Put(key []byte, value []byte) error
	Delete(key []byte) error
	Close() error
	GetEthDB() ethdb.Database
}

type LevelDB struct {
	db *leveldb.DB
}

func NewLevelDB(path string) (*LevelDB, error) {
	opts := &opt.Options{
		Filter: filter.NewBloomFilter(10),
	}
	
	db, err := leveldb.OpenFile(path, opts)
	if err != nil {
		if errors.IsCorrupted(err) {
			db, err = leveldb.RecoverFile(path, nil)
		}
		if err != nil {
			return nil, err
		}
	}
	
	return &LevelDB{db: db}, nil
}

func (ldb *LevelDB) Get(key []byte) ([]byte, error) {
	value, err := ldb.db.Get(key, nil)
	if err == leveldb.ErrNotFound {
		return nil, nil
	}
	return value, err
}

func (ldb *LevelDB) Put(key []byte, value []byte) error {
	return ldb.db.Put(key, value, nil)
}

func (ldb *LevelDB) Delete(key []byte) error {
	return ldb.db.Delete(key, nil)
}

func (ldb *LevelDB) Close() error {
	return ldb.db.Close()
}

func (ldb *LevelDB) GetEthDB() ethdb.Database {
	return &EthDBWrapper{ldb}
}

// EthDBWrapper wraps our database to implement ethdb.Database interface
type EthDBWrapper struct {
	db *LevelDB
}

func (w *EthDBWrapper) Has(key []byte) (bool, error) {
	val, err := w.db.Get(key)
	if err != nil {
		return false, err
	}
	return val != nil, nil
}

func (w *EthDBWrapper) Get(key []byte) ([]byte, error) {
	return w.db.Get(key)
}

func (w *EthDBWrapper) Put(key []byte, value []byte) error {
	return w.db.Put(key, value)
}

func (w *EthDBWrapper) Delete(key []byte) error {
	return w.db.Delete(key)
}

func (w *EthDBWrapper) NewBatch() ethdb.Batch {
	return &BatchWrapper{batch: &leveldb.Batch{}, db: w.db}
}

func (w *EthDBWrapper) NewIterator(prefix []byte, start []byte) ethdb.Iterator {
	// Simple implementation - in production, use proper iterator
	return &IteratorWrapper{}
}

func (w *EthDBWrapper) Stat(property string) (string, error) {
	return "", nil
}

func (w *EthDBWrapper) Compact(start []byte, limit []byte) error {
	return nil
}

func (w *EthDBWrapper) Close() error {
	return w.db.Close()
}

// BatchWrapper implements ethdb.Batch
type BatchWrapper struct {
	batch *leveldb.Batch
	db    *LevelDB
}

func (b *BatchWrapper) Put(key, value []byte) error {
	b.batch.Put(key, value)
	return nil
}

func (b *BatchWrapper) Delete(key []byte) error {
	b.batch.Delete(key)
	return nil
}

func (b *BatchWrapper) ValueSize() int {
	return b.batch.Len()
}

func (b *BatchWrapper) Write() error {
	return b.db.db.Write(b.batch, nil)
}

func (b *BatchWrapper) Reset() {
	b.batch.Reset()
}

func (b *BatchWrapper) Replay(w ethdb.KeyValueWriter) error {
	return nil
}

// IteratorWrapper implements ethdb.Iterator
type IteratorWrapper struct{}

func (i *IteratorWrapper) Next() bool    { return false }
func (i *IteratorWrapper) Error() error  { return nil }
func (i *IteratorWrapper) Key() []byte   { return nil }
func (i *IteratorWrapper) Value() []byte { return nil }
func (i *IteratorWrapper) Release()      {}
