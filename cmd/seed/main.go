package main

import (
	"flag"
	"fmt"
	"log"

	"social_server/internal/config"
	"social_server/internal/database"
	"social_server/internal/database/seeder"
)

func main() {
	// Command line flags
	var (
		count   = flag.Int("count", 100, "Number of users to seed")
		cleanup = flag.Bool("cleanup", false, "Cleanup all seeded data")
		verify  = flag.Bool("verify", false, "Verify seeding results")
		reset   = flag.Bool("reset", false, "Reset database and seed fresh data")
		dev     = flag.Bool("dev", false, "Seed development data (20 users)")
		prod    = flag.Bool("prod", false, "Seed production data (1000 users)")
		stats   = flag.Bool("stats", false, "Show seeding statistics")
	)
	flag.Parse()

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Connect to database
	db, err := database.NewPostgresConnection(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Create seeder
	s := seeder.NewSeeder(db.DB)

	// Execute based on flags
	switch {
	case *cleanup:
		if err := s.CleanupAll(); err != nil {
			log.Fatalf("Failed to cleanup data: %v", err)
		}
		fmt.Println("✅ Database cleanup completed successfully")

	case *verify:
		if err := s.VerifySeeding(); err != nil {
			log.Fatalf("Seeding verification failed: %v", err)
		}
		fmt.Println("✅ Seeding verification passed")

	case *reset:
		if err := s.ResetAndSeed(); err != nil {
			log.Fatalf("Failed to reset and seed: %v", err)
		}
		fmt.Println("✅ Database reset and seeding completed successfully")

	case *dev:
		if err := s.SeedDevelopmentData(); err != nil {
			log.Fatalf("Failed to seed development data: %v", err)
		}
		fmt.Println("✅ Development data seeding completed successfully")

	case *prod:
		if err := s.SeedProductionData(); err != nil {
			log.Fatalf("Failed to seed production data: %v", err)
		}
		fmt.Println("✅ Production data seeding completed successfully")

	case *stats:
		statistics, err := s.GetStats()
		if err != nil {
			log.Fatalf("Failed to get stats: %v", err)
		}
		fmt.Println("📊 Database Statistics:")
		for key, value := range statistics {
			fmt.Printf("  %s: %v\n", key, value)
		}

	default:
		if *count > 0 {
			if err := s.SeedUsers(*count); err != nil {
				log.Fatalf("Failed to seed %d users: %v", *count, err)
			}
			fmt.Printf("✅ Successfully seeded %d users with default profiles\n", *count)
		} else {
			if err := s.SeedAll(); err != nil {
				log.Fatalf("Failed to seed all data: %v", err)
			}
			fmt.Println("✅ All seeding completed successfully")
		}

		// Automatically verify after seeding
		if err := s.VerifySeeding(); err != nil {
			log.Printf("⚠️  Warning: Seeding verification failed: %v", err)
		} else {
			fmt.Println("✅ Seeding verification passed")
		}
	}

	// Show final stats
	if !*stats {
		statistics, err := s.GetStats()
		if err == nil {
			fmt.Printf("\n📊 Current database contains %v users\n", statistics["users"])
		}
	}
}
