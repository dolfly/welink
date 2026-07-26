package service

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestStoreBoundedEvictsOldestEntry(t *testing.T) {
	cache := make(map[string]int)
	order := make([]string, 0, contactAnalysisCacheLimit)

	for i := 0; i <= contactAnalysisCacheLimit; i++ {
		key := string(rune('a' + i))
		storeBounded(cache, &order, key, i, contactAnalysisCacheLimit)
	}

	if len(cache) != contactAnalysisCacheLimit {
		t.Fatalf("cache size = %d, want %d", len(cache), contactAnalysisCacheLimit)
	}
	if len(order) != contactAnalysisCacheLimit {
		t.Fatalf("order size = %d, want %d", len(order), contactAnalysisCacheLimit)
	}
	if _, ok := cache["a"]; ok {
		t.Fatal("oldest cache entry was not evicted")
	}
}

func TestStoreBoundedRefreshesExistingEntry(t *testing.T) {
	cache := map[string]int{"first": 1, "second": 2}
	order := []string{"first", "second"}

	storeBounded(cache, &order, "first", 3, contactAnalysisCacheLimit)

	if got := cache["first"]; got != 3 {
		t.Fatalf("refreshed value = %d, want 3", got)
	}
	if len(order) != 2 || order[0] != "second" || order[1] != "first" {
		t.Fatalf("refresh order = %v, want [second first]", order)
	}
}

func TestAnalysisFlightGroupCoalescesConcurrentCalls(t *testing.T) {
	var group analysisFlightGroup
	var executions int32
	start := make(chan struct{})
	const callers = 8

	var wg sync.WaitGroup
	wg.Add(callers)
	results := make(chan int, callers)
	for i := 0; i < callers; i++ {
		go func() {
			defer wg.Done()
			<-start
			value, err, _ := group.Do("same-key", func() (interface{}, error) {
				atomic.AddInt32(&executions, 1)
				time.Sleep(10 * time.Millisecond)
				return 42, nil
			})
			if err != nil {
				t.Errorf("Do returned error: %v", err)
				return
			}
			results <- value.(int)
		}()
	}
	close(start)
	wg.Wait()
	close(results)

	if got := atomic.LoadInt32(&executions); got != 1 {
		t.Fatalf("executions = %d, want 1", got)
	}
	for result := range results {
		if result != 42 {
			t.Fatalf("result = %d, want 42", result)
		}
	}
}
