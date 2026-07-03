# Configuration Strategy

## Overview

Configuration must be flexible for both local development and production cloud deployment. The strategy uses a three-tier hierarchy: defaults, YAML file, and environment variables.

## Configuration Hierarchy

```
Priority (Highest to Lowest):
┌─────────────────────────────┐
│ Environment Variables       │  HIGHEST: CI/CD, containers, cloud
├─────────────────────────────┤
│ YAML Configuration File     │  MIDDLE: Local customization
├─────────────────────────────┤
│ Hardcoded Defaults in Code  │  LOWEST: Safe fallback
└─────────────────────────────┘
```

## YAML Configuration File

### Location

```
server/configs/server.yml
```

### Default Content

```yaml
# Server networking
bind_address: "0.0.0.0:9000"
port: 9000

# Tick configuration
tick_rate: 20          # Ticks per second

# Server capacity
max_players: 1000      # Maximum concurrent players
max_rooms: 100         # Maximum concurrent rooms

# Logging
log_level: debug       # debug, info, warn, error
```

### Design Rationale

- **YAML over JSON**: More readable, fewer syntax errors
- **No secrets**: Only runtime tuning parameters
- **Reasonable defaults**: Server works without YAML file
- **Comments supported**: Future documentation in config

### Future Extensions

```yaml
# Room configuration
default_room_size: 5
min_players_to_start: 2
max_game_duration: 3600

# Network tuning
socket_buffer_size: 4096
read_timeout: 1000
packet_max_size: 4096

# Database (future)
database_url: "postgres://localhost/mabo"
database_pool_size: 10

# Observability
metrics_enabled: true
metrics_port: 9090
```

## Environment Variables

### Current Variables

```
SERVER_BIND_ADDRESS    # Override bind_address
SERVER_PORT            # Override port
SERVER_TICK_RATE       # Override tick_rate
SERVER_MAX_PLAYERS     # Override max_players
SERVER_MAX_ROOMS       # Override max_rooms
LOG_LEVEL              # Override log_level
```

### Naming Convention

```
PREFIX_COMPONENT_SETTING

SERVER_     → Application prefix (allows multiple servers if needed)
TICK_RATE   → Config key in uppercase with underscores
```

### Example Usage

```bash
# Development: Use defaults + YAML
go run ./cmd/server

# Container: Override specific settings
docker run -e SERVER_PORT=9000 -e LOG_LEVEL=info mabo-server

# Kubernetes: Each replica gets different config
kubectl set env deployment/mabo-server SERVER_PORT=9000
kubectl set env deployment/mabo-server LOG_LEVEL=warn
```

## Loading Strategy

### Load Order

```
1. Create defaults in memory
   └─ Port: 9000, Tick Rate: 20, etc.

2. Load YAML file (if exists)
   └─ Overlay on defaults

3. Apply environment variables
   └─ Final override layer

4. Validate configuration
   └─ Check ranges, consistency
```

### Code Implementation

```go
// Load YAML
cfg, err := config.Load("configs/server.yml")

// Internally:
// 1. Set defaults
// 2. Unmarshal YAML over defaults
// 3. Apply environment variables
// 4. Validate
```

## Configuration Validation

### Constraints

```
Port:           1-65535
Tick Rate:      1-120 TPS
Max Players:    1-10000
Max Rooms:      1-1000
Log Level:      debug|info|warn|error
Bind Address:   Valid IP:port
```

### Validation Logic

```go
if cfg.Port < 1 || cfg.Port > 65535 {
    return fmt.Errorf("invalid port: %d", cfg.Port)
}

if cfg.TickRate < 1 || cfg.TickRate > 120 {
    return fmt.Errorf("invalid tick rate: %d", cfg.TickRate)
}
```

### Error Handling

If configuration is invalid:
```
1. Log specific error message
2. Return error to caller
3. Server refuses to start
4. Exit with code 1
```

Prevents silent failures in production.

## Local Development

### Workflow

```
1. Clone repository
2. cd server
3. make run
   └─ Loads configs/server.yml
   └─ Uses environment defaults if not set
   └─ Server starts on port 9000
```

### Custom Local Config

```yaml
# server/configs/server.yml (local)
port: 9001              # Different port
log_level: debug        # Verbose logging
tick_rate: 10           # Slower for testing
```

