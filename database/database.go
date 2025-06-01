
package database

import (
	"github.com/ethereum/go-ethereum/ethdb"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/errors"
	"github.com/syndtr/goleveldb/leveldb/filter"
	"github.com/syndtr/goleveldb/leveldb/opt"
	"github.com/syndtr/goleveldb/leveldb/util"
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
	iter := w.db.db.NewIterator(util.BytesPrefix(prefix), nil)
	if start != nil {
		iter.Seek(start)
	}
	return &IteratorWrapper{iter: iter}
}

func (w *EthDBWrapper) Stat(property string) (string, error) {
	return "", nil
}

func (w *EthDBWrapper) Compact(start []byte, limit []byte) error {
	return w.db.db.CompactRange(util.Range{Start: start, Limit: limit})
}

func (w *EthDBWrapper) Close() error {
	return w.db.Close()
}

func (w *EthDBWrapper) Ancient(kind string, number uint64) ([]byte, error) {
	return nil, ethdb.ErrNotFound
}

func (w *EthDBWrapper) AncientDatadir() (string, error) {
	return "", nil
}

func (w *EthDBWrapper) AncientRange(kind string, start, count, maxBytes uint64) ([][]byte, error) {
	return [][]byte{}, nil
}

func (w *EthDBWrapper) AncientSize(kind string) (uint64, error) {
	return 0, nil
}

func (w *EthDBWrapper) HasAncient(kind string, number uint64) (bool, error) {
	return false, nil
}

func (w *EthDBWrapper) ModifyAncients(fn func(ethdb.AncientWriteOp) error) (int64, error) {
	return 0, nil
}

func (w *EthDBWrapper) ReadAncients(fn func(ethdb.AncientReaderOp) error) (err error) {
	return nil
}

func (w *EthDBWrapper) TruncateHead(n uint64) error {
	return nil
}

func (w *EthDBWrapper) TruncateTail(n uint64) error {
	return nil
}

func (w *EthDBWrapper) Sync() error {
	return nil
}

func (w *EthDBWrapper) MigrateTable(s string, f func([]byte) ([]byte, error)) error {
	return nil
}

func (w *EthDBWrapper) NewSnapshot() (ethdb.Snapshot, error) {
	return &SnapshotWrapper{}, nil
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
type IteratorWrapper struct {
	iter *leveldb.Iterator
}

func (i *IteratorWrapper) Next() bool {
	return i.iter.Next()
}

func (i *IteratorWrapper) Error() error {
	return i.iter.Error()
}

func (i *IteratorWrapper) Key() []byte {
	return i.iter.Key()
}

func (i *IteratorWrapper) Value() []byte {
	return i.iter.Value()
}

func (i *IteratorWrapper) Release() {
	i.iter.Release()
}

// SnapshotWrapper implements ethdb.Snapshot
type SnapshotWrapper struct{}

func (s *SnapshotWrapper) Has(key []byte) (bool, error) {
	return false, nil
}

func (s *SnapshotWrapper) Get(key []byte) ([]byte, error) {
	return nil, ethdb.ErrNotFound
}

func (s *SnapshotWrapper) Release() {}
