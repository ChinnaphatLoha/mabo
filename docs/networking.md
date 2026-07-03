# Networking Architecture

## Overview

The networking layer is responsible for transport protocol handling, connection lifecycle, and packet I/O. It is intentionally decoupled from game logic to allow protocol changes and testing in isolation.

## Network Stack Layers

```
┌────────────────────────────────────────┐
│  Application Layer                     │
│  (Rooms, Players, Simulation)          │
└────────────────┬─────────────────────┘
                 │
        ┌────────▼────────┐
        │ Transport Layer │
        │ (Server iface)  │
        └────────┬────────┘
                 │
        ┌────────▼──────────┐
        │ Network Layer     │
        │ (UDP listener)    │
        └────────┬──────────┘
                 │
        ┌────────▼──────────┐
        │ Physical Layer    │
        │ (OS socket)       │
        └────────┬──────────┘
                 │
        ┌────────▼──────────┐
        │ Internet          │
        │ (Client)          │
        └────────┬──────────┘
```

## UDP Transport

### Characteristics

- **Protocol**: UDP (connectionless, unreliable)
- **Port**: 9000 (configurable)
- **Buffer Size**: 4096 bytes (configurable)
- **Timeout**: 1 second read timeout

### Why UDP

- **Low latency**: No connection handshake or acknowledgements
- **Bandwidth efficient**: Minimal overhead per packet
- **Suitable for real-time**: Packet loss acceptable if state snapshots are frequent
- **Game standard**: Proven choice for MMOs and competitive games

### Alternative Protocols (Future)

The architecture allows swapping UDP for:
- **ENet** — Reliable UDP abstraction
- **WebSocket** — Browser clients
- **TCP** — Fallback for firewalled networks
- **Custom protocols** — Proprietary optimizations

Implementation only requires implementing the `Server` interface.

## Packet Structure

### Binary Format

```
┌─────────┬──────────────────────┐
│ Byte 0  │ Bytes 1+             │
├─────────┼──────────────────────┤
│ Packet  │ Variable-length      │
│ ID      │ Payload              │
│ (u16)   │ (Protocol-dependent) │
└─────────┴──────────────────────┘
```

### Packet Processing Pipeline

```
1. Socket reads UDP datagram (4096 bytes max)
   │
   ▼
2. Extract packet ID from first byte
   │
   ▼
3. Create Packet struct with ID and remaining bytes
   │
   ▼
4. Call application PacketHandler.HandlePacket()
   │
   ▼
5. Application layer processes based on packet ID
   │
   ▼
6. Application generates response (if needed)
   │
   ▼
7. Response sent back to client address
```

## Connection Lifecycle

### Server Perspective

```
┌──────────────┐
│   Listening  │  Server waits for packets
└──────┬───────┘
       │ Client sends Connect packet
       ▼
┌──────────────┐
│  Registered  │  Client address is known
└──────┬───────┘
       │ Receive input packets
       ▼
┌──────────────┐
│   Active     │  Processing player input
└──────┬───────┘
       │ Client goes silent (timeout) or sends Disconnect
       ▼
┌──────────────┐
│  Inactive    │  Awaiting reconnection
└──────┬───────┘
       │ Reconnect window expires or explicit disconnect
       ▼
┌──────────────┐
│  Removed     │  Session cleaned up
└──────────────┘
```

### Client Perspective

```
┌──────────────┐
│  Disconnected│  Not connected
└──────┬───────┘
       │ User clicks "Connect"
       ▼
┌──────────────┐
│   Connecting │  Waiting for server acknowledgement
└──────┬───────┘
       │ Receive acknowledgement (next state snapshot)
       ▼
┌──────────────┐
│  Connected   │  Receiving state snapshots
└──────┬───────┘
       │ Send input every frame (if changed)
       ▼
┌──────────────┐
│   Sending    │  Transmitting player input
└──────┬───────┘
       │ Network outage or user disconnect
       ▼
┌──────────────┐
│  Reconnecting│  Attempting to rejoin
└──────┬───────┘
       │ Server acknowledges
       ▼
┌──────────────┐
│  Connected   │  Back in sync
└──────────────┘
```

