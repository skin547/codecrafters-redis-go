package store

import "time"

type Store struct {
	db  map[string]string
	exp map[string]int64
}

func NewStore() *Store {
	return &Store{db: map[string]string{}, exp: map[string]int64{}}
}

func (k *Store) Get(key string) (string, bool) {
	if exp, exist := k.exp[key]; exist {
		now := time.Now().UnixNano() / int64(time.Millisecond)
		if exp < now {
			delete(k.exp, key)
			delete(k.db, key)
			return "", false
		}
	}
	val, ok := k.db[key]
	return val, ok
}

func (k *Store) Set(key string, value string) string {
	k.db[key] = value
	return "OK"
}

func (k *Store) SetPx(key string, value string, exp int64) string {
	now := time.Now().UnixNano() / int64(time.Millisecond)
	k.db[key] = value
	k.exp[key] = now + exp
	return "OK"
}
