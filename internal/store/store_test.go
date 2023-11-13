package store

import (
	"testing"
	"time"
)

func TestStore_Set_And_Get(t *testing.T) {
	store := NewStore()
	key := "key"
	value := "value"

	_, exist := store.Get(key)
	if exist {
		t.Errorf("Expected  key %s to not exist, but it existed", key)
	}

	store.Set(key, value)

	got, ok := store.Get(key)
	if !ok {
		t.Errorf("Expected key %s to exist, but it did not existed", key)
	}
	if got != value {
		t.Errorf("Expected value '%s', but got '%s'", value, got)
	}
}

func TestStore_SetPx_And_Get(t *testing.T) {
	store := NewStore()
	key := "key"
	value := "value"
	expiration := int64(1000)

	store.SetPx(key, value, expiration)

	got, ok := store.Get(key)
	if !ok {
		t.Errorf("Expected key %s to exist, but it did not existed", key)
	}
	if got != value {
		t.Errorf("Expected value '%s', but got '%s'", value, got)
	}

	time.Sleep(time.Duration(expiration) * time.Millisecond)
	_, exist := store.Get(key)
	if exist {
		t.Errorf("Expected key '%s' to have expired, but it existed", key)
	}
}
