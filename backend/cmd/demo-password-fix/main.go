package main

import (
	"fmt"
	"log"

	"github.com/Timmy00125/VolunteerSync-Project/backend/internal/modules/auth/models"
)

func main() {
	fmt.Println("=== Password Verification Fix Demonstration ===")

	// Create a test user
	user := &models.User{
		Email: "test@example.com",
	}

	// Set password
	password := "SecurePassword123!"
	fmt.Printf("Setting password: %s\n", password)

	if err := user.SetPassword(password); err != nil {
		log.Fatalf("Failed to set password: %v", err)
	}

	fmt.Printf("Password hash generated: %s\n\n", user.PasswordHash)

	// Test correct password
	fmt.Println("Testing correct password...")
	if err := user.VerifyPassword(password); err != nil {
		fmt.Printf("❌ FAILED: %v\n", err)
	} else {
		fmt.Println("✓ SUCCESS: Password verified correctly!")
	}

	// Test incorrect password
	fmt.Println("\nTesting incorrect password...")
	wrongPassword := "WrongPassword123!"
	if err := user.VerifyPassword(wrongPassword); err != nil {
		fmt.Printf("✓ SUCCESS: Incorrect password correctly rejected: %v\n", err)
	} else {
		fmt.Println("❌ FAILED: Incorrect password was accepted!")
	}

	fmt.Println("\n=== Fix Summary ===")
	fmt.Println("The bug was in the verifyArgon2Hash function.")
	fmt.Println("fmt.Sscanf with '%s$%s' doesn't work because %s stops at whitespace,")
	fmt.Println("not at the $ delimiter. This caused all password verifications to fail.")
	fmt.Println("\nThe fix uses strings.Split() to properly parse the 'salt$hash' format.")
}
