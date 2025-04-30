package ecs

import "testing"

const testEntitiesLen = 100_000
const maxComponentsLen = 1024

func BenchmarkComponentBitTable_SetAndTest(b *testing.B) {
	// using a fixed maximum components length
	table := NewComponentBitTable(maxComponentsLen)
	b.ReportAllocs()
	for b.Loop() {
		for i := 0; i < testEntitiesLen; i++ {
			entity := Entity(i)
			comp := ComponentId(i % maxComponentsLen)
			table.Create(entity)
			table.Set(entity, comp)
			if !table.Test(entity, comp) {
				b.Fatalf("BitTable: expected entity %d to have component %d set", entity, comp)
			}
		}
	}
}

func BenchmarkComponentBitTable_Delete(b *testing.B) {
	// using a fixed maximum components length
	table := NewComponentBitTable(maxComponentsLen)
	// Setup - create and set components for all entities
	for i := 0; i < testEntitiesLen; i++ {
		entity := Entity(i)
		table.Create(entity)
		table.Set(entity, ComponentId(i%maxComponentsLen))
	}

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		// Delete all entities
		for i := 0; i < testEntitiesLen; i++ {
			entity := Entity(i)
			table.Delete(entity)
		}

		// Recreate for next iteration
		for i := 0; i < testEntitiesLen; i++ {
			entity := Entity(i)
			table.Create(entity)
			table.Set(entity, ComponentId(i%maxComponentsLen))
		}
	}
}

func BenchmarkComponentBitTable_AllSet(b *testing.B) {
	table := NewComponentBitTable(maxComponentsLen)
	// Prepare entities with multiple components each
	for i := 0; i < testEntitiesLen; i++ {
		entity := Entity(i)
		table.Create(entity)
		// Give each entity 100 components
		for j := 0; j < 100; j++ {
			table.Set(entity, ComponentId(j%maxComponentsLen))
		}
	}

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		for i := 0; i < testEntitiesLen; i++ {
			entity := Entity(i)
			count := 0
			table.AllSet(entity, func(id ComponentId) bool {
				count++
				return true
			})
			if count != 100 {
				b.Fatalf("Expected 3 components, got %d", count)
			}
		}
	}
}
