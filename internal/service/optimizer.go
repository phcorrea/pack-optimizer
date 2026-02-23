package service

import (
	"errors"
	"fmt"
	"sort"
)

const maxInt32Value = int(^uint32(0) >> 1)

var (
	ErrInvalidItemsOrdered  = errors.New("items_ordered must be greater than zero")
	ErrInvalidPackSizes     = errors.New("pack_sizes must contain at least one positive integer")
	ErrOptimizationTooLarge = errors.New("optimization range is too large")
	errNoPackingPlan        = errors.New("no valid packing combination found")
	errReconstructPlan      = errors.New("unable to reconstruct packing combination")
)

const maxTableEntries = 2_000_000

type PackBreakdown struct {
	Size  int `json:"size"`
	Count int `json:"count"`
}

type Plan struct {
	ItemsOrdered int             `json:"items_ordered"`
	TotalItems   int             `json:"total_items"`
	TotalPacks   int             `json:"total_packs"`
	Packs        []PackBreakdown `json:"packs"`
}

// NormalizePackSizes validates pack sizes, removes duplicates, and returns
// a descending-sorted slice so larger packs are evaluated first.
func NormalizePackSizes(packSizes []int) ([]int, error) {
	if len(packSizes) == 0 {
		return nil, ErrInvalidPackSizes
	}

	// seen removes duplicates to improve the time complexity of buildOptimalPackingTable.
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

// Optimize computes the fulfillment plan that meets or exceeds itemsOrdered
// with minimum overfill and, for that total, the minimum number of packs.
func Optimize(itemsOrdered int, packSizes []int) (Plan, error) {
	if itemsOrdered <= 0 {
		return Plan{}, ErrInvalidItemsOrdered
	}
	if itemsOrdered > maxInt32Value {
		return Plan{}, fmt.Errorf("%w: %d exceeds int32 max value %d", ErrInvalidItemsOrdered, itemsOrdered, maxInt32Value)
	}

	normalized, err := NormalizePackSizes(packSizes)
	if err != nil {
		return Plan{}, err
	}

	table, err := newPackingTable(itemsOrdered, normalized)
	if err != nil {
		return Plan{}, err
	}
	table.buildOptimalPackingTable()

	chosenTotal, err := table.chooseFulfillmentTotal()
	if err != nil {
		return Plan{}, err
	}

	breakdown, err := table.buildBreakdown(chosenTotal)
	if err != nil {
		return Plan{}, err
	}

	return Plan{
		ItemsOrdered: itemsOrdered,
		TotalItems:   chosenTotal,
		TotalPacks:   table.minPacks[chosenTotal],
		Packs:        breakdown,
	}, nil
}

type packingTable struct {
	itemsOrdered     int
	sortedPackSizes  []int
	fulfillmentLimit int
	minPacks         []int
	prevTotal        []int
	prevPack         []int
	unreachablePacks int
}

// newPackingTable allocates and initializes the DP state used by Optimize.
// It expects sortedPackSizes to be normalized and sorted in descending order.
// All totals start as unreachable except total=0 (the base case), and
// backtracking pointers are seeded so buildBreakdown can reconstruct a valid plan.
func newPackingTable(itemsOrdered int, sortedPackSizes []int) (packingTable, error) {
	// unset marks entries that do not have a predecessor yet.
	const unset = -1

	// Pack sizes are pre-sorted descending, so index 0 is the largest pack.
	largestPackSize := sortedPackSizes[0]
	// Last index is the smallest pack because the slice is sorted descending.
	smallestPackSize := sortedPackSizes[len(sortedPackSizes)-1]

	// We only need totals up to itemsOrdered + largestPackSize - 1.
	// Any larger total would never be the minimum valid fulfillment.
	fulfillmentLimit64 := int64(itemsOrdered) + int64(largestPackSize) - 1
	if fulfillmentLimit64 <= 0 {
		return packingTable{}, fmt.Errorf("%w: invalid fulfillment range", ErrOptimizationTooLarge)
	}
	if fulfillmentLimit64+1 > maxTableEntries {
		return packingTable{}, fmt.Errorf("%w: requires %d table entries (max %d)", ErrOptimizationTooLarge, fulfillmentLimit64+1, maxTableEntries)
	}
	fulfillmentLimit := int(fulfillmentLimit64)

	// Worst-case pack count at fulfillmentLimit: fill entirely with the smallest pack.
	worstCasePacks := fulfillmentLimit / smallestPackSize
	// Sentinel value meaning "this total is unreachable".
	unreachablePacks := worstCasePacks + 1

	minPacks := make([]int, fulfillmentLimit+1)
	prevTotal := make([]int, fulfillmentLimit+1)
	prevPack := make([]int, fulfillmentLimit+1)

	for i := range minPacks {
		minPacks[i] = unreachablePacks
		prevTotal[i] = unset
		prevPack[i] = unset
	}

	// Base case: reaching total=0 requires zero packs with no predecessor.
	minPacks[0] = 0
	prevTotal[0] = 0
	prevPack[0] = 0

	return packingTable{
		itemsOrdered:     itemsOrdered,
		sortedPackSizes:  sortedPackSizes,
		fulfillmentLimit: fulfillmentLimit,
		minPacks:         minPacks,
		prevTotal:        prevTotal,
		prevPack:         prevPack,
		unreachablePacks: unreachablePacks,
	}, nil
}

// buildOptimalPackingTable populates minPacks and backtracking pointers for
// every reachable total up to fulfillmentLimit.
func (t *packingTable) buildOptimalPackingTable() {
	for total := 1; total < len(t.minPacks); total++ {
		for _, packSize := range t.sortedPackSizes {
			predecessor := total - packSize
			if predecessor < 0 || t.minPacks[predecessor] == t.unreachablePacks {
				continue
			}

			// If this pack reaches `total` using fewer packs than any previously
			// found combination, record it as the new best.
			candidate := t.minPacks[predecessor] + 1
			if candidate < t.minPacks[total] {
				t.minPacks[total] = candidate
				t.prevTotal[total] = predecessor
				t.prevPack[total] = packSize
			}
		}
	}
}

// chooseFulfillmentTotal returns the smallest reachable total that is
// at least itemsOrdered, satisfying the no-underfill constraint.
func (t *packingTable) chooseFulfillmentTotal() (int, error) {
	for total := t.itemsOrdered; total < len(t.minPacks); total++ {
		if t.minPacks[total] != t.unreachablePacks {
			// When a total that can be fulfilled is found, return it immediately. This ensures we get the closest fulfillment without going under.
			return total, nil
		}
	}
	return -1, errNoPackingPlan
}

// buildBreakdown reconstructs the chosen solution by following prevTotal and
// prevPack from chosenTotal back to zero, then groups counts by pack size.
func (t *packingTable) buildBreakdown(chosenTotal int) ([]PackBreakdown, error) {
	packCounts := make(map[int]int)
	for total := chosenTotal; total > 0; total = t.prevTotal[total] {
		packSize := t.prevPack[total]
		if packSize <= 0 {
			return nil, errReconstructPlan
		}
		packCounts[packSize]++
	}

	breakdown := make([]PackBreakdown, 0, len(packCounts))
	for _, size := range t.sortedPackSizes {
		if count := packCounts[size]; count > 0 {
			breakdown = append(breakdown, PackBreakdown{Size: size, Count: count})
		}
	}

	return breakdown, nil
}
