package main

import (
	"database/sql"
	"embed"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/lib/pq"

	"github.com/reveliant/mabinouze/api"
)

// Flags
var httpHost, httpSocket string
var httpPort uint
var debug bool

//go:embed sql/*
var sqlLib embed.FS

// Database connector
var db *sql.DB

func execSQL(filename string) {
	sqlCmd, err := sqlLib.ReadFile(filename)
	if err != nil {
		log.Fatal(err.Error())
	}

	_, err = db.Query(string(sqlCmd))
	if err != nil {
		log.Fatal(err.Error())
	}
}

func serve() {
	// Create server with timeout
	srv := &http.Server{
		Addr:    fmt.Sprintf("%s:%d", httpHost, httpPort),
		Handler: api.NewRouter(db, debug),
		// set timeout due CWE-400 - Potential Slowloris Attack
		ReadHeaderTimeout: 5 * time.Second,
	}

	// Serve from IP socket or UNIX socket
	if httpSocket == "" {
		log.Println("Listening on", srv.Addr)
		if err := srv.ListenAndServe(); err != nil {
			log.Printf("Failed to start server: %v", err)
		}
	} else {
		log.Println("Listening on", httpSocket)
		unixListener, err := net.Listen("unix", httpSocket)
		if err != nil {
			log.Printf("Failed to start server: %v", err)
			return
		}
		srv.Serve(unixListener)
	}
}

func main() {
	log.SetFlags(0)

	// Application whole flags
	flag.BoolVar(&debug, "debug", false, "Enable debug mode")
	// Database connector flags
	var dbHost, dbName, dbUser string
	var dbPort uint
	flag.StringVar(&dbHost, "dbhost", "localhost", "Database host")
	flag.UintVar(&dbPort, "dbport", 5432, "Database port")
	flag.StringVar(&dbName, "dbname", "mabinouze", "Database name")
	flag.StringVar(&dbUser, "dbuser", "mabinouze", "Database user")
	// HTTP Server flags
	flag.StringVar(&httpHost, "host", "", "Listening host or IP address")
	flag.UintVar(&httpPort, "port", 8080, "Listening port")
	flag.StringVar(&httpSocket, "socket", "", "Listening socket")
	// Commands flags
	var cmdInit, cmdDemo, cmdClean bool
	flag.BoolVar(&cmdInit, "init-db", false, "Initialize database with schema")
	flag.BoolVar(&cmdDemo, "demo-db", false, "Fill database with demo content")
	flag.BoolVar(&cmdClean, "clean-db", false, "Clean expired rounds from database")
	// Parse flags
	flag.Parse()

	// Open DB
	dbConfig := pq.Config{
		Host:     dbHost,
		Port:     uint16(dbPort),
		User:     dbUser,
		Database: dbName,
		//Passfile: dbPassfile
	}
	conn, err := pq.NewConnectorConfig(dbConfig)
	if err != nil {
		log.Fatal(err.Error())
	}
	db = sql.OpenDB(conn)
	defer db.Close()

	if cmdInit {
		log.Println("Initialize database with schema")
		execSQL("sql/schema.sql")
	}
	if cmdDemo {
		log.Println("Fill with demo content")
		execSQL("sql/demo.sql")
	}
	if cmdClean {
		log.Println("Clean expired rounds")
		execSQL("sql/clean.sql")
	}

	if cmdInit || cmdDemo || cmdClean {
		return
	}

	log.Println("Start server")
	serve()
}
