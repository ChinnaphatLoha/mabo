# Logging Strategy

## Overview

Structured logging is essential for debugging, monitoring, and understanding server behavior in production. The logging system uses JSON output with configurable levels and contextual fields.

## Logging Architecture

### Log Flow

```
Application Code
    │
    ├─ logger.Info("message", "field", value)
    │
    ▼
Logger Instance
    │
    ├─ Filter by level (debug/info/warn/error)
    ├─ Add timestamp
    ├─ Format as JSON
    │
    ▼
JSON Handler
    │
    ├─ Marshal to JSON
    │
    ▼
Output Stream (stdout)
    │
    ├─ To container logs
    ├─ To log aggregator (ELK, CloudWatch)
    ├─ To file (if redirected)
    │
    ▼
Operational Visibility
```

## Log Levels

### Available Levels

```
DEBUG   - Detailed diagnostic information
INFO    - General informational messages
WARN    - Warning conditions that should be investigated
ERROR   - Error conditions that failed to complete normally
```

### Level Hierarchy

```
More Verbose                Less Verbose
DEBUG > INFO > WARN > ERROR > (none)

Configuration: LOG_LEVEL=debug
Outputs: DEBUG, INFO, WARN, ERROR

Configuration: LOG_LEVEL=warn
Outputs: WARN, ERROR
Filters: DEBUG, INFO hidden
```

### When to Use Each Level

**DEBUG**
```
logger.Debug("packet received", "source", "192.168.1.1:9000", "size", 256)
logger.Debug("tick advanced", "tick", 42, "elapsed_ms", 48)
logger.Debug("database query", "table", "players", "duration_ms", 2)
```
Use for: Detailed tracing, packet I/O, loop iterations

**INFO**
```
logger.Info("server started", "port", 9000, "tick_rate", 20)
logger.Info("player connected", "player_id", "abc123", "room_id", "room_001")
logger.Info("graceful shutdown", "connected_players", 42)
```
Use for: Major lifecycle events, important state changes

**WARN**
```
logger.Warn("high latency detected", "player_id", "abc123", "latency_ms", 500)
logger.Warn("packet drop", "lost_count", 5, "total_sent", 1000)
logger.Warn("database slow", "query", "get_player", "duration_ms", 200)
```
Use for: Unusual conditions that should be monitored

**ERROR**
```
logger.Error("player disconnected unexpectedly", "player_id", "abc123", "reason", "timeout")
logger.Error("database connection failed", "error", "connection refused")
logger.Error("invalid packet", "source", "192.168.1.1:9000", "reason", "malformed")
```
Use for: Failures, exceptions, unrecoverable errors

## Structured Logging

### JSON Output Format

Every log entry is valid JSON:

```json
{
  "time": "2026-07-03T10:30:45.123Z",
  "level": "INFO",
  "msg": "player connected",
  "player_id": "abc123",
  "room_id": "room_001",
  "remote": "192.168.1.100:54321"
}
```

### Structured Fields

Fields are key-value pairs, not string formatting:

```go
// BAD (unstructured)
logger.Info(fmt.Sprintf("Player %s joined room %s", playerID, roomID))

// GOOD (structured)
logger.Info("player joined", "player_id", playerID, "room_id", roomID)
```

