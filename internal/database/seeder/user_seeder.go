package seeder

import (
	"fmt"
	"log"
	"math/rand"
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"social_server/internal/models/postgres"
)

type UserSeeder struct {
	db *gorm.DB
}

func NewUserSeeder(db *gorm.DB) *UserSeeder {
	return &UserSeeder{db: db}
}

var (
	firstNames = []string{
		"Nguyễn", "Trần", "Lê", "Phạm", "Hoàng", "Huỳnh", "Phan", "Vũ", "Võ", "Đặng",
		"Bùi", "Đỗ", "Hồ", "Ngô", "Dương", "Lý", "John", "Michael", "Sarah", "Emily",
		"David", "Jessica", "Robert", "Ashley", "James", "Amanda", "William", "Melissa",
		"Christopher", "Michelle", "Daniel", "Kimberly", "Matthew", "Amy", "Anthony", "Angela",
		"Mark", "Brenda", "Donald", "Emma", "Steven", "Olivia", "Paul", "Cynthia", "Andrew", "Marie",
		"Joshua", "Janet", "Kenneth", "Catherine", "Kevin", "Frances", "Brian", "Christine",
	}

	lastNames = []string{
		"Văn Minh", "Thị Lan", "Văn Nam", "Thị Hoa", "Văn Đức", "Thị Mai", "Văn Hùng", "Thị Linh",
		"Văn Tuấn", "Thị Nga", "Văn Khang", "Thị Trang", "Smith", "Johnson", "Williams", "Brown",
		"Jones", "Garcia", "Miller", "Davis", "Rodriguez", "Martinez", "Hernandez", "Lopez",
		"Gonzalez", "Wilson", "Anderson", "Thomas", "Taylor", "Moore", "Jackson", "Martin",
		"Lee", "Perez", "Thompson", "White", "Harris", "Sanchez", "Clark", "Ramirez",
		"Lewis", "Robinson", "Walker", "Young", "Allen", "King", "Wright", "Scott",
	}

	companies = []string{
		"viettel", "vnpt", "fpt", "techcombank", "vietcombank", "bidv", "mb", "acb",
		"google", "microsoft", "apple", "amazon", "meta", "netflix", "spotify", "uber",
		"shopee", "lazada", "grab", "gojek", "zalo", "tiki", "sendo", "vingroup",
	}

	bioTemplates = []string{
		"Passionate developer who loves coding and learning new technologies.",
		"Coffee enthusiast and tech lover. Building amazing apps every day.",
		"Software engineer with a passion for clean code and great user experiences.",
		"Full-stack developer who enjoys solving complex problems.",
		"Tech enthusiast, gamer, and lifelong learner.",
		"Building the future one line of code at a time.",
		"Love traveling, photography, and creating awesome software.",
		"Backend developer specializing in scalable systems.",
		"Frontend wizard who creates beautiful user interfaces.",
		"DevOps engineer passionate about automation and efficiency.",
		"Mobile app developer creating apps that matter.",
		"Data scientist turning data into insights.",
		"Product manager who loves bringing ideas to life.",
		"UI/UX designer focused on user-centered design.",
		"Entrepreneur building the next big thing.",
		"Student passionate about technology and innovation.",
		"Freelancer helping businesses grow through technology.",
		"Open source contributor and community builder.",
		"Tech blogger sharing knowledge and experiences.",
		"Startup founder disrupting traditional industries.",
	}
)

func (s *UserSeeder) SeedUsers(count int) error {
	log.Printf("Starting to seed %d users with default profiles...", count)

	return s.db.Transaction(func(tx *gorm.DB) error {
		for i := 1; i <= count; i++ {
			// Generate user data
			firstName := firstNames[rand.Intn(len(firstNames))]
			lastName := lastNames[rand.Intn(len(lastNames))]
			company := companies[rand.Intn(len(companies))]

			email := fmt.Sprintf("%s.%s%d@%s.com",
				generateSlug(firstName),
				generateSlug(lastName),
				rand.Intn(999)+1,
				company)

			// Hash password (all users will have password "password123")
			hashedPassword, err := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
			if err != nil {
				return fmt.Errorf("failed to hash password for user %d: %w", i, err)
			}

			// Create user
			user := &postgres.User{
				PasswordHash: string(hashedPassword),
				Email:        email,
				Role:         postgres.RoleUser,
				IsOnline:     rand.Intn(10) < 3, // 30% chance of being online
				IsVerified:   rand.Intn(10) < 7, // 70% chance of being verified
				IsActive:     true,
				IsBanned:     false,
				Settings: postgres.UserSettings{
					PrivacyProfileVisibility:    getRandomVisibility(),
					PrivacyShowOnlineStatus:     rand.Intn(10) < 8, // 80% show online status
					PrivacyAllowFriendRequests:  rand.Intn(10) < 9, // 90% allow friend requests
					NotificationsEmail:          rand.Intn(10) < 7, // 70% email notifications
					NotificationsPush:           rand.Intn(10) < 8, // 80% push notifications
					NotificationsFriendRequests: rand.Intn(10) < 9, // 90% friend request notifications
					NotificationsMessages:       rand.Intn(10) < 9, // 90% message notifications
					NotificationsPosts:          rand.Intn(10) < 8, // 80% post notifications
				},
				CreatedAt: getRandomCreatedAt(),
				UpdatedAt: time.Now(),
			}

			// Set last seen for offline users
			if !user.IsOnline {
				lastSeen := time.Now().Add(-time.Duration(rand.Intn(24*7)) * time.Hour) // Within last week
				user.LastSeen = &lastSeen
			}

			if err := tx.Create(user).Error; err != nil {
				return fmt.Errorf("failed to create user %d: %w", i, err)
			}

			// Create default profile
			displayName := fmt.Sprintf("%s %s", firstName, lastName)
			bio := bioTemplates[rand.Intn(len(bioTemplates))]

			profile := &postgres.Profile{
				UserID:      user.ID,
				FirstName:   firstName,
				LastName:    lastName,
				DisplayName: displayName,
				Bio:         bio,
				Phone:       generatePhoneNumber(),
				DateOfBirth: generateDateOfBirth(),
				CreatedAt:   user.CreatedAt,
				UpdatedAt:   time.Now(),
			}

			if err := tx.Create(profile).Error; err != nil {
				return fmt.Errorf("failed to create profile for user %d: %w", i, err)
			}

			if i%10 == 0 {
				log.Printf("Created %d users so far...", i)
			}
		}

		log.Printf("Successfully seeded %d users with default profiles", count)
		return nil
	})
}

