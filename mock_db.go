package cache_proxy

import (
	"errors"
	"fmt"
	"github.com/brianvoe/gofakeit/v6"
	"sync"
	"time"
)

var ErrNotFound = errors.New("not found")

type UserDatabase struct {
	total    int
	qryCount int
	mu       sync.Mutex
	users    map[string]*gofakeit.PersonInfo
}

func NewUserDatabase(size int) *UserDatabase {
	newUsers := func(size int) map[string]*gofakeit.PersonInfo {
		faker := gofakeit.NewCrypto()
		gofakeit.SetGlobalFaker(faker)
		users := make(map[string]*gofakeit.PersonInfo, size)
		for i := 0; i < size; i++ {
			id := fmt.Sprintf("id[%v]", i)
			users[id] = gofakeit.Person()
		}
		return users
	}

	return &UserDatabase{
		total: size,
		users: newUsers(size),
	}
}

func (db *UserDatabase) QueryUserById(id string) (user *gofakeit.PersonInfo, err error) {
	db.mu.Lock()
	defer db.mu.Unlock()

	db.qryCount++
	time.Sleep(10 * time.Microsecond)
	v, ok := db.users[id]
	if !ok {
		return nil, ErrNotFound
	}
	return v, nil
}

func (db *UserDatabase) GetUserIds() []string {
	size := len(db.users)
	ids := make([]string, 0, size)
	for i := 0; i < size; i++ {
		ids = append(ids, fmt.Sprintf("id[%v]", i))
	}
	return ids
}
