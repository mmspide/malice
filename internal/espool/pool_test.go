package espool

import (
	"testing"

	"github.com/malice-plugins/pkgs/database/elasticsearch"
)

func TestPoolInit(t *testing.T) {
	// Reset pool before test
	pool := &Pool{}

	// Test with nil database
	err := pool.Init(nil)
	if err != ErrNilDatabase {
		t.Errorf("Init(nil) = %v, want %v", err, ErrNilDatabase)
	}

	// Test that only one initialization happens
	db1 := &elasticsearch.Database{
		URL:   "http://localhost:9200",
		Index: "test1",
	}

	err1 := pool.Init(db1)
	// First init might succeed or fail depending on Elasticsearch availability
	// but subsequent inits should not change it

	db2 := &elasticsearch.Database{
		URL:   "http://localhost:9201",
		Index: "test2",
	}

	err2 := pool.Init(db2)

	// Both should return same error or success state
	if (err1 != nil) != (err2 != nil) {
		t.Errorf("Init should be idempotent: first error=%v, second error=%v", err1, err2)
	}
}

func TestPoolGet(t *testing.T) {
	pool := &Pool{}

	// Get without initialization should return error
	_, err := pool.Get()
	if err == nil {
		t.Error("Get() without initialization should return error")
	}

	if err != ErrNotInitialized {
		t.Errorf("Get() without initialization = %v, want %v", err, ErrNotInitialized)
	}
}

func TestGlobalPool(t *testing.T) {
	// Reset global pool
	ResetGlobal()

	// Get without initialization should return error
	_, err := GetGlobal()
	if err == nil {
		t.Error("GetGlobal() without initialization should return error")
	}

	if err != ErrNotInitialized {
		t.Errorf("GetGlobal() without initialization = %v, want %v", err, ErrNotInitialized)
	}

	// Test with nil database
	err = InitGlobal(nil)
	if err != ErrNilDatabase {
		t.Errorf("InitGlobal(nil) = %v, want %v", err, ErrNilDatabase)
	}
}