func (s *UserSeeder) CleanupUsers() error {
	log.Println("Cleaning up seeded users...")

	return s.db.Transaction(func(tx *gorm.DB) error {
		// Delete all profiles first (foreign key constraint)
		if err := tx.Unscoped().Where("1 = 1").Delete(&postgres.Profile{}).Error; err != nil {
			return fmt.Errorf("failed to delete profiles: %w", err)
		}

		// Delete all users
		if err := tx.Unscoped().Where("1 = 1").Delete(&postgres.User{}).Error; err != nil {
			return fmt.Errorf("failed to delete users: %w", err)
		}

		// Reset auto-increment sequences
		if err := tx.Exec("ALTER SEQUENCE users_id_seq RESTART WITH 1").Error; err != nil {
			log.Printf("Warning: failed to reset users sequence: %v", err)
		}

		if err := tx.Exec("ALTER SEQUENCE profiles_id_seq RESTART WITH 1").Error; err != nil {
			log.Printf("Warning: failed to reset profiles sequence: %v", err)
		}

		log.Println("Successfully cleaned up all seeded data")
		return nil
	})
}

// Helper functions
func generateSlug(name string) string {
	// Simple slug generation - convert to lowercase and remove special characters
	slug := ""
	for _, char := range name {
		if (char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') || (char >= '0' && char <= '9') {
			if char >= 'A' && char <= 'Z' {
				slug += string(char + 32) // Convert to lowercase
			} else {
				slug += string(char)
			}
		}
	}
	return slug
}

func generatePhoneNumber() string {
	// Generate Vietnamese phone numbers
	prefixes := []string{"084", "085", "088", "091", "094", "096", "097", "098", "032", "033", "034", "035", "036", "037", "038", "039"}
	prefix := prefixes[rand.Intn(len(prefixes))]
	number := fmt.Sprintf("%s%07d", prefix, rand.Intn(9999999))
	return number
}

func generateDateOfBirth() *time.Time {
	// Generate dates between 18-65 years ago
	minAge := 18
	maxAge := 65

	now := time.Now()
	minDate := now.AddDate(-maxAge, 0, 0)
	maxDate := now.AddDate(-minAge, 0, 0)

	delta := maxDate.Unix() - minDate.Unix()
	randomSeconds := rand.Int63n(delta)

	dob := time.Unix(minDate.Unix()+randomSeconds, 0)
	return &dob
}

func getRandomVisibility() string {
	visibilities := []string{"public", "friends", "private"}
	weights := []int{6, 3, 1} // 60% public, 30% friends, 10% private

	total := 0
	for _, weight := range weights {
		total += weight
	}

	random := rand.Intn(total)
	cumulative := 0

	for i, weight := range weights {
		cumulative += weight
		if random < cumulative {
			return visibilities[i]
		}
	}

	return "public" // fallback
}

func getRandomCreatedAt() time.Time {
	// Generate created dates within the last 2 years
	now := time.Now()
	twoYearsAgo := now.AddDate(-2, 0, 0)

	delta := now.Unix() - twoYearsAgo.Unix()
	randomSeconds := rand.Int63n(delta)

	return time.Unix(twoYearsAgo.Unix()+randomSeconds, 0)
}

// Seed 100 users by default
func (s *UserSeeder) Seed() error {
	return s.SeedUsers(100)
}

// Get count of users in database
func (s *UserSeeder) GetUserCount() (int64, error) {
	var count int64
	err := s.db.Model(&postgres.User{}).Count(&count).Error
	return count, err
}

// Check if seeding is needed
func (s *UserSeeder) NeedsSeeding() (bool, error) {
	count, err := s.GetUserCount()
	if err != nil {
		return false, err
	}
	return count == 0, nil
}
