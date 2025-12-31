package docker

// NOTE: Legacy docker-machine code has been deprecated and removed.
// Docker-machine support was removed from official Docker in favor of Docker Desktop
// and other container runtimes. This file is kept for reference only.

// Historical context:
// - This code used docker-machine to create virtual machines for Docker
// - Docker Desktop has replaced this functionality on macOS and Windows
// - For Linux users, Docker daemon is available natively
//
// If docker-machine functionality is needed, refer to the git history:
// git log --all -p -- malice/docker/machine.go

// TODO: Consider migration guide for users still using docker-machine
