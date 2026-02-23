package service

import (
	"errors"
	"testing"
)

func TestOptimize_ChallengeExamples(t *testing.T) {
	tests := []struct {
		name      string
		packSizes []int
		ordered   int
		total     int
		totalPack int
		packs     []PackBreakdown
	}{
		{
			name:      "order 1",
			packSizes: []int{250, 500, 1000, 2000, 5000},
			ordered:   1,
			total:     250,
			totalPack: 1,
			packs:     []PackBreakdown{{Size: 250, Count: 1}},
		},
		{
			name:      "order 251",
			packSizes: []int{250, 500, 1000, 2000, 5000},
			ordered:   251,
			total:     500,
			totalPack: 1,
			packs:     []PackBreakdown{{Size: 500, Count: 1}},
		},
		{
			name:      "order 501",
			packSizes: []int{250, 500, 1000, 2000, 5000},
			ordered:   501,
			total:     750,
			totalPack: 2,
			packs:     []PackBreakdown{{Size: 500, Count: 1}, {Size: 250, Count: 1}},
		},
		{
			name:      "order 12001",
			packSizes: []int{250, 500, 1000, 2000, 5000},
			ordered:   12001,
			total:     12250,
			totalPack: 4,
			packs:     []PackBreakdown{{Size: 5000, Count: 2}, {Size: 2000, Count: 1}, {Size: 250, Count: 1}},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			plan, err := Optimize(tc.ordered, tc.packSizes)
			if err != nil {
				t.Fatalf("Optimize returned error: %v", err)
			}

			if plan.TotalItems != tc.total {
				t.Fatalf("TotalItems = %d, want %d", plan.TotalItems, tc.total)
			}

			if plan.TotalPacks != tc.totalPack {
				t.Fatalf("TotalPacks = %d, want %d", plan.TotalPacks, tc.totalPack)
			}

			if len(plan.Packs) != len(tc.packs) {
				t.Fatalf("len(Packs) = %d, want %d", len(plan.Packs), len(tc.packs))
			}

			for i := range tc.packs {
				if plan.Packs[i] != tc.packs[i] {
					t.Fatalf("Packs[%d] = %+v, want %+v", i, plan.Packs[i], tc.packs[i])
				}
			}
		})
	}
}

func TestOptimize_EdgeCases(t *testing.T) {
	tests := []struct {
		name      string
		packSizes []int
		ordered   int
		total     int
		totalPack int
		packs     []PackBreakdown
	}{
		{
			name:      "unsorted and duplicate pack sizes",
			packSizes: []int{500, 250, 250, 1000},
			ordered:   501,
			total:     750,
			totalPack: 2,
			packs:     []PackBreakdown{{Size: 500, Count: 1}, {Size: 250, Count: 1}},
		},
		{
			name:      "single pack size requires over shipment",
			packSizes: []int{53},
			ordered:   500000,
			total:     500002,
			totalPack: 9434,
			packs:     []PackBreakdown{{Size: 53, Count: 9434}},
		},
		{
			name:      "large exact order with non-standard packs",
			packSizes: []int{23, 31, 53},
			ordered:   500000,
			total:     500000,
			totalPack: 9438,
			packs:     []PackBreakdown{{Size: 53, Count: 9429}, {Size: 31, Count: 7}, {Size: 23, Count: 2}},
		},
		{
			name:      "no exact match chooses least over shipment",
			packSizes: []int{6, 10},
			ordered:   13,
			total:     16,
			totalPack: 2,
			packs:     []PackBreakdown{{Size: 10, Count: 1}, {Size: 6, Count: 1}},
		},
		{
			name:      "tie on shipped items chooses fewer packs",
			packSizes: []int{4, 6, 9},
			ordered:   11,
			total:     12,
			totalPack: 2,
			packs:     []PackBreakdown{{Size: 6, Count: 2}},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			plan, err := Optimize(tc.ordered, tc.packSizes)
			if err != nil {
				t.Fatalf("Optimize returned error: %v", err)
			}

			if plan.TotalItems != tc.total {
				t.Fatalf("TotalItems = %d, want %d", plan.TotalItems, tc.total)
			}
			if plan.TotalPacks != tc.totalPack {
				t.Fatalf("TotalPacks = %d, want %d", plan.TotalPacks, tc.totalPack)
			}
			if len(plan.Packs) != len(tc.packs) {
				t.Fatalf("len(Packs) = %d, want %d", len(plan.Packs), len(tc.packs))
			}
			for i := range tc.packs {
				if plan.Packs[i] != tc.packs[i] {
					t.Fatalf("Packs[%d] = %+v, want %+v", i, plan.Packs[i], tc.packs[i])
				}
			}
		})
	}
}

func TestOptimize_InvalidInput(t *testing.T) {
	maxInt32 := int(^uint32(0) >> 1)

	_, err := Optimize(0, []int{250, 500})
	if !errors.Is(err, ErrInvalidItemsOrdered) {
		t.Fatalf("expected ErrInvalidItemsOrdered, got %v", err)
	}

	_, err = Optimize(100, []int{})
	if !errors.Is(err, ErrInvalidPackSizes) {
		t.Fatalf("expected ErrInvalidPackSizes, got %v", err)
	}

	_, err = Optimize(100, []int{0, 250})
	if !errors.Is(err, ErrInvalidPackSizes) {
		t.Fatalf("expected ErrInvalidPackSizes for zero pack size, got %v", err)
	}

	_, err = Optimize(100, []int{-5, 250})
	if !errors.Is(err, ErrInvalidPackSizes) {
		t.Fatalf("expected ErrInvalidPackSizes for negative pack size, got %v", err)
	}

	_, err = Optimize(maxInt32+1, []int{250, 500})
	if !errors.Is(err, ErrInvalidItemsOrdered) {
		t.Fatalf("expected ErrInvalidItemsOrdered for items_ordered above int32 max, got %v", err)
	}

	_, err = Optimize(100, []int{maxInt32 + 1, 250})
	if !errors.Is(err, ErrInvalidPackSizes) {
		t.Fatalf("expected ErrInvalidPackSizes for pack size above int32 max, got %v", err)
	}
}

func TestOptimize_BigNumbersReturnsRangeError(t *testing.T) {
	maxInt32 := int(^uint32(0) >> 1)

	_, err := Optimize(maxInt32-100, []int{maxInt32 - 1000, maxInt32 - 999})
	if !errors.Is(err, ErrOptimizationTooLarge) {
		t.Fatalf("expected ErrOptimizationTooLarge, got %v", err)
	}
}
