package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv" // Import the godotenv package
	"github.com/wexlerdev/randomwords/internal/database"
	_ "github.com/lib/pq"     // PostgreSQL driver
)

type Config struct {
	dbQueries *database.Queries
}

func main() {
	err := godotenv.Load()
	if err != nil {
		// Log an error if the .env file isn't found, but don't fatal
		// as variables might be set directly in the environment for production.
		log.Println("No .env file found or error loading .env:", err)
		log.Println("Attempting to read environment variables directly.")
	}

	connStr := os.Getenv("DB_URL")
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("Error opening database connection: %v", err)
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		log.Fatalf("Error connecting to the database: %v", err)
	}
	fmt.Println("Successfully connected to PostgreSQL database")

	queriesObj := database.New(db)
	cfg := Config{
		dbQueries: queriesObj,
	}

	cfg.populateDbFromFile("funWords.txt")
}
