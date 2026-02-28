package service

import (
	"errors"
	"reflect"
	"sync"
	"testing"
)

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
