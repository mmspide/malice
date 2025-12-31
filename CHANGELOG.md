Change Log
==========

All notable changes to this project will be documented in this file.

[v0.4.0] - 2025-12-31
---------------------
**Modernization Release - Security, Performance & Reliability**

### Added

- **Security:** Path traversal prevention with `validateAndNormalizePath()` function
- **Performance:** Elasticsearch connection pooling with singleton pattern (20-30% improvement)
- **Reliability:** Graceful shutdown with context-based cancellation
- **Error Handling:** Custom error types (ScanError, ValidationError, PluginError)
- **Concurrency:** Bounded concurrency with semaphore pattern (max 10 plugins)
- **Logging:** Structured logging throughout the application
- **Testing:** Comprehensive unit tests for new features
- **Documentation:** CODE_IMPROVEMENTS.md, BEST_PRACTICES.md, RELEASE_NOTES_v0.4.0.md

### Fixed

- Goroutine leaks in watch.go (context cancellation)
- Resource leaks in Elasticsearch connections (pooling)
- Unbounded plugin concurrency (semaphore pattern)
- Missing timeouts in scan operations (10-min scan, 2-min operation timeout)
- Confusing error messages (typed errors with context)

### Changed

- Input validation at entry points (all commands)
- Error propagation to use typed errors instead of log.Fatal
- Logging format to structured log.WithFields/WithError patterns
- Scan command to support operation-level timeouts

### Removed

- 90 lines of deprecated docker-machine code
- 50+ lines of legacy docker.go implementation
- Duplicate utility functions (consolidated in internal/util)

### Documentation

See [RELEASE_NOTES_v0.4.0.md](RELEASE_NOTES_v0.4.0.md) for complete details.

[latest]
--------

### Fixed

### Added

### Removed

### Changed

[v0.2.0] - 2016-10-08
---------------------

### Fixed

### Added

-	added support for ElasticSearch through use of **blacktop/elk**
-	add zip plugin place holder
-	add nsrl lookup plugin
-	add totalhash lookup plugin
-	Docs !!!
-	release binaries

### Removed

-	support for RethinkDB

### Changed

-	upgrade to the elastic stack 5.0.0

[v0.1.0] - 2016-08-14
---------------------

### Fixed

-	improved zsh completions to include new features

### Added

-	added the ability to watch a folder for new files and then scan them with the `watch` subcommand
-	added ability to update plugin/all plugins from source with the `--source` flag
-	ability to mark plugins as build from source only in `plugins/plugins.toml` config file
-	tini
-	gosu

### Removed

### Changed
