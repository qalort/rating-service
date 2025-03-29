package db

import (
        "database/sql"
        "fmt"
        "os"
        "strings"

        _ "github.com/lib/pq"
)

// NewPostgresConnection creates a new PostgreSQL connection
func NewPostgresConnection() (*sql.DB, error) {
        host := os.Getenv("PGHOST")
        if host == "" {
                host = "localhost"
        }

        port := os.Getenv("PGPORT")
        if port == "" {
                port = "5432"
        }

        user := os.Getenv("PGUSER")
        if user == "" {
                user = "postgres"
        }

        password := os.Getenv("PGPASSWORD")
        if password == "" {
                password = ""
        }

        dbname := os.Getenv("PGDATABASE")
        if dbname == "" {
                dbname = "ratings"
        }

        // Check if we have a DATABASE_URL (common in some hosting environments)
        dbURL := os.Getenv("DATABASE_URL")
        if dbURL != "" {
                // Make sure we have SSL mode required
                if !strings.Contains(dbURL, "sslmode=") {
                        dbURL += "?sslmode=require"
                }
                return sql.Open("postgres", dbURL)
        }

        // Otherwise construct the connection string from individual parameters
        psqlInfo := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=require",
                host, port, user, password, dbname)

        db, err := sql.Open("postgres", psqlInfo)
        if err != nil {
                return nil, err
        }

        // Test the connection
        if err = db.Ping(); err != nil {
                return nil, err
        }

        // Set connection pool settings
        db.SetMaxOpenConns(25)
        db.SetMaxIdleConns(10)

        return db, nil
}
