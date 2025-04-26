/*
This Source Code Form is subject to the terms of the Mozilla
Public License, v. 2.0. If a copy of the MPL was not distributed
with this file, You can obtain one at http://mozilla.org/MPL/2.0/.

===-===-===-===-===-===-===-===-===-===
Donations during this file development:
-===-===-===-===-===-===-===-===-===-===

none :)

Thank you for your support!
*/

package ecs

import (
	"math/bits"
	"testing"
)

func TestNewComponentBitTable(t *testing.T) {
	// Test with different max component sizes
	tests := []struct {
		name               string
		maxComponentsLen   int
		expectedBitsetSize int
	}{
		{"Small", 10, 1},
		{"Medium", 64, 1},
		{"Large", 192, 3},
		{"Below", 172, 3},
		{"Above", 200, 4},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			table := NewComponentBitTable(tt.maxComponentsLen)

			if table.bitsetSize != tt.expectedBitsetSize {
				t.Errorf("Expected bitsetSize %d, got %d", tt.expectedBitsetSize, table.bitsetSize)
			}

			if cap(table.bits) != initialBookSize {
				t.Errorf("Expected %d preallocated chunks, got %d", initialBookSize, cap(table.bits))
			}
		})
	}
}

func TestComponentBitTable_Set(t *testing.T) {
	table := NewComponentBitTable(100)

	// Set a bit for a new entity
	entity1 := Entity(1)
	table.Set(entity1, ComponentId(5))

	// Verify bit was set
	bitsId, ok := table.lookup[entity1]
	if !ok {
		t.Fatalf("Entity %d not found in lookup", entity1)
	}

	chunkId := bitsId / pageSize
	bitsetId := bitsId % pageSize
	offset := int(ComponentId(5) / bits.UintSize)
	mask := uint(1 << (ComponentId(5) % bits.UintSize))

	if (table.bits[chunkId][bitsetId+offset] & mask) == 0 {
		t.Errorf("Expected bit to be set for entity %d, component %d", entity1, 5)
	}

	// Set multiple bits for same entity
	table.Set(entity1, ComponentId(10))
	table.Set(entity1, ComponentId(63))
	table.Set(entity1, ComponentId(64))

	// Set bits for a different entity
	entity2 := Entity(2)
	table.Set(entity2, ComponentId(5))
}

func TestComponentBitTable_Unset(t *testing.T) {
	table := NewComponentBitTable(100)
	entity := Entity(1)

	// Set and then unset
	table.Set(entity, ComponentId(5))
	table.Set(entity, ComponentId(10))
	table.Unset(entity, ComponentId(5))

	// Verify bit was unset
	bitsId := table.lookup[entity]
	chunkId := bitsId / pageSize
	bitsetId := bitsId % pageSize
	offset := int(ComponentId(5) / bits.UintSize)
	mask := uint(1 << (ComponentId(5) % bits.UintSize))

	if (table.bits[chunkId][bitsetId+offset] & mask) != 0 {
		t.Errorf("Expected bit to be unset for entity %d, component %d", entity, 5)
	}

	// Verify other bit is still set
	offset = int(ComponentId(10) / bits.UintSize)
	mask = uint(1 << (ComponentId(10) % bits.UintSize))

	if (table.bits[chunkId][bitsetId+offset] & mask) == 0 {
		t.Errorf("Expected bit to still be set for entity %d, component %d", entity, 10)
	}
}

func TestComponentBitTable_Test(t *testing.T) {
	table := NewComponentBitTable(100)

	// Test for non-existent entity
	if table.Test(Entity(999), ComponentId(5)) {
		t.Error("Test should return false for non-existent entity")
	}

	// Set up an entity with some components
	entity := Entity(42)
	table.Set(entity, ComponentId(5))
	table.Set(entity, ComponentId(64))

	// Test for set components
	if !table.Test(entity, ComponentId(5)) {
		t.Error("Test should return true for set component 5")
	}
	if !table.Test(entity, ComponentId(64)) {
		t.Error("Test should return true for set component 64")
	}

	// Test for unset component
	if table.Test(entity, ComponentId(10)) {
		t.Error("Test should return false for unset component 10")
	}

	// Test after unsetting a component
	table.Unset(entity, ComponentId(5))
	if table.Test(entity, ComponentId(5)) {
		t.Error("Test should return false after component is unset")
	}

	// Test components at boundaries
	entity2 := Entity(43)
	table.Set(entity2, ComponentId(0))
	table.Set(entity2, ComponentId(64)) // Last bit in first uint
	table.Set(entity2, ComponentId(65)) // First bit in second uint

	if !table.Test(entity2, ComponentId(0)) {
		t.Error("Test should return true for set component at uint boundary (0)")
	}
	if !table.Test(entity2, ComponentId(64)) {
		t.Error("Test should return true for set component at uint boundary (64)")
	}
	if !table.Test(entity2, ComponentId(65)) {
		t.Error("Test should return true for set component at uint boundary (65)")
	}
}