## Docker Deployment

### Dockerfile Configuration

```dockerfile
# Use defaults baked into image
# ENV variables passed at runtime

# Local development
docker run -p 9000:9000 mabo-server

# Production
docker run \
  -e SERVER_PORT=9000 \
  -e LOG_LEVEL=warn \
  -e SERVER_MAX_PLAYERS=1000 \
  mabo-server
```

### Docker Compose

```yaml
version: '3.8'
services:
  game-server:
    build:
      context: ../server
    ports:
      - "9000:9000"
    environment:
      SERVER_PORT: 9000
      LOG_LEVEL: info
      SERVER_MAX_PLAYERS: 1000
    restart: unless-stopped
```

## Cloud Deployment

### Kubernetes ConfigMap (future)

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: mabo-config
data:
  bind_address: "0.0.0.0:9000"
  tick_rate: "20"
  max_players: "1000"
```

### Kubernetes Deployment

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: mabo-server
spec:
  replicas: 3
  template:
    spec:
      containers:
      - name: server
        image: mabo-server:latest
        env:
        - name: SERVER_PORT
          value: "9000"
        - name: LOG_LEVEL
          value: "warn"
        - name: DATABASE_URL
          valueFrom:
            secretKeyRef:
              name: db-secret
              key: url
```

### Secrets Management (future)

```bash
# Store sensitive data in external secret system
# (AWS Secrets Manager, Vault, K8s Secrets)

# Never commit secrets to repository
# Example .gitignore:
.env
secrets/
*.key
```

## Configuration Metadata

### Current

```go
type Config struct {
    BindAddress string `yaml:"bind_address"`
    Port        int    `yaml:"port"`
    TickRate    int    `yaml:"tick_rate"`
    MaxPlayers  int    `yaml:"max_players"`
    MaxRooms    int    `yaml:"max_rooms"`
    LogLevel    string `yaml:"log_level"`
}
```

### Future Extensions

```go
type Config struct {
    // ... existing fields ...
    
    Database struct {
        URL      string `yaml:"url" env:"DATABASE_URL"`
        PoolSize int    `yaml:"pool_size" env:"DATABASE_POOL_SIZE"`
    } `yaml:"database"`
    
    Observability struct {
        MetricsEnabled bool   `yaml:"metrics_enabled"`
        MetricsPort    int    `yaml:"metrics_port"`
        TracingEnabled bool   `yaml:"tracing_enabled"`
    } `yaml:"observability"`
}
```

## Configuration Documentation

### For Operators

```
SERVER_PORT (default: 9000)
  The UDP port the server listens on.
  Must be unique per server instance.
  Requires > 1024 to run as non-root.

SERVER_TICK_RATE (default: 20)
  Game simulation tick rate in Hz.
  Higher = more responsive, more CPU
  Range: 1-120

SERVER_MAX_PLAYERS (default: 1000)
  Maximum concurrent players.
  Affects memory usage and bandwidth.
  Plan: 1000 players = ~500MB RAM, 100 Mbps network
```

## Testing Configuration

### Unit Tests

```go
TestConfigLoadDefaults()    // No file, uses defaults
TestConfigLoadYAML()        // Loads from file
TestConfigEnvOverride()     // Environment overrides
TestConfigValidation()      // Invalid values rejected
TestConfigMerge()           // Hierarchy works correctly
```

### Integration Tests

```go
TestConfigServerStarts()    // Server boots with config
TestConfigPortBound()       // Correct port is bound
TestConfigLogging()         // Log level applied
```

## Best Practices

### Do's

✓ Use environment variables for secrets
✓ Validate configuration at startup
✓ Log final effective configuration (redacted)
✓ Provide clear error messages
✓ Document all configuration options

### Don'ts

✗ Hardcode server-specific settings
✗ Store secrets in YAML files
✗ Accept dynamic config changes (restart instead)
✗ Use different formats for different environments
✗ Make developers guess what environment variables exist

## Summary

The configuration system provides:

✓ **Flexibility** — Local dev, Docker, Cloud
✓ **Security** — Secrets external, no hardcoding
✓ **Clarity** — Explicit hierarchy, documented
✓ **Validation** — Prevents invalid startup
✓ **Portability** — Same code, different configs
