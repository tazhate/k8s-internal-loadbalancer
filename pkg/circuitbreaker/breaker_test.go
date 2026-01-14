package circuitbreaker

import (
	"errors"
	"testing"
	"time"
)

func TestCircuitBreakerStates(t *testing.T) {
	cb := New(3, time.Second, time.Second, 2)

	// Initial state should be closed
	if cb.State() != "closed" {
		t.Errorf("Initial state should be closed, got %s", cb.State())
	}

	// Simulate failures to open the circuit
	failFunc := func() error {
		return errors.New("test error")
	}

	// First failure
	cb.Execute(failFunc)
	if cb.State() != "closed" {
		t.Errorf("State should still be closed after 1 failure, got %s", cb.State())
	}

	// Second failure - should open
	cb.Execute(failFunc)
	if cb.State() != "open" {
		t.Errorf("State should be open after 2 consecutive failures, got %s", cb.State())
	}

	// Try to execute when open - should fail immediately
	err := cb.Execute(func() error {
		return nil
	})
	if err == nil {
		t.Error("Execute should fail when circuit is open")
	}
}

func TestCircuitBreakerRecovery(t *testing.T) {
	cb := New(2, 100*time.Millisecond, 100*time.Millisecond, 2)

	// Open the circuit
	failFunc := func() error {
		return errors.New("test error")
	}
	cb.Execute(failFunc)
	cb.Execute(failFunc)

	if cb.State() != "open" {
		t.Fatalf("Circuit should be open, got %s", cb.State())
	}

	// Wait for timeout
	time.Sleep(150 * time.Millisecond)

	// Should be half-open now
	successFunc := func() error {
		return nil
	}

	// First success in half-open
	err := cb.Execute(successFunc)
	if err != nil {
		t.Errorf("Execute should succeed in half-open state: %v", err)
	}

	// Second success should close the circuit
	err = cb.Execute(successFunc)
	if err != nil {
		t.Errorf("Execute should succeed: %v", err)
	}

	// Give it a moment to transition
	time.Sleep(10 * time.Millisecond)

	if cb.State() != "closed" {
		t.Errorf("Circuit should be closed after successful requests, got %s", cb.State())
	}
}

func TestCircuitBreakerStats(t *testing.T) {
	cb := New(5, time.Second, time.Second, 3)

	successFunc := func() error {
		return nil
	}
	failFunc := func() error {
		return errors.New("test error")
	}

	// Execute some successful requests
	for i := 0; i < 3; i++ {
		cb.Execute(successFunc)
	}

	stats := cb.Stats()
	if stats["total_successes"].(uint32) != 3 {
		t.Errorf("Expected 3 successes, got %v", stats["total_successes"])
	}

	// Execute some failures
	for i := 0; i < 2; i++ {
		cb.Execute(failFunc)
	}

	stats = cb.Stats()
	if stats["total_failures"].(uint32) != 2 {
		t.Errorf("Expected 2 failures, got %v", stats["total_failures"])
	}
}
