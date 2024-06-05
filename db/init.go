package db

import (
	"database/sql"
	"fmt"
	"github.com/linxGnu/grocksdb"
	"net/url"
	"os"
	"os/exec"
	"strings"

	"github.com/Masterminds/squirrel"
	"github.com/bakape/meguca/auth"
	"github.com/bakape/meguca/common"
	"github.com/bakape/meguca/config"
	"github.com/bakape/meguca/util"
	"github.com/go-playground/log"
	_ "github.com/lib/pq" // Postgres driver
)

var (
	// ConnArgs specifies the PostgreSQL connection URL
	connectionURL string

	// Stores the postgres database instance
	sqlDB *sql.DB

	// Statement builder and cacher
	sq squirrel.StatementBuilderType

	// Embedded database for temporary storage
	rocksDB *grocksdb.DB
)

// Connects to PostgreSQL database and performs schema upgrades
func LoadDB() error {
	return loadDB(config.Server.Database, "")
}

// Create and load testing database. Call close() to clean up temporary
// resources.
func LoadTestDB(suffix string) (close func() error, err error) {
	common.IsTest = true

	run := func(line ...string) error {
		c := exec.Command(line[0], line[1:]...)
		c.Stdout = os.Stdout
		c.Stderr = os.Stderr
		return c.Run()
	}
	connURL, err := url.Parse(config.Server.Test.Database)
	if err != nil {
		return
	}
	user := connURL.User.Username()
	dbName := fmt.Sprintf("%s_%s", strings.Trim(connURL.Path, "/"), suffix)

	err = run(
		"psql",
		"-c", "drop database if exists "+dbName,
		config.Server.Database,
	)
	if err != nil {
		return
	}

	close = func() (err error) {
		_, err = os.Stat("db.db")
		if err != nil {
			if os.IsNotExist(err) {
				return nil
			}
			return
		}

		rocksDB.Close()
		if err != nil {
			return
		}
		return os.Remove("db.db")
	}

	fmt.Println("creating test database:", dbName)
	err = run(
		"psql",
		"-c",
		fmt.Sprintf(
			"create database %s with owner %s encoding UTF8",
			dbName, user,
		),
		config.Server.Database,
	)
	if err != nil {
		return
	}

	connURL.Path = "/" + dbName
	err = loadDB(connURL.String(), suffix)
	return
}

func loadDB(connURL, dbSuffix string) (err error) {
	// Enable binary parameters for more efficient encoding of []byte
	u, err := url.Parse(connURL)
	if err != nil {
		return
	}
	u.Query().Set("binary_parameters", "yes")
	connURL = u.String()

	// Set, for creating extra connections using Listen()
	connectionURL = connURL

	sqlDB, err = sql.Open("postgres", connURL)
	if err != nil {
		return
	}

	sq = squirrel.StatementBuilder.
		RunWith(squirrel.NewStmtCacheProxy(sqlDB)).
		PlaceholderFormat(squirrel.Dollar)

	var exists bool
	const q = `select exists (
			select 1 from information_schema.tables
				where table_schema = 'public' and table_name = 'main'
		)`
	err = sqlDB.QueryRow(q).Scan(&exists)
	if err != nil {
		return
	}

	tasks := make([]func() error, 0, 16)
	if !exists {
		tasks = append(tasks, initDB)
	} else if err = runMigrations(); err != nil {
		return
	}

	// Run these is parallel
	tasks = append(
		tasks,
		func() error {
			tasks := []func() error{loadConfigs, loadBans, handleSpamScores, prepareUpdatePostsStmt, prepareInsertPostStmt, prepareLinkStatement, prepareInsertImageStmt, PrepareGetGeneralStatement}
			if config.Server.ImagerMode != config.ImagerOnly {
				tasks = append(tasks, loadBanners, loadLoadingAnimations,
					loadThreadPostCounts)
			}
			if err := util.Parallel(tasks...); err != nil {
				return err
			}

			// Depends on loadBanners and loadLoadingAnimations, so has to be
			// sequential
			return loadBoardConfigs()
		},
	)

	err = util.Waterfall(tasks...)
	if err != nil {
		return
	}

	if !common.IsTest {
		go runCleanupTasks()
	}

	return nil
}

// initDB initializes a database
func initDB() (err error) {
	log.Info("initializing database")

	_, err = sqlDB.Exec(
		`create table main (
			id text primary key,
			val text not null
		)`,
	)
	if err != nil {
		return
	}
	_, err = sqlDB.Exec(
		`insert into main (id, val)
		values ('version', '0'),
				('pyu', '0')`,
	)
	if err != nil {
		return
	}
	return runMigrations()
}

// CreateAdminAccount writes a fresh admin account with the default password to
// the database
func CreateAdminAccount(tx *sql.Tx) (err error) {
	hash, err := auth.BcryptHash("password", 10)
	if err != nil {
		return err
	}
	return RegisterAccount(tx, "admin", hash)
}

// CreateSystemAccount create an inaccessible account used for automatic internal purposes
func CreateSystemAccount(tx *sql.Tx) (err error) {
	password, err := auth.RandomID(32)
	if err != nil {
		return
	}
	hash, err := auth.BcryptHash(password, 10)
	if err != nil {
		return
	}
	return RegisterAccount(tx, "system", hash)
}

// ClearTables deletes the contents of specified DB tables. Only used for tests.
func ClearTables(tables ...string) error {
	for _, t := range tables {
		// Clear open post body bucket
		if rocksDBisOpen() {
			switch t {
			case "boards", "threads", "posts":
				readOpts := grocksdb.NewDefaultReadOptions()
				defer readOpts.Destroy()

				writeOpts := grocksdb.NewDefaultWriteOptions()
				defer writeOpts.Destroy()
				iter := rocksDB.NewIterator(readOptions)
				iter.SeekToFirst()
				firstSlice := iter.Key()
				first := firstSlice.Data()
				firstSlice.Free()
				iter.SeekToLast()
				lastSlice := iter.Key()
				last := lastSlice.Data()
				lastSlice.Free()
				iter.Close()
				wb := grocksdb.NewWriteBatch()
				wb.DeleteRange(first, last)
				err := rocksDB.Write(writeOpts, wb)
				if err != nil {
					return err
				}
			}
		}

		if _, err := sqlDB.Exec(`DELETE FROM ` + t); err != nil {
			return err
		}
	}
	return nil
}
