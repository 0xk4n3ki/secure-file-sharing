package database

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

var PG_Client *sql.DB = DBinstance()

func DBinstance() *sql.DB {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	username := os.Getenv("PGUSER")
	if username == "" {
		log.Fatal("Username not found")
	}

	password := os.Getenv("PGPASSWORD")
	if password == "" {
		log.Fatal("Password not found")
	}

	dbName := "go-oauth"

	adminConnStr := os.Getenv("ADMIN_DB")
	if adminConnStr == "" {
		adminConnStr = fmt.Sprintf("postgres://%s:%s@localhost:5432/postgres?sslmode=disable", username, password)
	}
	adminDB, err := sql.Open("postgres", adminConnStr)
	if err != nil {
		log.Fatal(err)
	}
	defer adminDB.Close()

	var exists bool
	err = adminDB.QueryRow(`SELECT EXISTS(SELECT 1 FROM pg_database WHERE datname=$1)`, dbName).Scan(&exists)
	if err != nil {
		log.Fatal(err)
	}

	if !exists {
		_, err = adminDB.Exec(fmt.Sprintf(`CREATE DATABASE "%s"`, dbName))
		if err != nil {
			log.Fatalf("Error creating database %s: %v", dbName, err)
		}
		fmt.Printf("Database %s created\n", dbName)
	}

	connStr := os.Getenv("PG_URL")
	if connStr == "" {
		connStr = fmt.Sprintf("postgres://%s:%s@localhost:5432/%s?sslmode=disable", username, password, dbName)
	}

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}

	if err := db.Ping(); err != nil {
		log.Fatal(err)
	}

	fmt.Println("Connected to Postgresql")

	return db
}

func CreateUserTable(db *sql.DB) {
	query := `
	CREATE TABLE IF NOT EXISTS users (
		id SERIAL PRIMARY KEY,
		first_name TEXT NOT NULL,
		last_name TEXT NOT NULL,
		email TEXT UNIQUE NOT NULL,
		token TEXT,
		refresh_token TEXT,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		user_id UUID DEFAULT gen_random_uuid()
	);`

	_, err := db.Exec(query)
	if err != nil {
		log.Fatalf("Error creating users table: %v", err)
	}
}

func EnablePgCrypto(db *sql.DB) {
	_, err := db.Exec(`CREATE EXTENSION IF NOT EXISTS pgcrypto;`)
	if err != nil {
		log.Fatalf("Error enabling pgcrypto extension: %v", err)
	}
}
