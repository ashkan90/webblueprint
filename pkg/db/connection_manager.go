package db

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"sync"
	"time"

	_ "github.com/lib/pq" // PostgreSQL driver
)

// Config holds database connection configuration
type Config struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
	SSLMode  string
}

// ConnectionManager manages database connections
type ConnectionManager struct {
	db     *sql.DB
	config Config
	mu     sync.RWMutex
}

var (
	instance *ConnectionManager
	once     sync.Once
)

// GetConnectionManager returns the singleton instance of ConnectionManager
func GetConnectionManager() *ConnectionManager {
	once.Do(func() {
		instance = &ConnectionManager{}
	})
	return instance
}

// Initialize initializes the database connection
func (cm *ConnectionManager) Initialize(config Config) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if cm.db != nil {
		return nil // Already initialized
	}

	connStr := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		config.Host, config.Port, config.User, config.Password, config.DBName, config.SSLMode,
	)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return fmt.Errorf("failed to open database connection: %w", err)
	}

	// Set connection pool parameters
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(10)
	db.SetConnMaxLifetime(5 * time.Minute)

	// Verify connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	cm.db = db
	cm.config = config
	log.Println("Database connection initialized successfully")
	return nil
}

// GetDB returns the database connection
func (cm *ConnectionManager) GetDB() *sql.DB {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	return cm.db
}

// Close closes the database connection
func (cm *ConnectionManager) Close() error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if cm.db == nil {
		return nil
	}

	err := cm.db.Close()
	if err != nil {
		return fmt.Errorf("error closing database connection: %w", err)
	}

	cm.db = nil
	log.Println("Database connection closed")
	return nil
}

// WithTransaction executes a function within a transaction
func (cm *ConnectionManager) WithTransaction(ctx context.Context, fn func(*sql.Tx) error) error {
	db := cm.GetDB()
	if db == nil {
		return fmt.Errorf("database connection not initialized")
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	// Execute the function
	err = fn(tx)
	if err != nil {
		// If there's an error, attempt to rollback
		if rbErr := tx.Rollback(); rbErr != nil {
			// Log rollback error, but return the original error
			log.Printf("Error rolling back transaction: %v", rbErr)
		}
		return err
	}

	// If everything was successful, commit the transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
