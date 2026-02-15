package tools

import (
	"context"
	"database/sql"
	"dsoob/backend/include"
	"path"
	"sync"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

var Database *sql.DB

func DatabaseSetup(stop context.Context, await *sync.WaitGroup) {
	p := path.Join(DATA_DIRECTORY, "database", "main.db")
	t := time.Now()

	db, err := sql.Open("sqlite3", p)
	if err != nil {
		LoggerDatabase.Log(FATAL, "Cannot open database: %s", err.Error())
		return
	}
	if _, err := db.Exec(include.DatabaseSchema); err != nil {
		LoggerDatabase.Log(FATAL, "Cannot update database: %s", err.Error())
		return
	}
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	Database = db

	// Shutdown Logic
	await.Add(1)
	go func() {
		defer await.Done()
		<-stop.Done()
		Database.Close()
		LoggerDatabase.Log(INFO, "Closed")
	}()
	LoggerDatabase.Log(INFO, "Ready in %s", time.Since(t))
}
