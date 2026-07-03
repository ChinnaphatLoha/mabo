# Packet Protocol Design

## Overview

The packet protocol defines the contract between client and server. Every message is identified by a packet ID, followed by a variable-length payload.

## Packet Structure

### Wire Format

```
┌─────────┬──────────────────────────┐
│ Byte 0  │ Bytes 1+                 │
├─────────┼──────────────────────────┤
│ Packet  │ Payload (variable)       │
│ ID (u8) │ Content depends on ID    │
└─────────┴──────────────────────────┘

Max packet size: 4096 bytes (including ID byte)
```

### Packet ID Space

The packet ID space is partitioned to group related functionality:

```
┌──────────────────┬─────────────────────┐
│ ID Range         │ Category            │
├──────────────────┼─────────────────────┤
│ 1-49             │ Session             │
│ 50-99            │ Reserved            │
│ 100-149          │ Ping/Health         │
│ 150-199          │ Input               │
│ 200-249          │ State               │
│ 250-299          │ Events              │
│ 300+             │ Future Extension    │
└──────────────────┴─────────────────────┘
```

## Session Packets (1-49)

### Packet 1: Connect

**Direction**: Client → Server

**Purpose**: Initiate connection

**Payload**:
```
[empty for MVP]

Future extensions:
- Client version
- Player name
- Authentication token
- Device info
```

**Server Response**: Sends state snapshot (Packet 200) acknowledging connection

**Status**: ✓ Defined

### Packet 2: Disconnect

**Direction**: Either direction

**Purpose**: Gracefully end session

**Payload**:
```
[empty for MVP]

Future extensions:
- Reason (logout, timeout, error)
- Custom message
```

**Status**: ✓ Defined

---

## Reserved Packets (50-99)

Reserved for future use:
- 50-74: Session extensions
- 75-99: Reserved for expansion

---

## Ping/Health Packets (100-149)

### Packet 100: Ping

**Direction**: Client → Server (or Server → Client for keep-alive)

**Purpose**: Latency measurement and connection keep-alive

**Payload**:
```
┌──────────────────┐
│ Ping ID (u32)    │  Echo this value back
│ Timestamp (i64)  │  Client-side nanoseconds (for latency calc)
└──────────────────┘
```

**Status**: ✓ Defined

### Packet 101: Pong

**Direction**: Server → Client (or Client → Server)

**Purpose**: Response to Ping, round-trip latency measurement

**Payload**:
```
┌──────────────────┐
│ Ping ID (u32)    │  Echoed from Ping packet
│ Timestamp (i64)  │  Echoed from Ping packet
└──────────────────┘
```

**Status**: ✓ Defined

---

## Input Packets (150-199)

Reserved for player input and commands:

### Packet 150: PlayerInput (future)

**Direction**: Client → Server

**Purpose**: Player action input (move, attack, ability)

**Planned Payload**:
```
┌──────────────────┐
│ Tick Number (u32)│  Tick input is for
│ Input Flags (u32)│  Bit flags for actions
│ Data (variable)  │  Direction, target, etc.
└──────────────────┘
```

**Status**: ⏳ Planned

### Packet 151: CommandAck (future)

**Direction**: Server → Client

**Purpose**: Acknowledge input reception

**Planned Payload**:
```
┌──────────────────┐
│ Tick Number (u32)│  Which tick was applied
│ Sequence (u32)   │  Command sequence number
└──────────────────┘
```

**Status**: ⏳ Planned

---

## State Packets (200-249)

### Packet 200: StateSnapshot

**Direction**: Server → Client

**Purpose**: Broadcast authoritative game state

**Planned Structure**:
```
┌──────────────────────┐
│ Tick Number (u32)    │
│ Timestamp (i64)      │
│ Player Count (u16)   │
│ Players (variable)   │
│  └─ ID, Position, Health, State
│ Entity Count (u16)   │
│ Entities (variable)  │
│  └─ ID, Position, Type, State
│ Event Count (u16)    │
│ Events (variable)    │
│  └─ Type, Data
└──────────────────────┘
```

**Frequency**: Every tick (20/sec)

**Status**: ⏳ Planned

### Packet 201: DeltaSnapshot (future optimization)

**Direction**: Server → Client

**Purpose**: Send only changed state (bandwidth optimization)

**Planned Payload**:
```
┌──────────────────────┐
│ Tick Number (u32)    │
│ Changed Players[]    │  Only those with changes
│ Changed Entities[]   │  Only those with changes
│ Removed Players[]    │  Disconnected
│ Removed Entities[]   │  Destroyed
│ New Events[]         │  Only new events
└──────────────────────┘
```

