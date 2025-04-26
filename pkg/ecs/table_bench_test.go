package ecs

import "testing"

func BenchmarkComponentByteTable_SetAndTest(b *testing.B) {
	// using a fixed maximum components length
	table := NewComponentByteTable(1024)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		entity := Entity(i)
		comp := ComponentId(i % 1024)
		table.Set(entity, comp)
		if !table.Test(entity, comp) {
			b.Fatalf("ByteTable: expected entity %d to have component %d set", entity, comp)
		}
	}
}

func BenchmarkComponentBitTable_SetAndTest(b *testing.B) {
	// using a fixed maximum components length
	table := NewComponentBitTable(1024)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		entity := Entity(i)
		comp := ComponentId(i % 1024)
		table.Set(entity, comp)
		if !table.Test(entity, comp) {
			b.Fatalf("BitTable: expected entity %d to have component %d set", entity, comp)
		}
	}
}
