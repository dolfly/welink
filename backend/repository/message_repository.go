package repository

import (
	"database/sql"
	"fmt"
	"log"
	"sync"
	"welink/backend/pkg/db"
)

type MessageRepository struct {
	dbMgr    *db.DBManager
	tableMap map[string][]int // tableName -> []dbIndex
	mu       sync.RWMutex
}

func NewMessageRepository(mgr *db.DBManager) *MessageRepository {
	repo := &MessageRepository{
		dbMgr:    mgr,
		tableMap: make(map[string][]int),
	}
	repo.buildIndex()
	return repo
}

// buildIndex 启动时扫描所有库，建立“表名 -> 数据库索引”的映射
func (r *MessageRepository) buildIndex() {
	log.Println("Building message table index...")
	tableMap := make(map[string][]int)
	for idx, mdb := range r.dbMgr.MessageDBs {
		rows, err := mdb.Query("SELECT name FROM sqlite_master WHERE type='table' AND name LIKE 'Msg_%'")
		if err != nil {
			continue
		}
		for rows.Next() {
			var name string
			if err := rows.Scan(&name); err == nil {
				tableMap[name] = append(tableMap[name], idx)
			}
		}
		rows.Close()
	}
	r.mu.Lock()
	r.tableMap = tableMap
	r.mu.Unlock()
	log.Printf("Index built: %d tables found.", len(tableMap))
}

// DBsForUsername returns only the message databases that contain this user's
// message table. Most tables live in a small subset of message_N.db files, so
// callers should use this instead of probing every database and handling
// "no such table" errors on the hot path.
func (r *MessageRepository) DBsForUsername(username string) []*sql.DB {
	indices := r.DBIndicesForUsername(username)
	dbs := make([]*sql.DB, 0, len(indices))
	for _, idx := range indices {
		if idx >= 0 && idx < len(r.dbMgr.MessageDBs) {
			dbs = append(dbs, r.dbMgr.MessageDBs[idx])
		}
	}
	return dbs
}

// DBIndicesForUsername returns the stable MessageDBs indices containing the
// user's message table. Callers that also need per-database metadata (for
// example Name2Id maps) can retain the original database index without
// probing every message database.
func (r *MessageRepository) DBIndicesForUsername(username string) []int {
	tableName := db.GetTableName(username)
	r.mu.RLock()
	indices := append([]int(nil), r.tableMap[tableName]...)
	r.mu.RUnlock()
	return indices
}

type UserMsgStats struct {
	TotalCount int64
	FirstTime  int64
	LastTime   int64
}

func (r *MessageRepository) GetUserStats(username string) UserMsgStats {
	tableName := db.GetTableName(username)
	r.mu.RLock()
	dbIndices, ok := r.tableMap[tableName]
	r.mu.RUnlock()

	if !ok {
		return UserMsgStats{}
	}

	var stats UserMsgStats
	stats.FirstTime = 9999999999
	var mu sync.Mutex
	var wg sync.WaitGroup

	for _, idx := range dbIndices {
		wg.Add(1)
		go func(mdb *sql.DB) {
			defer wg.Done()
			query := fmt.Sprintf("SELECT COUNT(*), MIN(create_time), MAX(create_time) FROM [%s]", tableName)
			var count int64
			var minT, maxT sql.NullInt64
			err := mdb.QueryRow(query).Scan(&count, &minT, &maxT)
			if err != nil || count == 0 {
				return
			}

			mu.Lock()
			stats.TotalCount += count
			if minT.Valid && minT.Int64 < stats.FirstTime {
				stats.FirstTime = minT.Int64
			}
			if maxT.Valid && maxT.Int64 > stats.LastTime {
				stats.LastTime = maxT.Int64
			}
			mu.Unlock()
		}(r.dbMgr.MessageDBs[idx])
	}
	wg.Wait()

	if stats.FirstTime == 9999999999 {
		stats.FirstTime = 0
	}
	return stats
}
