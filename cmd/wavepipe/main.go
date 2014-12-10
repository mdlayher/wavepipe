// Command wavepipe serves a HTTP API for the wavepipe media server.
// Please see the project on GitHub for usage: https://github.com/mdlayher/wavepipe.
package main

import (
	"flag"
	"log"
	"net"
	"net/http"
	"os"
	"path"
	"time"

	"github.com/mdlayher/wavepipe/bindata"
	"github.com/mdlayher/wavepipe/data"
	"github.com/mdlayher/wavepipe/data/models"
	"github.com/mdlayher/wavepipe/wptest"

	"github.com/stretchr/graceful"
)

const (
	// sqlite3 is the name of the sqlite3 driver for the database
	sqlite3 = "sqlite3"

	// sqlite3DBAsset is the name of the bindata asset which stores the sqlite schema
	sqlite3SchemaAsset = "res/sqlite3/wavepipe.sql"

	// driver is the database/sql driver used for the database instance
	driver = sqlite3
)

// version is the current git hash, injected by the Go linker
var version string

var (
	// db is the DSN used for the database instance
	db string

	// host is the address to which the HTTP server is bound
	host string

	// noRoot disables creation of a root account on database creation
	noRoot bool

	// timeout is the duration the server will wait before forcibly closing
	// ongoing HTTP connections
	timeout time.Duration
)

func init() {
	// Set up flags
	flag.StringVar(&db, "db", "wavepipe.db", "DSN for database instance")
	flag.StringVar(&host, "host", ":8080", "HTTP server host")
	flag.BoolVar(&noRoot, "no-root", false, "disable creation of root account for new database")
	flag.DurationVar(&timeout, "timeout", 5*time.Second, "HTTP graceful timeout duration")
}

func main() {
	flag.Parse()

	// Report information on startup
	log.Printf("wavepipe: starting [pid: %d] [version: %s]", os.Getpid(), version)

	// Determine if database newly created
	var created bool
	var err error

	// If database is sqlite3, perform initial setup
	if driver == sqlite3 {
		// Attempt setup, check if already created
		created, err = sqlite3Setup(db)
		if err != nil {
			log.Fatal(err)
		}

		if created {
			log.Println("wavepipe: created sqlite3 database:", db)
		} else {
			log.Println("wavepipe: using sqlite3 database:", db)
		}
	}

	// Open database connection
	wpdb := &data.DB{}
	if err := wpdb.Open(driver, db); err != nil {
		log.Fatal(err)
	}

	// Unless skipped, perform initial root user setup for sqlite3
	if driver == sqlite3 && created && !noRoot {
		// Generate root user
		root := &models.User{
			Username: "root",
			RoleID:   models.RoleAdmin,
		}

		// Generate a random password
		password := wptest.RandomString(12)
		log.Println("wavepipe: creating root user: root /", password)
		if err := root.SetPassword(password); err != nil {
			log.Fatal(err)
		}

		// Save root user
		if err := wpdb.InsertUser(root); err != nil {
			log.Fatal(err)
		}
	} else if noRoot {
		log.Println("wavepipe: skipping creation of root user")
	}

	// Start HTTP server using wavepipe handler on specified host
	log.Println("wavepipe: listening:", host)
	if err := graceful.ListenAndServe(&http.Server{
		Addr: host,
		// TODO(mdlayher): replace with customizable http.Handler
		Handler: http.NewServeMux(),
	}, timeout); err != nil {
		// Ignore error on failed "accept" when closing
		if nErr, ok := err.(*net.OpError); !ok || nErr.Op != "accept" {
			log.Fatal(err)
		}
	}

	log.Println("wavepipe: shutting down")

	// Close database connection
	if err := wpdb.Close(); err != nil {
		log.Fatal(err)
	}

	log.Println("wavepipe: graceful shutdown complete")
}

// sqlite3Setup performs setup routines specific to a sqlite3 database.
// On success, it returns a boolean indicating if the database was created.
// On failure, it returns an error.
func sqlite3Setup(dsn string) (bool, error) {
	// Check if database already exists at specified location
	dbPath := path.Clean(dsn)
	_, err := os.Stat(dbPath)
	if err == nil {
		// Database exists, skip setup
		return false, nil
	}

	// Any other errors, return now
	if !os.IsNotExist(err) {
		return false, err
	}

	// Retrieve sqlite3 database schema asset
	asset, err := bindata.Asset(sqlite3SchemaAsset)
	if err != nil {
		return false, err
	}

	// Open empty database file at target path
	wpdb := &data.DB{}
	if err := wpdb.Open(sqlite3, dbPath); err != nil {
		return false, err
	}

	// Execute schema to build database
	if _, err := wpdb.Exec(string(asset)); err != nil {
		return false, err
	}

	// Close database
	if err := wpdb.Close(); err != nil {
		return false, err
	}

	return true, nil
}