func TestComponentBitTable_AllSet(t *testing.T) {
	table := NewComponentBitTable(200)
	entity := Entity(1)

	// Set several components
	expectedComponents := []ComponentId{5, 10, 64, 128, 199}
	for _, id := range expectedComponents {
		table.Set(entity, id)
	}

	// Use AllSet to collect components
	var foundComponents []ComponentId
	table.AllSet(entity, func(id ComponentId) bool {
		foundComponents = append(foundComponents, id)
		return true
	})

	// Verify all components were found
	if len(foundComponents) != len(expectedComponents) {
		t.Errorf("Expected %d components, found %d", len(expectedComponents), len(foundComponents))
	}

	// Check each component
	componentMap := make(map[ComponentId]bool)
	for _, id := range foundComponents {
		componentMap[id] = true
	}

	for _, id := range expectedComponents {
		if !componentMap[id] {
			t.Errorf("Component %d not found in AllSet results", id)
		}
	}

	// Test early termination
	count := 0
	table.AllSet(entity, func(id ComponentId) bool {
		count++
		return count < 3 // Stop after finding 2 components
	})

	if count != 3 {
		t.Errorf("Early termination didn't work as expected. Count: %d", count)
	}
}

func TestComponentBitTable_extend(t *testing.T) {
	// Create a table with a small chunk size for testing
	table := NewComponentBitTable(20)
	// Set bits to force extension
	for i := 0; i < pageSize*table.bitsetSize; i++ {
		table.Set(Entity(i), ComponentId(1))
	}

	if len(table.bits) > 1 {
		t.Errorf("Expected table to be not extended, got %d chunks", len(table.bits))
	}

	table.Set(Entity(pageSize*table.bitsetSize), ComponentId(1))

	if len(table.bits) != 2 {
		t.Errorf("Expected table to extend up to 2 chunks, got %d chunks", len(table.bits))
	}

	for i := pageSize * table.bitsetSize; i < pageSize*table.bitsetSize*2; i++ {
		table.Set(Entity(i), ComponentId(1))
	}

	if len(table.bits) != 2 {
		t.Errorf("Expected table to extend up to 2 chunks, got %d chunks", len(table.bits))
	}

	table.Set(Entity(pageSize*table.bitsetSize*2), ComponentId(1))

	if len(table.bits) != 3 {
		t.Errorf("Expected table to extend up to 3 chunks, got %d chunks", len(table.bits))
	}
}

func TestComponentBitTable_EdgeCases(t *testing.T) {
	table := NewComponentBitTable(100)

	// Test AllSet on non-existent entity
	called := false
	table.AllSet(Entity(999), func(id ComponentId) bool {
		called = true
		return true
	})

	if called {
		t.Errorf("AllSet callback should not be called for non-existent entity")
	}

	// Test setting across bit boundaries
	entity := Entity(42)
	table.Set(entity, ComponentId(0))  // Last bit in first uint
	table.Set(entity, ComponentId(64)) // Last bit in first uint
	table.Set(entity, ComponentId(65)) // First bit in second uint

	// Verify both bits
	var found []ComponentId
	table.AllSet(entity, func(id ComponentId) bool {
		found = append(found, id)
		return true
	})

	if len(found) != 3 || !contains(found, ComponentId(0)) || !contains(found, ComponentId(64)) || !contains(found, ComponentId(65)) {
		t.Errorf("Expected components 0, 64 and 65, got %v", found)
	}
}

// Helper function to check if slice contains a value
func contains(s []ComponentId, id ComponentId) bool {
	for _, v := range s {
		if v == id {
			return true
		}
	}
	return false
}
