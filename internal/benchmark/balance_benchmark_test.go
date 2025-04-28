package benchmark

import (
	"sync"
	"testing"
)

type Balance struct {
	mu        sync.RWMutex
	current   float64
	withdrawn float64
}

func NewBalance(initial float64) *Balance {
	return &Balance{
		current:   initial,
		withdrawn: 0,
	}
}

func (b *Balance) GetBalance() (float64, float64) {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.current, b.withdrawn
}

func (b *Balance) AddAccrual(amount float64) error {
	if amount <= 0 {
		return nil
	}

	b.mu.Lock()
	defer b.mu.Unlock()

	b.current += amount
	return nil
}

func (b *Balance) Withdraw(amount float64) error {
	if amount <= 0 {
		return nil
	}

	b.mu.Lock()
	defer b.mu.Unlock()

	if b.current < amount {
		return nil
	}

	b.current -= amount
	b.withdrawn += amount
	return nil
}

func BenchmarkBalanceGetBalance(b *testing.B) {
	balance := NewBalance(1000.0)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		balance.GetBalance()
	}
}

func BenchmarkBalanceAddAccrual(b *testing.B) {
	balance := NewBalance(1000.0)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		balance.AddAccrual(10.0)
	}
}

func BenchmarkBalanceWithdraw(b *testing.B) {
	balance := NewBalance(1000.0)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		balance.Withdraw(1.0)
	}
}

func BenchmarkBalanceConcurrentOperations(b *testing.B) {
	balance := NewBalance(1000.0)

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			balance.GetBalance()
			balance.AddAccrual(1.0)
			balance.GetBalance()
			balance.Withdraw(0.5)
		}
	})
}
