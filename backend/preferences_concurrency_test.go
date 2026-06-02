package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"testing"
)

// TestPreferencesConcurrentWrites 验证 H1 修复：
//   - 并发的 updatePreferences / savePreferences 不会写坏文件（始终是合法 JSON）
//   - 用 updatePreferences 做"读-改-写"不会丢更新（last-write-wins 竞态已被锁消除）
//
// 配合 `go test -race` 运行可同时检测数据竞争。
func TestPreferencesConcurrentWrites(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "preferences.json")
	t.Setenv("PREFERENCES_PATH", path)

	// 初始写一份基线
	if err := savePreferences(defaultPreferences()); err != nil {
		t.Fatalf("初始化 preferences 失败：%v", err)
	}

	const writers = 16
	const itersPerWriter = 30

	var wg sync.WaitGroup

	// 一组 writer 各自反复 updatePreferences，往 BlockedUsers 里追加自己的唯一标记。
	// 若读-改-写有丢更新，最终条数会少于 writers*itersPerWriter。
	for w := 0; w < writers; w++ {
		wg.Add(1)
		go func(w int) {
			defer wg.Done()
			for i := 0; i < itersPerWriter; i++ {
				marker := mkMarker(w, i)
				if _, err := updatePreferences(func(p *Preferences) {
					p.BlockedUsers = append(p.BlockedUsers, marker)
				}); err != nil {
					t.Errorf("updatePreferences 失败：%v", err)
					return
				}
			}
		}(w)
	}

	// 另一组并发读，确保读到的永远是合法 JSON（非原子写会读到截断/空文件）。
	var stop sync.WaitGroup
	done := make(chan struct{})
	for r := 0; r < 4; r++ {
		stop.Add(1)
		go func() {
			defer stop.Done()
			for {
				select {
				case <-done:
					return
				default:
					data, err := os.ReadFile(path)
					if err != nil {
						continue // rename 间隙偶发 ENOENT 可接受
					}
					var p Preferences
					if err := json.Unmarshal(data, &p); err != nil {
						t.Errorf("并发读到损坏的 preferences.json：%v", err)
						return
					}
				}
			}
		}()
	}

	wg.Wait()
	close(done)
	stop.Wait()

	// 校验无丢更新：所有 marker 都在
	final := loadPreferences()
	got := make(map[string]bool, len(final.BlockedUsers))
	for _, m := range final.BlockedUsers {
		got[m] = true
	}
	missing := 0
	for w := 0; w < writers; w++ {
		for i := 0; i < itersPerWriter; i++ {
			if !got[mkMarker(w, i)] {
				missing++
			}
		}
	}
	if missing > 0 {
		t.Errorf("检测到丢更新：%d/%d 个标记缺失（读-改-写竞态未被消除）", missing, writers*itersPerWriter)
	}
}

func mkMarker(w, i int) string {
	return "u-" + itoa(w) + "-" + itoa(i)
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var b [20]byte
	pos := len(b)
	for n > 0 {
		pos--
		b[pos] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		pos--
		b[pos] = '-'
	}
	return string(b[pos:])
}
