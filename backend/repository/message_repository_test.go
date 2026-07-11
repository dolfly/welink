package repository

import (
	"database/sql"
	"fmt"
	"testing"

	"welink/backend/pkg/db"
)

func TestDBsForUsernameReturnsOnlyContainingDatabases(t *testing.T) {
	first, err := sql.Open("sqlite", "file:repo-first?mode=memory&cache=shared")
	if err != nil {
		t.Fatal(err)
	}
	defer first.Close()
	second, err := sql.Open("sqlite", "file:repo-second?mode=memory&cache=shared")
	if err != nil {
		t.Fatal(err)
	}
	defer second.Close()

	username := "wxid_perf_test"
	tableName := db.GetTableName(username)
	if _, err := second.Exec(fmt.Sprintf("CREATE TABLE [%s] (create_time INTEGER)", tableName)); err != nil {
		t.Fatal(err)
	}

	mgr := &db.DBManager{MessageDBs: []*sql.DB{first, second}}
	repo := NewMessageRepository(mgr)

	got := repo.DBsForUsername(username)
	if len(got) != 1 || got[0] != second {
		t.Fatalf("DBsForUsername() = %v, want only second DB", got)
	}
	if got := repo.DBsForUsername("wxid_missing"); len(got) != 0 {
		t.Fatalf("missing username returned %d DBs, want 0", len(got))
	}
	indices := repo.DBIndicesForUsername(username)
	if len(indices) != 1 || indices[0] != 1 {
		t.Fatalf("DBIndicesForUsername() = %v, want [1]", indices)
	}
}
