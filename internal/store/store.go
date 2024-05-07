package store

import (
	"fmt"
	"sync"
	"time"
)

type Store struct {
	db  sync.Map
	exp sync.Map
}

func NewStore() *Store {
	fmt.Println("Initialize key value store...")
	store := &Store{db: sync.Map{}, exp: sync.Map{}}
	return store
}

func (k *Store) Get(key string) (string, bool) {
	value, ok := k.db.Load(key)
	if !ok {
		return "", false
	}
	expiration, expOk := k.exp.Load(key)
	if expOk {
		expTime := expiration.(int64)
		now := time.Now().UnixNano() / int64(time.Millisecond)
		if expTime < now {
			k.db.Delete(key)
			k.exp.Delete(key)
			return "", false
		}
	}
	return value.(string), ok
}

func (k *Store) Set(key string, value string) string {
	k.db.Store(key, value)
	return "OK"
}

func (k *Store) SetPx(key string, value string, exp int64) string {
	now := time.Now().UnixNano() / int64(time.Millisecond)
	k.db.Store(key, value)
	k.exp.Store(key, now+exp)
	return "OK"
}
