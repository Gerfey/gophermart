package benchmark

import (
	"testing"

	"golang.org/x/crypto/bcrypt"
)

func BenchmarkPasswordHashing(b *testing.B) {
	password := "password123"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	}
}

func BenchmarkPasswordVerification(b *testing.B) {
	password := "password123"
	hash, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = bcrypt.CompareHashAndPassword(hash, []byte(password))
	}
}

func BenchmarkPasswordHashingWithDifferentCosts(b *testing.B) {
	password := "password123"

	b.Run("Cost=4", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = bcrypt.GenerateFromPassword([]byte(password), 4)
		}
	})

	b.Run("Cost=10", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = bcrypt.GenerateFromPassword([]byte(password), 10)
		}
	})

	b.Run("Cost=14", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = bcrypt.GenerateFromPassword([]byte(password), 14)
		}
	})
}

func BenchmarkPasswordVerificationWithDifferentPasswords(b *testing.B) {
	passwords := []string{
		"short",
		"password123",
		"verylongpasswordthatismorethan30characters",
		"P@ssw0rd!WithSpecialChars",
	}

	for _, pw := range passwords {
		b.Run(pw, func(b *testing.B) {
			hash, _ := bcrypt.GenerateFromPassword([]byte(pw), bcrypt.DefaultCost)

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = bcrypt.CompareHashAndPassword(hash, []byte(pw))
			}
		})
	}
}
