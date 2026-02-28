package service

import "sync"

var defaultPackSizes = []int{250, 500, 1000, 2000, 5000}

// PackSizeService manages configured pack sizes.
type PackSizeService interface {
	GetPackSizes() []int
	SetPackSizes(packSizes []int) error
}

// InMemoryPackSizeService stores pack sizes in memory and is safe for concurrent use.
type InMemoryPackSizeService struct {
	mu        sync.RWMutex
	packSizes []int
}

var (
	packSizeServiceOnce     sync.Once
	packSizeServiceInstance PackSizeService
	packSizeServiceInitErr  error
)

// GetPackSizeService returns the singleton pack size service.
func GetPackSizeService() (PackSizeService, error) {
	packSizeServiceOnce.Do(func() {
		packSizeServiceInstance, packSizeServiceInitErr = NewInMemoryPackSizeService(defaultPackSizes)
	})

	if packSizeServiceInitErr != nil {
		return nil, packSizeServiceInitErr
	}

	return packSizeServiceInstance, nil
}

// NewInMemoryPackSizeService creates a pack size service with an initial set of sizes.
func NewInMemoryPackSizeService(initialPackSizes []int) (*InMemoryPackSizeService, error) {
	normalized, err := NormalizePackSizes(initialPackSizes)
	if err != nil {
		return nil, err
	}

	return &InMemoryPackSizeService{
		packSizes: normalized,
	}, nil
}

// GetPackSizes returns a copy of currently configured pack sizes.
func (s *InMemoryPackSizeService) GetPackSizes() []int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]int, len(s.packSizes))
	copy(result, s.packSizes)
	return result
}

// SetPackSizes validates and replaces currently configured pack sizes.
func (s *InMemoryPackSizeService) SetPackSizes(packSizes []int) error {
	normalized, err := NormalizePackSizes(packSizes)
	if err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.packSizes = normalized
	return nil
}
