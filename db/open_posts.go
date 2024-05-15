package db

import (
	"database/sql"
	"encoding/binary"
	"encoding/json"
	"github.com/bakape/meguca/common"
	"github.com/bakape/meguca/pb"
	"github.com/go-playground/log"
	"github.com/linxGnu/grocksdb"
	"google.golang.org/protobuf/proto"
	"sync"
	"sync/atomic"
)

const (
	rocksDBNotOpened = iota // Not opened yet in this server instance
	rocksDBOpen             // Opened and ready for operation
	rocksDBClosed           // Closed for graceful restart
)

var (
	// Current state of RocksDB database.
	// Should only be accessed using atomic operations.
	rocksDBState uint32

	// Ensures RocksDB is opened only once
	rocksDBOnce  sync.Once
	writeOptions = grocksdb.NewDefaultWriteOptions()
	readOptions  = grocksdb.NewDefaultReadOptions()

	// Column family handles
	openBodyCF    *grocksdb.ColumnFamilyHandle
	nekoTVStateCF *grocksdb.ColumnFamilyHandle
)

// Close DB and release resources
func Close() (err error) {
	if atomic.LoadUint32(&rocksDBState) == rocksDBOpen {
		atomic.StoreUint32(&rocksDBState, rocksDBClosed)
		openBodyCF.Destroy()
		nekoTVStateCF.Destroy()
		rocksDB.Close()
	}
	return
}

// Need to drop any incoming requests, when DB is closed during graceful restart
func rocksDBisOpen() bool {
	return atomic.LoadUint32(&rocksDBState) == rocksDBOpen
}

// Open RocksDB, only when needed. This helps prevent conflicts on swapping
// the database accessing process during graceful restarts.
// If RocksDB has already been closed, return open=false.
func tryOpenRocksDB() (open bool, err error) {
	rocksDBOnce.Do(func() {
		opts := grocksdb.NewDefaultOptions()
		opts.OptimizeForPointLookup(64)
		opts.SetCreateIfMissing(true)
		opts.SetCompression(grocksdb.NoCompression)
		opts.SetCreateIfMissingColumnFamilies(true)

		// Define column family names
		cfNames := []string{"default", "open_body", "nekotv_state"}

		// Open the DB with column families
		db, cfHandles, err := grocksdb.OpenDbColumnFamilies(opts, "rdb.db", cfNames, []*grocksdb.Options{opts, opts, opts})
		if err != nil {
			return
		}
		rocksDB = db
		openBodyCF = cfHandles[1]
		nekoTVStateCF = cfHandles[2]

		atomic.StoreUint32(&rocksDBState, rocksDBOpen)
		return
	})
	if err != nil {
		return
	}

	open = rocksDBisOpen()
	return
}

// SetOpenBody sets the open body of a post
func SetOpenBody(id uint64, body []byte) (err error) {
	ok, err := tryOpenRocksDB()
	if err != nil || !ok {
		return
	}

	buf := encodeUint64(id)
	err = rocksDB.PutCF(writeOptions, openBodyCF, buf[:], body)
	return
}

// Encode uint64 for storage in RocksDB without heap allocations
func encodeUint64(i uint64) [8]byte {
	var buf [8]byte
	binary.LittleEndian.PutUint64(buf[:], i)
	return buf
}

// Same as encodeUint64, but allocates on the heap. In some cases, where the
// buffer must persist after the end of the transaction, this is needed.
func encodeUint64Heap(i uint64) []byte {
	buf := make([]byte, 8)
	binary.LittleEndian.PutUint64(buf, i)
	return buf
}

// GetOpenBody retrieves an open body of a post
func GetOpenBody(id uint64) (body string, err error) {
	ok, err := tryOpenRocksDB()
	if err != nil || !ok {
		return
	}

	buf := encodeUint64(id)
	bodySlice, err := rocksDB.GetCF(readOptions, openBodyCF, buf[:])
	if err != nil {
		return
	}
	defer bodySlice.Free()
	body = string(bodySlice.Data())
	return
}

func deleteOpenPostBody(id uint64) (err error) {
	ok, err := tryOpenRocksDB()
	if err != nil || !ok {
		return
	}

	buf := encodeUint64(id)
	return rocksDB.DeleteCF(writeOptions, openBodyCF, buf[:])
}

// Inject open post bodies from the embedded database into the posts
func injectOpenBodies(posts []*common.Post) (err error) {
	if len(posts) == 0 {
		return
	}

	ok, err := tryOpenRocksDB()
	if err != nil || !ok {
		return
	}

	for _, p := range posts {
		p.Body, err = GetOpenBody(p.ID)
		if err != nil {
			return
		}
	}
	return
}

