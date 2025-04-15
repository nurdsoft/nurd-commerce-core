package cache

import (
	"context"
	"testing"
	"time"
)

func TestCacheSetAndGet(t *testing.T) {
	ctx := context.Background()
	cache := New()

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

func TestCacheDelete(t *testing.T) {
	ctx := context.Background()
	cache := New()

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

func TestCacheClear(t *testing.T) {
	ctx := context.Background()
	cache := New()

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

func TestCacheDeleteByPattern(t *testing.T) {
	ctx := context.Background()
	cache := New()

	// Set some keys
	cache.Set(ctx, "key1", []byte("value1"), 5*time.Minute)
	cache.Set(ctx, "key2", []byte("value2"), 5*time.Minute)
	cache.Set(ctx, "key-pattern-1", []byte("value3"), 5*time.Minute)
	cache.Set(ctx, "key-pattern-2", []byte("value4"), 5*time.Minute)

	// Delete keys by pattern
	err := cache.DeleteByPattern(ctx, "key-pattern-.*")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Check if the keys matching the pattern are deleted
	_, err = cache.Get(ctx, "key-pattern-1")
	if err == nil {
		t.Errorf("expected an error, got nil")
	}
	_, err = cache.Get(ctx, "key-pattern-2")
	if err == nil {
		t.Errorf("expected an error, got nil")
	}

	// Check if the other keys are still present
	_, err = cache.Get(ctx, "key1")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	_, err = cache.Get(ctx, "key2")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}