Benefits:
- Queryable in log aggregators: `player_id:abc123`
- Type-safe (values aren't strings)
- Machine-readable
- Better performance

### Common Fields

```
Timing
├─ time        - Timestamp (ISO 8601)
├─ duration_ms - Operation duration

Identity
├─ player_id   - Player identifier
├─ room_id     - Room identifier
├─ session_id  - Session identifier
├─ request_id  - Correlation ID across systems

Networking
├─ remote      - Client IP and port
├─ packet_id   - Packet type
├─ size        - Payload size
├─ latency_ms  - Round-trip latency

System
├─ tick        - Current tick number
├─ connected   - Connected player count
├─ error       - Error message/code
```

## Contextual Logging

### Creating Context-Aware Loggers

```go
// Main logger
mainLogger := logger.New("debug")

// Player-specific logger
playerLogger := mainLogger.With(
    "player_id", "abc123",
    "room_id", "room_001",
)

// All logs from playerLogger include player_id and room_id
playerLogger.Info("player input received", "action", "move_right")
// Output:
// {
//   "level": "INFO",
//   "msg": "player input received",
//   "player_id": "abc123",
//   "room_id": "room_001",
//   "action": "move_right"
// }
```

### Request ID Tracing

```go
// Add request ID for correlating logs across services
requestID := generateUUID()

reqLogger := logger.With(
    "request_id", requestID,
    "client_ip", clientIP,
)

// All logs for this request include request_id
reqLogger.Info("processing request")
reqLogger.Debug("database query", "table", "players")
reqLogger.Debug("cache hit", "key", "player:abc123")
reqLogger.Info("request completed", "status", "success")

// In log aggregator, search by request_id to see entire flow
```

## Environment-Specific Configuration

### Development

```
LOG_LEVEL=debug

Output: DEBUG, INFO, WARN, ERROR
Includes: Detailed diagnostic info, package names
Use: Local debugging
```

### Testing

```
LOG_LEVEL=warn

Output: WARN, ERROR (only problems)
Suppresses: Normal operation noise
Use: Focus on failures
```

### Staging

```
LOG_LEVEL=info

Output: INFO, WARN, ERROR (no debug noise)
Includes: Lifecycle events, important changes
Use: Monitor test environment
```

### Production

```
LOG_LEVEL=warn

Output: WARN, ERROR (only actionable issues)
Suppresses: Normal operation details
Sent to: Log aggregator (CloudWatch, Datadog, etc.)
Use: Focus on issues
```

## Logging Best Practices

### Do's

✓ Use structured fields, not formatted strings
✓ Include contextual information (player_id, room_id)
✓ Use appropriate log levels (ERROR for failures, INFO for milestones)
✓ Add error messages to ERROR logs
✓ Include timestamps and correlation IDs
✓ Keep messages concise and descriptive

### Don'ts

✗ Log passwords, tokens, or sensitive data
✗ Log sensitive customer information
✗ Use DEBUG level for important information
✗ Create unstructured strings with sprintf
✗ Log entire objects/structs (extract fields)
✗ Add redundant information (timestamp is automatic)

### Examples

```go
// BAD
logger.Info(fmt.Sprintf("Player abc123 in room room_001 performed action: move_right at tick 42"))

// GOOD
logger.Info("player input processed",
    "player_id", playerID,
    "room_id", roomID,
    "action", action,
    "tick", tick,
)

// BAD
logger.Error("Database error: " + err.Error())

// GOOD
logger.Error("database query failed",
    "query", "get_player",
    "error", err.Error(),
    "player_id", playerID,
)

// BAD
logger.Debug("Server running")

// GOOD
logger.Info("server started",
    "bind_address", cfg.BindAddress,
    "tick_rate", cfg.TickRate,
    "max_players", cfg.MaxPlayers,
)
```

## Performance Considerations

### Efficiency

- **Structured logging is fast**: JSON encoding is optimized
- **No allocations for skipped levels**: Suppressed logs don't allocate memory
- **Async writing**: Consider async handler for high-volume logs
- **Sampling**: Log every Nth high-frequency event (future optimization)

### Overhead

```
Approximate per-log cost:
- Structured field: ~1-2 microseconds
- JSON encoding: ~5-10 microseconds
- Network transmission: ~100+ microseconds (if remote)

20 TPS server, 100 logs/tick:
2000 logs/sec × 10 µs = 20 ms overhead (acceptable)
```

## Log Aggregation (Future)

### Integration Points

```
Application logs (stdout)
    │
    ├─ Docker logs
    │   ├─ kubectl logs (Kubernetes)
    │   ├─ docker logs (local)
    │
    ├─ Log aggregator
    │   ├─ Elasticsearch (ELK Stack)
    │   ├─ CloudWatch (AWS)
    │   ├─ DataDog
    │   ├─ Splunk
    │
    ├─ Analysis
    │   ├─ Real-time dashboards
    │   ├─ Alerting
    │   ├─ Historical analysis
```

### ELK Stack Example (future)

```
Server outputs JSON to stdout
    │
    ├─ Logstash collects logs
    ├─ Parses JSON
    ├─ Indexes in Elasticsearch
    │
    ├─ Kibana visualizes
    │   ├─ Real-time dashboard
    │   ├─ Search: player_id:abc123
    │   ├─ Graph: latency over time
```

## Log Retention

### Strategy

```
Development: Keep all logs (local filesystem)
Testing: 1 week retention
Staging: 2 week retention
Production: 30 day retention (compliance)
```

### Rotation (if file-based)

```
Rotate daily or at 1GB (whichever first)
Keep 7 compressed archives
Delete older files
```

## Sensitive Information

### What NOT to Log

```
✗ Passwords
✗ API keys or tokens
✗ Personal data (SSN, email)
✗ Credit card numbers
✗ OAuth bearer tokens
✗ Session cookies
```

### Redaction Strategy

```go
// If you must log auth data, redact it
token := "secret_abc123def456"
redactedToken := token[:6] + "***" // "secret_***"
logger.Debug("authentication", "token", redactedToken)
```

## Testing Logs

### Unit Tests

```go
TestLogStructure()   // Verify JSON format
TestLogLevels()      // Correct level filtering
TestContextualFields() // Fields preserved
TestNoSensitiveData()  // Passwords not logged
```

### Integration Tests

```go
TestLogOutput()      // Logs appear on stdout
TestLogAggregation() // Logs collected by aggregator
```

## Summary

The logging system provides:

✓ **Structured JSON output** — Machine-readable, queryable
✓ **Configurable levels** — DEV debug, PROD warnings only
✓ **Contextual fields** — Trace requests and player sessions
✓ **Performance** — Low overhead, no allocations for skipped logs
✓ **Production-ready** — Integrates with log aggregators
✓ **Security** — No sensitive data leaks
