package benchmark

import (
	"strconv"
	"testing"

	"github.com/Gerfey/gophermart/internal/service"
)

func BenchmarkIsValidLuhnNumber(b *testing.B) {
	validNumbers := []string{
		"79927398713",
		"4561261212345467",
		"4561261212345475",
		"371449635398431",
		"378282246310005",
		"5555555555554444",
		"4111111111111111",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, number := range validNumbers {
			service.IsValidLuhnNumber(number)
		}
	}
}

func BenchmarkIsValidLuhnNumberAlternative(b *testing.B) {
	validNumbers := []string{
		"79927398713",
		"4561261212345467",
		"4561261212345475",
		"371449635398431",
		"378282246310005",
		"5555555555554444",
		"4111111111111111",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, number := range validNumbers {
			service.IsValidLuhnNumber(number)
		}
	}
}

func BenchmarkLuhnValidationLargeDataset(b *testing.B) {
	testNumbers := generateTestNumbers(1000)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for j := 0; j < 100; j++ {
			service.IsValidLuhnNumber(testNumbers[j])
		}
	}
}

func BenchmarkLuhnValidationOptimizedLargeDataset(b *testing.B) {
	testNumbers := generateTestNumbers(1000)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for j := 0; j < 100; j++ {
			service.IsValidLuhnNumber(testNumbers[j])
		}
	}
}

func generateTestNumbers(count int) []string {
	result := make([]string, count)
	for i := 0; i < count; i++ {
		baseNum := 1000000000 + i
		result[i] = strconv.Itoa(baseNum)
	}
	return result
}