## Data Flow Examples

### Player Joins Game

```
Client: Send Connect packet → Server
        ▼
Server: Receive packet
Server: Validate client
Server: Register session
Server: Send welcome packet with current game state
        ▼
Client: Receive state snapshot
Client: Load arena, position player
Client: Render world
```

### Player Inputs Action

```
Client: Local player presses "move right"
Client: Send input packet (move command) → Server
        ▼
Server: Receive input at tick boundary
Server: Validate against current state
Server: Apply movement to world
Server: Update player position
Server: Include new state in next snapshot
        ▼
Client: Receive snapshot with new player position
Client: Update local world
Client: Render updated position
```

### Server Broadcasts State

```
Server: Tick boundary reached (every 50ms)
Server: Snapshot current world state
Server: Create state packet with:
        - Player positions
        - Projectiles
        - Entity animations
        - Game events
Server: Send to all connected clients
        ▼
Client: Receive snapshot
Client: Merge received state with local prediction
Client: Reconcile client-side prediction
Client: Render updated world
```

### Player Disconnects

```
Client: User clicks "Quit" or network timeout
Client: Send Disconnect packet → Server
        ▼
Server: Receive Disconnect packet
Server: Mark session as inactive
Server: Wait for reconnect (5 seconds, configurable)
Server: If reconnect arrives, resume session
Server: If timeout expires, remove session
Server: Broadcast updated player list
        ▼
Other Clients: Receive updated state without disconnected player
```

## Latency and Bandwidth Considerations

### Latency Budget

For a 20 TPS server (50ms ticks):

```
Network roundtrip:  20-100ms
Server processing:  5-10ms
Client prediction:  0-50ms (client predicts next frame)
Client rendering:   16ms (60 FPS assumed)
─────────────────
Total perceived:    ~41-126ms
```

### Bandwidth Optimization

**Outgoing** (server → clients):
- State snapshots: ~500-1000 bytes per tick per player
- For 10 players: ~500KB/s raw
- Optimized with delta compression: ~100KB/s
- With UDP header overhead: ~120KB/s

**Incoming** (client → server):
- Input commands: ~10-20 bytes per tick per player
- For 10 players: ~10KB/s raw
- With UDP header: ~12KB/s

## Security Considerations

### Current Foundation

- All packets validated against server state (future)
- No client-side authority over game state
- Connection from client IP (simple rate limiting ready)

### Future Enhancements

- DTLS encryption (if using standard DTLS libraries)
- Nonce-based replay attack prevention
- Input validation against game rules
- Rate limiting per client
- Kick for suspicious behavior

## Protocol Versioning

Not yet implemented, but the architecture supports:

```
┌────────────┬─────────────┐
│ Version    │ Byte 1      │
├────────────┼─────────────┤
│ Payload    │ Bytes 2+    │
└────────────┴─────────────┘
```

This allows server and client to agree on protocol before processing.

## Packet Handler Interface

```go
type PacketHandler interface {
    HandlePacket(ctx context.Context, source net.Addr, pkt packet.Packet)
}
```

The application implements this interface to receive packets. The network layer calls this method for every valid packet received.

## Testing Strategy

**Unit Tests**:
- Packet parsing with various inputs
- Handler dispatch logic
- Configuration loading

**Integration Tests**:
- UDP listener with mock client packets
- Full packet pipeline
- Connection lifecycle

**Load Tests** (future):
- 1000 concurrent connections
- Maximum packets per second
- Bandwidth utilization

## Summary

The networking layer provides:
- ✓ Reliable UDP transport
- ✓ Connectionless packet I/O
- ✓ Pluggable handler interface
- ✓ Extensible for future protocols
- ✓ Suitable for real-time multiplayer
