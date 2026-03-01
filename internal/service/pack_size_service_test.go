package service

import (
	"errors"
	"reflect"
	"sync"
	"testing"
)

func TestNormalizePackSizes(t *testing.T) {
	tests := []struct {
		name      string
		input     []int
		want      []int
		expectErr bool
	}{
		{
			name:  "sorts descending and removes duplicates",
			input: []int{500, 250, 250, 1000},
			want:  []int{1000, 500, 250},
		},
		{
			name:  "single unique value from duplicates",
			input: []int{10, 10, 10},
			want:  []int{10},
		},
		{
			name:      "empty pack sizes",
			input:     []int{},
			expectErr: true,
		},
		{
			name:      "zero pack size",
			input:     []int{0, 250},
			expectErr: true,
		},
		{
			name:      "negative pack size",
			input:     []int{-5, 250},
			expectErr: true,
		},
		{
			name:      "pack size above int32 max",
			input:     []int{maxInt32Value + 1, 250},
			expectErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := NormalizePackSizes(tc.input)
			if tc.expectErr {
				if !errors.Is(err, ErrInvalidPackSizes) {
					t.Fatalf("expected ErrInvalidPackSizes, got %v", err)
				}
				return
			}

			if err != nil {
				t.Fatalf("NormalizePackSizes returned error: %v", err)
			}
			if !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("NormalizePackSizes() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestNewInMemoryPackSizeService(t *testing.T) {
	service, err := NewInMemoryPackSizeService([]int{250, 500, 1000})
	if err != nil {
		t.Fatalf("NewInMemoryPackSizeService returned error: %v", err)
	}

	if !reflect.DeepEqual(service.GetPackSizes(), []int{1000, 500, 250}) {
		t.Fatalf("unexpected initial pack sizes: %v", service.GetPackSizes())
	}
}

func TestNewInMemoryPackSizeService_InvalidInitialPackSizes(t *testing.T) {
	_, err := NewInMemoryPackSizeService([]int{})
	if !errors.Is(err, ErrInvalidPackSizes) {
		t.Fatalf("expected ErrInvalidPackSizes, got %v", err)
	}
}

func TestInMemoryPackSizeService_GetPackSizesReturnsCopy(t *testing.T) {
	service, err := NewInMemoryPackSizeService([]int{250, 500, 1000})
	if err != nil {
		t.Fatalf("NewInMemoryPackSizeService returned error: %v", err)
	}

	got := service.GetPackSizes()
	got[0] = 999999

	if !reflect.DeepEqual(service.GetPackSizes(), []int{1000, 500, 250}) {
		t.Fatalf("GetPackSizes should return isolated copy, got %v", service.GetPackSizes())
	}
}

func TestInMemoryPackSizeService_SetPackSizes(t *testing.T) {
	service, err := NewInMemoryPackSizeService([]int{250, 500, 1000})
	if err != nil {
		t.Fatalf("NewInMemoryPackSizeService returned error: %v", err)
	}

	if err := service.SetPackSizes([]int{10, 40, 20, 10}); err != nil {
		t.Fatalf("SetPackSizes returned error: %v", err)
	}

	if !reflect.DeepEqual(service.GetPackSizes(), []int{40, 20, 10}) {
		t.Fatalf("unexpected pack sizes after set: %v", service.GetPackSizes())
	}
}

func TestInMemoryPackSizeService_SetPackSizesInvalid(t *testing.T) {
	service, err := NewInMemoryPackSizeService([]int{250, 500, 1000})
	if err != nil {
		t.Fatalf("NewInMemoryPackSizeService returned error: %v", err)
	}

	if err := service.SetPackSizes([]int{0, 250}); !errors.Is(err, ErrInvalidPackSizes) {
		t.Fatalf("expected ErrInvalidPackSizes, got %v", err)
	}
}

func TestGetPackSizeService_Singleton(t *testing.T) {
	first, err := GetPackSizeService()
	if err != nil {
		t.Fatalf("GetPackSizeService returned error: %v", err)
	}

	second, err := GetPackSizeService()
	if err != nil {
		t.Fatalf("GetPackSizeService returned error: %v", err)
	}

	if first != second {
		t.Fatal("expected GetPackSizeService to return the same instance")
	}
}

func TestGetPackSizeService_SingletonConcurrent(t *testing.T) {
	const goroutines = 64

	var wg sync.WaitGroup
	instances := make(chan PackSizeService, goroutines)

	for range goroutines {
		wg.Add(1)
		go func() {
			defer wg.Done()

			service, err := GetPackSizeService()
			if err != nil {
				t.Errorf("GetPackSizeService returned error: %v", err)
				return
			}
			instances <- service
		}()
	}

	wg.Wait()
	close(instances)

	var singleton PackSizeService
	for instance := range instances {
		if singleton == nil {
			singleton = instance
			continue
		}
		if instance != singleton {
			t.Fatal("expected all goroutines to receive the same singleton instance")
		}
	}
}
