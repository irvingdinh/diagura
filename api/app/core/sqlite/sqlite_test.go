package sqlite

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"localhost/app/core/sqlite/driver"
)

func testConn(t *testing.T) *cachedConn {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "test.db")
	conn, err := driver.Open(path, driver.OpenReadWrite|driver.OpenCreate|driver.OpenNoMutex)
	if err != nil {
		t.Fatalf("driver.Open: %v", err)
	}
	if err := conn.Exec("PRAGMA journal_mode = WAL"); err != nil {
		_ = conn.Close()
		t.Fatalf("PRAGMA: %v", err)
	}
	if err := conn.Exec("CREATE TABLE t (id INTEGER PRIMARY KEY, val TEXT)"); err != nil {
		_ = conn.Close()
		t.Fatalf("CREATE TABLE: %v", err)
	}
	cc := newCachedConn(conn)
	t.Cleanup(func() {
		_ = cc.close()
		_ = os.RemoveAll(dir)
	})
	return cc
}

func TestCachedConnHit(t *testing.T) {
	cc := testConn(t)

	sql := "INSERT INTO t (val) VALUES (?)"

	stmt1, err := cc.prepare(sql)
	if err != nil {
		t.Fatalf("first prepare: %v", err)
	}

	stmt2, err := cc.prepare(sql)
	if err != nil {
		t.Fatalf("second prepare: %v", err)
	}

	if stmt1 != stmt2 {
		t.Error("cache miss on identical SQL: got different statement pointers")
	}

	if cc.ll.Len() != 1 {
		t.Errorf("cache size = %d, want 1", cc.ll.Len())
	}
}

func TestCachedConnEviction(t *testing.T) {
	cc := testConn(t)

	// Fill the cache to capacity with unique statements.
	for i := 0; i < stmtCacheCapacity; i++ {
		sql := fmt.Sprintf("SELECT %d", i)
		if _, err := cc.prepare(sql); err != nil {
			t.Fatalf("prepare %d: %v", i, err)
		}
	}

	if cc.ll.Len() != stmtCacheCapacity {
		t.Fatalf("cache size = %d, want %d", cc.ll.Len(), stmtCacheCapacity)
	}

	// The first entry (SELECT 0) should be at the back (LRU).
	back := cc.ll.Back().Value.(*cacheEntry)
	if back.sql != "SELECT 0" {
		t.Fatalf("LRU entry = %q, want %q", back.sql, "SELECT 0")
	}

	// Insert one more to trigger eviction.
	overflowSQL := fmt.Sprintf("SELECT %d", stmtCacheCapacity)
	if _, err := cc.prepare(overflowSQL); err != nil {
		t.Fatalf("prepare overflow: %v", err)
	}

	if cc.ll.Len() != stmtCacheCapacity {
		t.Errorf("cache size after eviction = %d, want %d", cc.ll.Len(), stmtCacheCapacity)
	}

	// "SELECT 0" should have been evicted.
	if _, ok := cc.index["SELECT 0"]; ok {
		t.Error("evicted entry still in index")
	}

	// The overflow entry should be at the front (MRU).
	front := cc.ll.Front().Value.(*cacheEntry)
	if front.sql != overflowSQL {
		t.Errorf("MRU entry = %q, want %q", front.sql, overflowSQL)
	}
}

func TestCachedConnMoveToFront(t *testing.T) {
	cc := testConn(t)

	// Insert two entries.
	if _, err := cc.prepare("SELECT 1"); err != nil {
		t.Fatal(err)
	}
	if _, err := cc.prepare("SELECT 2"); err != nil {
		t.Fatal(err)
	}

	// "SELECT 2" is at front, "SELECT 1" at back.
	if cc.ll.Front().Value.(*cacheEntry).sql != "SELECT 2" {
		t.Fatal("expected SELECT 2 at front")
	}

	// Access "SELECT 1" to move it to front.
	if _, err := cc.prepare("SELECT 1"); err != nil {
		t.Fatal(err)
	}

	if cc.ll.Front().Value.(*cacheEntry).sql != "SELECT 1" {
		t.Error("SELECT 1 should be at front after re-access")
	}
	if cc.ll.Back().Value.(*cacheEntry).sql != "SELECT 2" {
		t.Error("SELECT 2 should be at back after SELECT 1 re-access")
	}
}
