// filepath: /Users/bhupesh/Documents/work/experiments/cache/cache/memory_cache_test.go
package cache

import (
	"context"
	"strconv"
	"testing"
	"time"
)

func TestSetAndGet(t *testing.T) {
	ctx := context.Background()
	cache := NewMemoryCache()

	// Set a key
	err := cache.Set(ctx, "key1", []byte("value1"), 5*time.Minute)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Get the key
	value, err := cache.Get(ctx, "key1")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if string(value.([]byte)) != "value1" {
		t.Errorf("expected value1, got %s", value)
	}
}

func TestDelete(t *testing.T) {
	ctx := context.Background()
	cache := NewMemoryCache()

	// Set a key
	cache.Set(ctx, "key1", []byte("value1"), 5*time.Minute)

	// Delete the key
	err := cache.Delete(ctx, "key1")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Try to get the deleted key
	_, err = cache.Get(ctx, "key1")
	if err == nil {
		t.Errorf("expected an error, got nil")
	}
}

func TestClear(t *testing.T) {
	ctx := context.Background()
	cache := NewMemoryCache()

	// Set some keys
	cache.Set(ctx, "key1", []byte("value1"), 5*time.Minute)
	cache.Set(ctx, "key2", []byte("value2"), 5*time.Minute)

	// Clear the cache
	err := cache.Clear()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Try to get the cleared keys
	_, err = cache.Get(ctx, "key1")
	if err == nil {
		t.Errorf("expected an error, got nil")
	}
	_, err = cache.Get(ctx, "key2")
	if err == nil {
		t.Errorf("expected an error, got nil")
	}
}

func TestDeleteByPattern(t *testing.T) {
	ctx := context.Background()
	cache := NewMemoryCache()

	// Set some keys
	cache.Set(ctx, "key1", []byte("value1"), 5*time.Minute)
	cache.Set(ctx, "key2", []byte("value2"), 5*time.Minute)
	cache.Set(ctx, "key3", []byte("value3"), 5*time.Minute)
	cache.Set(ctx, "key-pattern-1", []byte("value4"), 5*time.Minute)
	cache.Set(ctx, "key-pattern-2", []byte("value5"), 5*time.Minute)
	cache.Set(ctx, "shipping_rate_958e389f-5bfa-480f-a94e-90db68b58605_42ae3e65-a813-4b06-8d7d-0e2290735fcc_2e36eab9-412a-4f7c-9f97-56b387eb71b2", []byte("value6"), 5*time.Minute)
	cache.Set(ctx, "shipping_rate_c48dff77-1647-4891-adc7-e038b7c3651a_42ae3e65-a813-4b06-8d7d-0e2290735fcc_2e36eab9-412a-4f7c-9f97-56b387eb71b2", []byte("value6"), 5*time.Minute)

	// Delete keys by pattern
	err := cache.DeleteByPattern(ctx, "key-pattern-\\d+")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Check if keys matching the pattern are deleted
	if _, err := cache.Get(ctx, "key-pattern-1"); err == nil {
		t.Errorf("expected key-pattern-1 to be deleted")
	}
	if _, err := cache.Get(ctx, "key-pattern-2"); err == nil {
		t.Errorf("expected key-pattern-2 to be deleted")
	}

	// Check if other keys are not deleted
	if _, err := cache.Get(ctx, "key1"); err != nil {
		t.Errorf("expected key1 to exist, got error %v", err)
	}
	if _, err := cache.Get(ctx, "key2"); err != nil {
		t.Errorf("expected key2 to exist, got error %v", err)
	}
	if _, err := cache.Get(ctx, "key3"); err != nil {
		t.Errorf("expected key3 to exist, got error %v", err)
	}

	err = cache.DeleteByPattern(ctx, "^shipping_rate_[^_]+_42ae3e65-a813-4b06-8d7d-0e2290735fcc_2e36eab9-412a-4f7c-9f97-56b387eb71b2$")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Check if keys matching the pattern are deleted
	if _, err := cache.Get(ctx, "shipping_rate_958e389f-5bfa-480f-a94e-90db68b58605_42ae3e65-a813-4b06-8d7d-0e2290735fcc_2e36eab9-412a-4f7c-9f97-56b387eb71b2"); err == nil {
		t.Errorf("expected shipping_rate_958e389f-5bfa-480f-a94e-90db68b58605_42ae3e65-a813-4b06-8d7d-0e2290735fcc_2e36eab9-412a-4f7c-9f97-56b387eb71b2 to be deleted")
	}

	if _, err := cache.Get(ctx, "shipping_rate_c48dff77-1647-4891-adc7-e038b7c3651a_42ae3e65-a813-4b06-8d7d-0e2290735fcc_2e36eab9-412a-4f7c-9f97-56b387eb71b2"); err == nil {
		t.Errorf("expected shipping_rate_c48dff77-1647-4891-adc7-e038b7c3651a_42ae3e65-a813-4b06-8d7d-0e2290735fcc_2e36eab9-412a-4f7c-9f97-56b387eb71b2 to be deleted")
	}
}

func BenchmarkDeleteByPattern(b *testing.B) {
	ctx := context.Background()
	cache := NewMemoryCache()

	// Set some keys
	for i := 0; i < 1000; i++ {
		cache.Set(ctx, "key-pattern-"+strconv.Itoa(i), []byte("value"), 5*time.Minute)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cache.DeleteByPattern(ctx, "key-pattern-\\d+")
	}
}
