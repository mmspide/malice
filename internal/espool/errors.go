package espool

import "fmt"

// ErrNilDatabase is returned when nil database is passed
var ErrNilDatabase = fmt.Errorf("database cannot be nil")

// ErrNotInitialized is returned when pool is accessed before initialization
var ErrNotInitialized = fmt.Errorf("elasticsearch pool not initialized")
