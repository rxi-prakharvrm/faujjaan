package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"clothes-shop/api/internal/db"
)

func main() {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Fatal("DATABASE_URL is required")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	pool, err := db.Connect(ctx, databaseURL)
	if err != nil {
		log.Fatalf("db connect error: %v", err)
	}
	defer pool.Close()

	fmt.Println("✅ Successfully connected to Supabase database!")

	// Read and execute migration files
	upSQL, err := os.ReadFile("migrations/000001_init.up.sql")
	if err != nil {
		log.Fatalf("failed to read migration file: %v", err)
	}

	fmt.Println("Running migration 000001_init.up.sql...")
	_, err = pool.Exec(ctx, string(upSQL))
	if err != nil {
		log.Fatalf("migration failed: %v", err)
	}
	fmt.Println("✅ Migration 000001_init.up.sql applied successfully!")

	// Run seed data
	seedSQL, err := os.ReadFile("migrations/000002_seed_dev.up.sql")
	if err != nil {
		fmt.Println("No seed file found, skipping...")
	} else {
		fmt.Println("Running seed 000002_seed_dev.up.sql...")
		_, err = pool.Exec(ctx, string(seedSQL))
		if err != nil {
			log.Fatalf("seed failed: %v", err)
		}
		fmt.Println("✅ Seed data applied successfully!")
	}

	fmt.Println("\n🎉 All migrations completed! Your Supabase database is ready.")
}