**Status**: ⏳ Planned for optimization phase

---

## Event Packets (250-299)

### Packet 250: GameEvent (future)

**Direction**: Server → Client or Client → Server

**Purpose**: Significant game occurrences (kill, level up, pickup)

**Planned Payload**:
```
┌──────────────────────┐
│ Event Type (u8)      │  Kill, Level, Pickup, etc.
│ Timestamp (i64)      │
│ Actor ID (u32)       │  Who triggered event
│ Target ID (u32)      │  Who was affected
│ Data (variable)      │  Event-specific
└──────────────────────┘
```

**Status**: ⏳ Planned

---

## Future Extension Packets (300+)

Reserved for:
- **300-349**: Chat and messaging
- **350-399**: Authentication and login
- **400-449**: Replay system
- **450-499**: Ranking and statistics
- **500+**: Custom extensions and modding

---

## Serialization Strategy

### MVP (Current)

**Packet format**: Raw binary with explicit offsets

```go
// Reading
id := buffer[0]
payload := buffer[1:len(buffer)]

// Writing
buffer[0] = byte(id)
copy(buffer[1:], payload)
```

**Status**: ✓ Implemented (basic)

### Phase 2 (Planned)

**Binary serialization**: Efficient encoding

```go
type ConnectPacket struct {
    ClientVersion uint32
    PlayerName    string  // Varint length + bytes
    Token         []byte  // Varint length + bytes
}
```

### Phase 3 (Future Optimization)

**Compression**: Reduce bandwidth

```
Raw snapshot: 1000 bytes
Compressed:   300 bytes (70% reduction)
```

Use zstd or similar compression library.

---

## Protocol Evolution

### Versioning (future)

Support multiple protocol versions for backward compatibility:

```
┌─────────────┬──────────────┐
│ Version (1) │ Packet (1)   │
├─────────────┼──────────────┤
│ Byte 0      │ Bytes 1+     │
└─────────────┴──────────────┘

Version 1: Current protocol
Version 2: Future with new features
```

### Deprecation

Old packets remain supported until deprecated:

```
Tick 0: Introduce new packet format
Tick 100,000: Mark old format deprecated
Tick 1,000,000: Remove support
```

---

## Error Handling

### Malformed Packets

If packet cannot be parsed:

```
1. Log error with client IP and packet ID
2. Discard packet
3. Continue (no crash)
4. If repeated: Consider rate limit or kick
```

### Oversized Packets

If packet exceeds 4096 bytes:

```
1. Discard
2. Log error
3. Consider closing connection if repeated
```

### Unknown Packet IDs

If ID is not recognized:

```
1. If in reserved range: Ignore (future compatibility)
2. If in extension range: Log warning
3. Continue (forward compatible)
```

---

## Bandwidth Analysis

### Typical Message Sizes

| Packet | Size | Frequency | Rate |
|--------|------|-----------|------|
| Input | 20 bytes | Every tick | 400 B/s |
| Ping/Pong | 16 bytes | Every 1s | 16 B/s |
| State | 500-1000 bytes | Every tick | 10-20 KB/s |
| Events | 20-50 bytes | Variable | ~100 B/s |
| **Total per player** | — | — | **~11 KB/s** |

For 10 concurrent players: ~110 KB/s

### Optimization Targets

- Delta snapshots: -70% state traffic
- Compression: -60-80% on snapshots
- Input prediction: -90% input traffic
- **Optimized total**: ~1-2 KB/s per player

---

## Testing

### Unit Tests

```go
TestPacketIDValidity() // ID in valid range
TestPacketSerialization() // Encode/decode round-trip
TestMalformedPacket() // Handles corrupt data
TestOversizedPacket() // Rejects > 4096 bytes
```

### Integration Tests

```go
TestClientServerHandshake() // Connect/Disconnect flow
TestPingPongLatency() // Measures round-trip correctly
TestStateSnapshot() // Full snapshot validated
```

### Load Tests

```
Send 1000 packets/sec
Measure parse time per packet
Verify no memory leaks
```

---

## Summary

The protocol provides:

✓ **Simple and efficient** — Binary format, minimal overhead
✓ **Extensible** — Reserved ranges for future growth
✓ **Explicit** — Every packet ID documented
✓ **Validated** — Malformed packets handled safely
✓ **Scalable** — Foundation for optimization
