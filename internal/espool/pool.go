package espool

import (
	"sync"

	"github.com/malice-plugins/pkgs/database/elasticsearch"
	log "github.com/sirupsen/logrus"
)

// Pool manages Elasticsearch client connections with singleton pattern
type Pool struct {
	client *elasticsearch.Database
	once   sync.Once
	mu     sync.RWMutex
	err    error
}

var (
	// globalPool is the singleton instance
	globalPool = &Pool{}
)

// Init initializes the Elasticsearch client with the given database configuration
func (p *Pool) Init(db *elasticsearch.Database) error {
	var initErr error
	p.once.Do(func() {
		p.mu.Lock()
		defer p.mu.Unlock()

		if db == nil {
			initErr = ErrNilDatabase
			p.err = initErr
			return
		}

		// Initialize the Elasticsearch connection
		if err := db.Init(); err != nil {
			initErr = err
			p.err = err
			log.WithError(err).Error("failed to initialize elasticsearch client")
			return
		}

		p.client = db
		log.WithFields(log.Fields{
			"url":   db.URL,
			"index": db.Index,
		}).Debug("elasticsearch client initialized")
	})

	return initErr
}

// Get returns the singleton Elasticsearch client
func (p *Pool) Get() (*elasticsearch.Database, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if p.err != nil {
		return nil, p.err
	}

	if p.client == nil {
		return nil, ErrNotInitialized
	}

	return p.client, nil
}

// Reset clears the singleton (useful for testing)
func (p *Pool) Reset() {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.client = nil
	p.err = nil
	p.once = sync.Once{}
}

// Global pool accessors

// InitGlobal initializes the global Elasticsearch pool
func InitGlobal(db *elasticsearch.Database) error {
	return globalPool.Init(db)
}

// GetGlobal returns the global Elasticsearch client
func GetGlobal() (*elasticsearch.Database, error) {
	return globalPool.Get()
}

// ResetGlobal clears the global pool (for testing)
func ResetGlobal() {
	globalPool.Reset()
}
