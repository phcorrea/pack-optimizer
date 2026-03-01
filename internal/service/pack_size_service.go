package service

import (
	"fmt"
	"sort"
	"sync"
)

var defaultPackSizes = []int{250, 500, 1000, 2000, 5000}

// NormalizePackSizes validates pack sizes, removes duplicates, and returns
// a descending-sorted slice so larger packs are evaluated first.
func NormalizePackSizes(packSizes []int) ([]int, error) {
	if len(packSizes) == 0 {
		return nil, ErrInvalidPackSizes
	}

	// seen removes duplicates to improve optimization performance.
	seen := make(map[int]struct{}, len(packSizes))
	normalized := make([]int, 0, len(packSizes))
	for _, size := range packSizes {
		if size <= 0 {
			return nil, fmt.Errorf("%w: %d", ErrInvalidPackSizes, size)
		}
		if size > maxInt32Value {
			return nil, fmt.Errorf("%w: %d exceeds int32 max value %d", ErrInvalidPackSizes, size, maxInt32Value)
		}
		if _, duplicate := seen[size]; duplicate {
			continue
		}

		seen[size] = struct{}{}
		normalized = append(normalized, size)
	}

	if len(normalized) == 0 {
		return nil, ErrInvalidPackSizes
	}

	sort.Sort(sort.Reverse(sort.IntSlice(normalized)))
	return normalized, nil
}

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