// Delete orphaned post bodies, that refer to posts already closed or deleted.
// This can happen on server restarts, board deletion, etc.
func cleanUpOpenPostBodies() (err error) {
	ok, err := tryOpenRocksDB()
	if err != nil || !ok {
		return
	}

	// Read IDs of all post bodies
	var ids []uint64
	it := rocksDB.NewIteratorCF(readOptions, openBodyCF)
	defer it.Close()
	for it.SeekToFirst(); it.Valid(); it.Next() {
		// Get the key
		key := it.Key()

		// Convert the key to uint64
		keyUint64 := binary.LittleEndian.Uint64(key.Data())
		key.Free()

		// Append the key to the keys slice
		ids = append(ids, keyUint64)
	}

	if err = it.Err(); err != nil {
		return
	}

	// Find bodies with closed parents
	toDelete := make([]uint64, 0, len(ids))
	return InTransaction(true, func(tx *sql.Tx) (err error) {
		var isOpen bool
		q, err := tx.Prepare(`select 'true' from posts
			where id = $1 and editing = 'true'`)
		if err != nil {
			return
		}
		for _, id := range ids {
			err = q.QueryRow(id).Scan(&isOpen)
			switch err {
			case nil:
			case sql.ErrNoRows:
				err = nil
				isOpen = false // Treat missing as closed
			default:
				return
			}
			if !isOpen {
				toDelete = append(toDelete, id)
			}
		}
		err = q.Close()
		if err != nil {
			return err
		}

		// Delete closed post bodies, if any
		if len(toDelete) == 0 {
			return
		}
		for _, id := range toDelete {
			err = deleteOpenPostBody(id)
			if err != nil {
				return
			}
		}
		return
	})
}

func GetNekoTVState(id uint64) (state pb.ServerState, err error) {
	ok, err := tryOpenRocksDB()
	if err != nil || !ok {
		return
	}

	buf := encodeUint64(id)
	value, err := rocksDB.GetCF(readOptions, nekoTVStateCF, buf[:])
	if err != nil {
		return
	}
	defer value.Free()

	err = proto.Unmarshal(value.Data(), &state)
	jso, _ := json.Marshal(state)
	log.Info("Get: ", string(jso))
	return
}

func SetNekoTVState(id uint64, state *pb.ServerState) (err error) {
	jso, _ := json.Marshal(state)
	ok, err := tryOpenRocksDB()
	if err != nil || !ok {
		return
	}

	buf := encodeUint64(id)
	value, err := proto.Marshal(state)
	if err != nil {
		return
	}

	err = rocksDB.PutCF(writeOptions, nekoTVStateCF, buf[:], value)
	log.Info("Set: ", string(jso))
	return
}

// DeleteNekoTVValue deletes a value for NekoTV
func DeleteNekoTVValue(id uint64) (err error) {
	ok, err := tryOpenRocksDB()
	if err != nil || !ok {
		return
	}

	buf := encodeUint64(id)
	err = rocksDB.DeleteCF(writeOptions, nekoTVStateCF, buf[:])
	return
}

// cleanUpNekoTVValues deletes orphaned NekoTV values
func cleanUpNekoTVValues() (err error) {
	ok, err := tryOpenRocksDB()
	if err != nil || !ok {
		return
	}

	// Read IDs of all NekoTV values
	var ids []uint64
	it := rocksDB.NewIteratorCF(readOptions, nekoTVStateCF)
	defer it.Close()
	for it.SeekToFirst(); it.Valid(); it.Next() {
		// Get the key
		key := it.Key()

		// Convert the key to uint64
		keyUint64 := binary.LittleEndian.Uint64(key.Data())
		key.Free()

		// Append the key to the keys slice
		ids = append(ids, keyUint64)
	}

	if err = it.Err(); err != nil {
		return
	}

	// Find values with closed parents
	toDelete := make([]uint64, 0, len(ids))
	return InTransaction(true, func(tx *sql.Tx) (err error) {
		var isOpen bool
		q, err := tx.Prepare(`select 'true' from posts
			where id = $1 and editing = 'true'`)
		if err != nil {
			return
		}
		for _, id := range ids {
			err = q.QueryRow(id).Scan(&isOpen)
			switch err {
			case nil:
			case sql.ErrNoRows:
				err = nil
				isOpen = false // Treat missing as closed
			default:
				return
			}
			if !isOpen {
				toDelete = append(toDelete, id)
			}
		}
		err = q.Close()
		if err != nil {
			return err
		}

		// Delete closed NekoTV values, if any
		if len(toDelete) == 0 {
			return
		}
		for _, id := range toDelete {
			err = DeleteNekoTVValue(id)
			if err != nil {
				return
			}
		}
		return
	})
}
