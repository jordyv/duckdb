package duckdb_test

import (
	"testing"
	"time"

	"github.com/alifiroozi80/duckdb"
	_ "github.com/marcboeker/go-duckdb/v2"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

// Define test structs
type User struct {
	ID    uint `gorm:"primaryKey"`
	Name  string
	Email string `gorm:"unique"`
}

type Product struct {
	ID    uint `gorm:"primaryKey"`
	Name  string
	Price float64
}

type Post struct {
	ID        uint `gorm:"primaryKey"`
	Content   string
	CreatedAt time.Time
}

// TestMigratorBasicSchema verifies basic schema creation.
func TestMigratorBasicSchema(t *testing.T) {
	db, err := gorm.Open(duckdb.Open("test.db"), &gorm.Config{})
	assert.NoError(t, err)

	// Migrate User table
	err = db.AutoMigrate(&Product{})
	assert.NoError(t, err)

	// Check if table exists
	assert.True(t, db.Migrator().HasTable(&Product{}))
	assert.True(t, db.Migrator().HasColumn(&Product{}, "Price"))
}

// TestMigratorDropTable verifies dropping a table.
func TestMigratorDropTable(t *testing.T) {
	db, err := gorm.Open(duckdb.Open("test.db"), &gorm.Config{})
	assert.NoError(t, err)

	db.AutoMigrate(&User{})
	assert.True(t, db.Migrator().HasTable(&User{}))

	// Drop table and verify
	db.Migrator().DropTable(&User{})
	assert.False(t, db.Migrator().HasTable(&User{}))
}

// TestUniqueConstraint tests that unique constraints are enforced.
func TestUniqueConstraint(t *testing.T) {
	db, err := gorm.Open(duckdb.Open("test.db"), &gorm.Config{})
	assert.NoError(t, err)

	db.AutoMigrate(&User{})
	assert.True(t, db.Migrator().HasColumn(&User{}, "Email"))

	// Create first user
	user1 := User{Name: "User1", Email: "user@example.com"}
	db.Create(&user1)

	// Attempt to create a second user with the same email
	user2 := User{Name: "User2", Email: "user@example.com"}
	result := db.Create(&user2)
	assert.Error(t, result.Error, "Expected unique constraint violation")
}

// TestDefaultValues verifies that default values are set correctly.
func TestDefaultValues(t *testing.T) {
	db, err := gorm.Open(duckdb.Open("test.db"), &gorm.Config{})
	assert.NoError(t, err)

	db.AutoMigrate(&Post{})

	// Insert a new post without specifying CreatedAt
	post := Post{Content: "Hello, World!"}
	db.Create(&post)

	// Verify CreatedAt has a value (defaulted to the current timestamp)
	assert.NotZero(t, post.CreatedAt)
}
