package seeder

import (
	"fmt"
	"log"
	"math/rand"
	"time"

	"gorm.io/gorm"
)

type Seeder struct {
	db         *gorm.DB
	userSeeder *UserSeeder
}

func NewSeeder(db *gorm.DB) *Seeder {
	// Initialize random seed
	rand.Seed(time.Now().UnixNano())

	return &Seeder{
		db:         db,
		userSeeder: NewUserSeeder(db),
	}
}

// SeedAll runs all seeders
func (s *Seeder) SeedAll() error {
	log.Println("Starting database seeding...")

	start := time.Now()

	// Check if we need to seed
	needsSeeding, err := s.userSeeder.NeedsSeeding()
	if err != nil {
		return fmt.Errorf("failed to check if seeding is needed: %w", err)
	}

	if !needsSeeding {
		log.Println("Database already has data. Skipping seeding.")
		return nil
	}

	// Seed users with default profiles
	log.Println("Seeding users...")
	if err := s.userSeeder.Seed(); err != nil {
		return fmt.Errorf("failed to seed users: %w", err)
	}

	// Future seeders can be added here
	// Example:
	// log.Println("Seeding posts...")
	// if err := s.postSeeder.Seed(); err != nil {
	//     return fmt.Errorf("failed to seed posts: %w", err)
	// }

	elapsed := time.Since(start)
	log.Printf("Database seeding completed successfully in %v", elapsed)

	return nil
}

// SeedUsers seeds only users
func (s *Seeder) SeedUsers(count int) error {
	log.Printf("Seeding %d users...", count)
	return s.userSeeder.SeedUsers(count)
}

// CleanupAll removes all seeded data
func (s *Seeder) CleanupAll() error {
	log.Println("Starting database cleanup...")

	start := time.Now()

	// Cleanup in reverse order of dependencies
	if err := s.userSeeder.CleanupUsers(); err != nil {
		return fmt.Errorf("failed to cleanup users: %w", err)
	}

	elapsed := time.Since(start)
	log.Printf("Database cleanup completed successfully in %v", elapsed)

	return nil
}

// GetStats returns seeding statistics
func (s *Seeder) GetStats() (map[string]interface{}, error) {
	userCount, err := s.userSeeder.GetUserCount()
	if err != nil {
		return nil, fmt.Errorf("failed to get user count: %w", err)
	}

	stats := map[string]interface{}{
		"users":     userCount,
		"timestamp": time.Now(),
	}

	return stats, nil
}

// VerifySeeding checks if seeding was successful
func (s *Seeder) VerifySeeding() error {
	log.Println("Verifying seeding results...")

	userCount, err := s.userSeeder.GetUserCount()
	if err != nil {
		return fmt.Errorf("failed to verify user count: %w", err)
	}

	if userCount == 0 {
		return fmt.Errorf("no users found after seeding")
	}

	// Verify that users have default profiles
	var profileCount int64
	if err := s.db.Model(&struct {
		ID        uint `gorm:"primaryKey"`
		UserID    uint
		IsDefault bool
	}{}).Where("is_default = ?", true).Count(&profileCount).Error; err != nil {
		return fmt.Errorf("failed to count default profiles: %w", err)
	}

	if profileCount != userCount {
		return fmt.Errorf("mismatch between users (%d) and default profiles (%d)", userCount, profileCount)
	}

	log.Printf("Seeding verification successful: %d users with %d default profiles", userCount, profileCount)
	return nil
}

// SeedDevelopmentData seeds a smaller dataset for development
func (s *Seeder) SeedDevelopmentData() error {
	log.Println("Seeding development data...")
	return s.userSeeder.SeedUsers(20) // Only 20 users for development
}

// SeedProductionData seeds a larger dataset for production-like testing
func (s *Seeder) SeedProductionData() error {
	log.Println("Seeding production-like data...")
	return s.userSeeder.SeedUsers(1000) // 1000 users for production testing
}

// ResetAndSeed completely resets the database and seeds fresh data
func (s *Seeder) ResetAndSeed() error {
	log.Println("Resetting database and seeding fresh data...")

	// Cleanup existing data
	if err := s.CleanupAll(); err != nil {
		log.Printf("Warning: cleanup failed, continuing anyway: %v", err)
	}

	// Seed fresh data
	if err := s.SeedAll(); err != nil {
		return fmt.Errorf("failed to seed after reset: %w", err)
	}

	// Verify seeding
	if err := s.VerifySeeding(); err != nil {
		return fmt.Errorf("seeding verification failed: %w", err)
	}

	return nil
}