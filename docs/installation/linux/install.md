Installation on Linux (Ubuntu 22.04 LTS)
=======================================

These instructions cover installing Docker, preparing the system for Elasticsearch (used by the Malice stack), and installing Malice either from a pre-built binary or from source on Ubuntu 22.04 LTS.

Prerequisites
-------------

- A 64-bit Ubuntu 22.04 LTS installation
- A non-root user with `sudo` privileges

1) Install Docker (recommended way)
-----------------------------------

Follow the official Docker Engine install steps for Ubuntu 22.04. Example commands:

```bash
sudo apt-get update
sudo apt-get install -y ca-certificates curl gnupg lsb-release
sudo mkdir -p /etc/apt/keyrings
curl -fsSL https://download.docker.com/linux/ubuntu/gpg | sudo gpg --dearmor -o /etc/apt/keyrings/docker.gpg
echo \
	"deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.gpg] https://download.docker.com/linux/ubuntu \
	$(lsb_release -cs) stable" | sudo tee /etc/apt/sources.list.d/docker.list > /dev/null
sudo apt-get update
sudo apt-get install -y docker-ce docker-ce-cli containerd.io docker-compose-plugin
sudo systemctl enable --now docker
sudo usermod -aG docker $USER
echo "You may need to log out and back in for group changes to take effect."
```

Notes:
- The package names and repository above are the official Docker packages. `docker-engine` and the old apt-key method are deprecated.

2) Prepare system for Elasticsearch
-----------------------------------

Elasticsearch (used by the Malice ELK stack) requires increasing the maximum memory map areas. Run:

```bash
sudo sysctl -w vm.max_map_count=262144
echo "vm.max_map_count=262144" | sudo tee /etc/sysctl.d/99-elasticsearch.conf
```

If you run Elasticsearch in Docker, make sure the container has enough memory and appropriate Java heap settings (e.g. `ES_JAVA_OPTS="-Xms2g -Xmx2g"`) when starting.

3) Download and install pre-compiled Malice binary
--------------------------------------------------

Grab the latest release from GitHub and install to `/usr/local/bin`:

```bash
# Replace v0.4.0 with the desired release tag if needed
VERSION=v0.4.0
ARCH=$(dpkg --print-architecture)
wget https://github.com/mmspide/malice/releases/download/${VERSION}/malice_${VERSION}_linux_${ARCH}.tar.gz -O /tmp/malice.tar.gz
sudo tar -xzf /tmp/malice.tar.gz -C /usr/local/bin/
sudo chmod +x /usr/local/bin/malice

# Verify
malice --version
```

Uninstall (binary):

```bash
sudo rm -f /usr/local/bin/malice
rm -rf ~/.malice
```

4) Build and install from source (Go 1.21+)
-------------------------------------------

Install Go (recommended: official tarball). Example for Go 1.21.x:

```bash
GO_VERSION=1.21.0
ARCH=$(dpkg --print-architecture)
wget https://go.dev/dl/go${GO_VERSION}.linux-${ARCH}.tar.gz -O /tmp/go.tar.gz
sudo rm -rf /usr/local/go && sudo tar -C /usr/local -xzf /tmp/go.tar.gz
export PATH=$PATH:/usr/local/go/bin
export GOPATH=$HOME/go
export PATH=$PATH:$GOPATH/bin

# Persist the PATH changes by adding the exports to ~/.profile or ~/.bashrc
```

Install build tools and clone the repo:

```bash
sudo apt-get install -y build-essential git
git clone https://github.com/mmspide/malice.git $GOPATH/src/github.com/mmspide/malice
cd $GOPATH/src/github.com/mmspide/malice
go install github.com/mmspide/malice@latest

# The 'malice' binary will be available in $GOPATH/bin or $HOME/go/bin
```

Uninstall (source):

```bash
rm -rf $GOPATH/src/github.com/mmspide/malice
rm -f $(which malice) || rm -f $GOPATH/bin/malice
rm -rf ~/.malice
```

5) Start the ELK stack (docker-compose)
---------------------------------------

If you prefer Docker Compose, use the provided `docker-compose.yml` (updated for v3.8+). Example:

```bash
# From project root
docker compose up -d

# Check services
docker compose ps
```

6) Additional notes
-------------------

- If you encounter issues connecting to Elasticsearch from containers, confirm `vm.max_map_count` is set and the host has sufficient memory.
- Use `journalctl -u docker.service` and `docker logs <container>` to troubleshoot.
- For production use, secure Elasticsearch with authentication and TLS as required.

References
----------
- Docker Engine install: https://docs.docker.com/engine/install/ubuntu/
- Elasticsearch Docker: https://www.elastic.co/guide/en/elasticsearch/reference/current/docker.html
- Go installs: https://go.dev/doc/install
