package asnlookup2

import (
	"context"
	"testing"
)

func TestParseASN(t *testing.T) {
	tests := []struct {
		input    string
		expected uint32
		hasError bool
	}{
		{"AS1", 1, false},
		{"AS10000", 10000, false},
		{"as100", 100, false}, // 大小写不敏感
		{"AS", 0, true},
		{"ASabc", 0, true},
		{"100", 0, true}, // 缺少AS前缀
	}

	for _, tt := range tests {
		result, err := ParseASN(tt.input)
		if tt.hasError && err == nil {
			t.Errorf("Expected error for input %s", tt.input)
		}
		if !tt.hasError && err != nil {
			t.Errorf("Unexpected error for input %s: %v", tt.input, err)
		}
		if !tt.hasError && result != tt.expected {
			t.Errorf("For input %s, expected %d, got %d", tt.input, tt.expected, result)
		}
	}
}

func TestLookupIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	config := DefaultConfig()
	config.DataDir = "./test_data"

	lookup, err := New(config)
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()
	if err := lookup.Start(ctx); err != nil {
		t.Fatal(err)
	}

	// 等待加载数据
	lookup.WaitForReady(context.Background())

	// 测试查询
	asn, err := ParseASN("AS214955")
	if err != nil {
		t.Fatal(err)
	}

	record, err := lookup.Query(asn)
	if err != nil {
		t.Error("Query failed:", err)
	}

	if record == nil {
		t.Error("Expected record for AS214955, got nil")
	}
	if record.Name != "Weikang Zeng" { // TODO: if future data changes, update this test case accordingly
		t.Errorf("Expected name 'Weikang Zeng', got %s", record.Name)
	}

	// 测试批量查询
	asns := []uint32{1, 10, 100}
	results, err := lookup.BatchQuery(asns)
	if err != nil {
		t.Error("Batch query failed:", err)
	}

	if len(results) == 0 {
		t.Error("Expected some results from batch query")
	}

	// 检查统计信息
	stats := lookup.Stats()
	if stats.MemorySize > config.MaxMemoryItems {
		t.Errorf("Memory size %d exceeds limit %d", stats.MemorySize, config.MaxMemoryItems)
	}

	lookup.Stop()
}
