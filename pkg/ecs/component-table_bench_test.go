package ecs

import "testing"

const testEntitiesLen = 100_000

func BenchmarkComponentBoolTable_SetAndTest(b *testing.B) {
	// using a fixed maximum components length
	table := NewComponentBoolTable(1024)
	b.ReportAllocs()
	for b.Loop() {
		for i := 0; i < testEntitiesLen; i++ {
			entity := Entity(i)
			comp := ComponentId(i % 1024)
			table.Set(entity, comp)
			if !table.Test(entity, comp) {
				b.Fatalf("ByteTable: expected entity %d to have component %d set", entity, comp)
			}
		}
	}

}

func BenchmarkComponentBitTable_SetAndTest(b *testing.B) {
	// using a fixed maximum components length
	table := NewComponentBitTable(1024)
	b.ReportAllocs()
	b.ReportAllocs()
	for b.Loop() {
		for i := 0; i < testEntitiesLen; i++ {
			entity := Entity(i)
			comp := ComponentId(i % 1024)
			table.Set(entity, comp)
			if !table.Test(entity, comp) {
				b.Fatalf("BitTable: expected entity %d to have component %d set", entity, comp)
			}
		}
	}
}
