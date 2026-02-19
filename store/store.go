package store

import (
	"fmt"
	"sync"

	"github.com/jamea/models"
)

// UserRecord represents a user in the in-memory store
type UserRecord struct {
	models.User
	Password string
}

var (
	users   map[string]*UserRecord // keyed by email
	masools []models.Masool
	mu      sync.RWMutex
	nextId  int64 = 1
)

// Init seeds the in-memory store with hardcoded admin users
func Init() {
	mu.Lock()
	defer mu.Unlock()

	users = map[string]*UserRecord{
		"ams_admin@gmail.com": {
			User: models.User{
				Id:     1,
				Name:   "AMS Admin",
				Email:  "ams_admin@gmail.com",
				Module: models.ModuleAms,
			},
			Password: "admin123",
		},
		"tbum_admin@gmail.com": {
			User: models.User{
				Id:     2,
				Name:   "TBUM Admin",
				Email:  "tbum_admin@gmail.com",
				Module: models.ModuleTbum,
			},
			Password: "admin123",
		},
		"ptbum_admin@gmail.com": {
			User: models.User{
				Id:     3,
				Name:   "PTBUM Admin",
				Email:  "ptbum_admin@gmail.com",
				Module: models.ModulePtbum,
			},
			Password: "admin123",
		},
	}
	masools = []models.Masool{}
}

// FindUserByEmail looks up a user by email (thread-safe)
func FindUserByEmail(email string) (*UserRecord, error) {
	mu.RLock()
	defer mu.RUnlock()

	user, exists := users[email]
	if !exists {
		return nil, fmt.Errorf("user not found")
	}
	return user, nil
}

// AddMasool adds a new Masool record to the store
func AddMasool(m models.Masool) {
	mu.Lock()
	defer mu.Unlock()
	m.Id = nextId
	nextId++
	masools = append(masools, m)
}

// GetMasools returns all Masool records
func GetMasools() []models.Masool {
	mu.RLock()
	defer mu.RUnlock()
	// Return a copy to avoid race conditions if the caller modifies it (shallow copy of slice is fine for now as we append)
	result := make([]models.Masool, len(masools))
	copy(result, masools)
	return result
}
