package db

import (
        "context"
        "database/sql"
        "fmt"
        "net/url"
        "os"
        "strings"
        "time"

        _ "github.com/go-sql-driver/mysql"
)

// PostgresURLToMySQLDSN converts a PostgreSQL database URL to a MySQL DSN format
func PostgresURLToMySQLDSN(postgresURL string) (string, error) {
        // Parse the PostgreSQL URL
        parsedURL, err := url.Parse(postgresURL)
        if err != nil {
                return "", fmt.Errorf("failed to parse database URL: %w", err)
        }

        // Extract user and password
        var user, password string
        if parsedURL.User != nil {
                user = parsedURL.User.Username()
                password, _ = parsedURL.User.Password()
        }

        // Extract host and port
        host := parsedURL.Hostname()
        port := parsedURL.Port()
        if port == "" {
                port = "5432" // Default PostgreSQL port
        }

        // Extract database name (remove leading slash)
        dbname := strings.TrimPrefix(parsedURL.Path, "/")

        // Construct the MySQL connection string
        // Format: [username[:password]@][protocol[(address)]]/dbname[?param1=value1&...&paramN=valueN]
        mysqlDSN := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true&timeout=10s",
                user, password, host, port, dbname)

        return mysqlDSN, nil
}

// NewMySQLConnection creates a new MySQL connection
func NewMySQLConnection() (*sql.DB, error) {
        host := os.Getenv("MYSQL_HOST")
        if host == "" {
                host = "localhost"
        }

        port := os.Getenv("MYSQL_PORT")
        if port == "" {
                port = "3306"
        }

        user := os.Getenv("MYSQL_USER")
        if user == "" {
                user = "user"
        }

        password := os.Getenv("MYSQL_PASSWORD")
        if password == "" {
                password = "password"
        }

        dbname := os.Getenv("MYSQL_DATABASE")
        if dbname == "" {
                dbname = "rating_system"
        }

        var db *sql.DB
        var err error

        // Check if we have a DATABASE_URL (common in some hosting environments)
        dbURL := os.Getenv("DATABASE_URL")
        if dbURL != "" {
                // Check if this is a PostgreSQL URL (starts with postgres:// or postgresql://)
                if strings.HasPrefix(dbURL, "postgres://") || strings.HasPrefix(dbURL, "postgresql://") {
                        // Convert the PostgreSQL URL to MySQL DSN format
                        mysqlDSN, err := PostgresURLToMySQLDSN(dbURL)
                        if err != nil {
                                return nil, err
                        }
                        db, err = sql.Open("mysql", mysqlDSN)
                } else {
                        // Assume it's already in MySQL format
                        db, err = sql.Open("mysql", dbURL)
                }
        } else {
                // Construct the connection string from individual parameters
                mysqlInfo := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true&timeout=5s",
                        user, password, host, port, dbname)
                db, err = sql.Open("mysql", mysqlInfo)
        }

        if err != nil {
                return nil, err
        }

        // Test the connection with a timeout
        ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
        defer cancel()
        
        if err = db.PingContext(ctx); err != nil {
                return nil, err
        }

        // Set connection pool settings
        db.SetMaxOpenConns(25)
        db.SetMaxIdleConns(10)
        db.SetConnMaxLifetime(time.Hour)

        return db, nil
}