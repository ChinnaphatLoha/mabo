# Tick System Design

## Overview

A tick-based simulation loop is the foundation for deterministic, authoritative multiplayer gameplay. This document describes the design of the 20 TPS (ticks-per-second) tick system that will drive the server's game simulation.

## Tick Rate: 20 Hz

### Decision Rationale

```
Tick rate: 20 TPS
Tick duration: 50 ms
Ticks per second: 20

Benefits:
- Input responsiveness: ~50ms from input to effect
- Bandwidth efficiency: 20 snapshots/sec vs 60+
- CPU efficiency: Simulation runs 20x/sec, not 60+
- Network standard: Common for competitive games (CS, Dota2)
- Headroom: Leaves room for lag compensation and prediction
```

### Alternatives Considered

| Rate | Advantage | Disadvantage |
|------|-----------|--------------|
| 10 TPS | Very efficient | Input lag 100ms, feels sluggish |
| 20 TPS | **Balanced** | **Standard choice** |
| 30 TPS | Slightly smoother | Less bandwidth savings |
| 60 TPS | Matches rendering | Doubles bandwidth, CPU load |

## Tick Loop Structure

### Phases per Tick

Each tick executes these phases in order:

```
┌─────────────────────────────────────────────────────────────┐
│                     START OF TICK                           │
├─────────────────────────────────────────────────────────────┤
│                                                              │
│  Phase 1: Input Collection (0-5ms)                         │
│  ├─ Receive packets from network layer                     │
│  ├─ Parse input commands                                   │
│  ├─ Buffer commands for validation                         │
│  └─ Track client sequence numbers                          │
│                                                             │
│  Phase 2: Input Validation (5-10ms)                        │
│  ├─ Check inputs against game rules                        │
│  ├─ Verify player state allows action                      │
│  ├─ Discard invalid/duplicate inputs                       │
│  └─ Queue valid commands for simulation                    │
│                                                             │
│  Phase 3: Simulation (10-35ms)                             │
│  ├─ Apply movement                                         │
│  ├─ Update positions                                       │
│  ├─ Check collisions                                       │
│  ├─ Process events                                         │
│  ├─ Update entity state                                    │
│  └─ Advance timers                                         │
│                                                             │
│  Phase 4: Snapshot Creation (35-40ms)                      │
│  ├─ Capture world state                                    │
│  ├─ Create delta from last snapshot                        │
│  ├─ Include relevant entities for each player              │
│  └─ Prepare state packet                                   │
│                                                             │
│  Phase 5: Broadcast (40-45ms)                              │
│  ├─ Send state snapshot to all players                     │
│  ├─ Include tick number and timestamp                      │
│  ├─ Log any send failures                                  │
│  └─ Update connection state                                │
│                                                             │
│  Phase 6: Cleanup (45-50ms)                                │
│  ├─ Remove disconnected sessions                           │
│  ├─ Garbage collect old commands                           │
│  ├─ Update statistics                                      │
│  └─ Wait until next tick boundary                          │
│                                                             │
├─────────────────────────────────────────────────────────────┤
│                      END OF TICK                            │
├─────────────────────────────────────────────────────────────┤
│                                                              │
│  Phase 6.5: Sleep (if early)                               │
│  └─ Sleep until 50ms boundary                              │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

## Deterministic Simulation

### Principles

All inputs produce the same result given the same starting state:

```
Input: Position(10,20) + Move(Right) + Velocity(5)
│
Tick 1: Position becomes (15,20)
Tick 2: Position becomes (20,20)
...always the same result for same input
```

### Fixed Time Step

Each tick advances the world by exactly 50ms of game time:

```go
// Pseudocode
deltaTime := 50 * time.Millisecond  // Always the same

for x := 0; x < 10; x++ {
    // Update each entity
    entity.Position += entity.Velocity * deltaTime
    entity.Update(deltaTime)
}
```

No frame-rate dependency, no time.Now() in simulation logic.

### No Floating-Point Errors

Use fixed-point or integer math for position/physics:

```go
// BAD (floating point errors accumulate)
position += velocity * deltaTime

// GOOD (fixed point, no errors)
position += velocity * (deltaTime / fixedStep)
```

## Packet Processing

### Input Packet Lifecycle

```
1. Network layer receives input packet
   └─ Packet ID, client ID, input data
   
2. Application queues for next tick
   └─ Buffered until Phase 1
   
3. Tick Phase 1: Collect inputs
   └─ Parse buffered packets
   
4. Tick Phase 2: Validate
   └─ Check against current state
   
5. Tick Phase 3: Apply
   └─ Update world state
   
6. Tick Phase 4-5: Broadcast
   └─ Send result to client
```

### Tick Numbering

Each tick has a unique number:

```
Tick 0: t=0ms
Tick 1: t=50ms
Tick 2: t=100ms
...
Tick 1000: t=50000ms
```

Clients track tick numbers to:
- Discard old packets
- Predict next state
- Reconcile with server

## State Snapshots

### Structure

```
Snapshot {
    TickNumber: uint32
    Timestamp: int64
    Players: []Player {
        ID, Position, Velocity, Health, Animation
    }
    Entities: []Entity {
        ID, Position, Type, State
    }
    Events: []Event {
        Type, Data, Timestamp
    }
}
```

### Broadcast Schedule

**Every tick**:
- Send to all connected clients
- Include full or delta state (configurable)

**Optimization** (future):
- Only send visible entities (frustum culling)
- Use delta compression (only changed values)
- Prioritize by distance/importance

## Client-Side Prediction & Reconciliation

### Current Status

⏳ Not yet implemented

### Design (for future)

```
Client Tick Loop:
┌──────────────────────┐
│ Local Prediction     │
├──────────────────────┤
│ Apply local input    │
│ Advance simulation   │
│ Render predicted     │
└──────────────────────┘
           │
           ▼
┌──────────────────────┐
│ Receive Server       │
│ Snapshot            │
├──────────────────────┤
│ Compare against      │
│ local prediction     │
│ Reconcile state      │
│ Smooth corrections   │
└──────────────────────┘
```

### Reconciliation Strategy

If server state differs from prediction:
1. Compare deltas
2. If < 5% difference: smooth correction
3. If > 5% difference: snap to authoritative state
4. Replay future inputs on corrected state

This prevents jittering while maintaining responsiveness.

## Latency & Jitter

### Latency Handling

Client input arrives at server with latency:

```
Client      Network     Server
│           (50ms)      │
│─ Input ─────────────>│
│           Tick 0      │
│           Tick 1      │
│           Tick 2 (input applied here)
```

Solution: **Input buffering**
- Clients include a time stamp with input
- Server applies at the correct historical tick
- Or: Accept latency as-is and smooth corrections

### Jitter Handling

Network jitter causes packets to arrive out-of-order:

```
Packet 1 arrives: Tick 5 state
Packet 3 arrives: Tick 7 state
Packet 2 arrives: Tick 6 state (late)
```

Solution: **Sequence numbers**
- Drop out-of-order packets
- Smoothly interpolate between authoritative states
- Resync if too far behind

## Tick Synchronization

### Server-Side

```go
tickBoundary := 50 * time.Millisecond

for {
    tickStart := time.Now()
    
    // Run all phases
    processInputs()
    simulateWorld()
    createSnapshot()
    broadcastState()
    
    // Sleep until next boundary
    elapsed := time.Now().Sub(tickStart)
    if elapsed < tickBoundary {
        time.Sleep(tickBoundary - elapsed)
    }
}
```

### Clock Synchronization (future)

For distributed servers, synchronize clocks with NTP:
- Ensures consistent tick boundaries across servers
- Enables smooth player migration between servers
- Prevents desync when servers communicate

## Scaling Considerations

### Single Server

- One tick loop
- Simple state management
- All players in sync

### Multiple Servers

```
┌───────────────────┐      ┌───────────────────┐
│  Server 1         │      │  Server 2         │
│  Tick 0,1,2,...   │      │  Tick 0,1,2,...   │
│  Players 1-5      │      │  Players 6-10     │
└───────────────────┘      └───────────────────┘
        │                            │
        └────────────┬───────────────┘
                     │
            Replicate state
            between servers
            (if players interact)
```

### Tick Batching (future)

Group multiple ticks before sending:

```
Send every tick:     20 packets/sec (current)
Send every 2 ticks:  10 packets/sec
Send every 5 ticks:  4 packets/sec (if applicable)
```

Trades responsiveness for bandwidth.

## Testing

### Unit Tests

```go
TestTickAdvancements() // Tick counter increments correctly
TestDeterministicMovement() // Same inputs = same result
TestInputBuffering() // Inputs queued correctly
```

### Integration Tests

```go
TestFullTickCycle() // Input → Simulate → Snapshot → Broadcast
TestMultiplePlayersSync() // Players advance together
TestInputValidation() // Invalid inputs rejected
TestStateSnapshot() // Snapshot contains expected data
```

### Load Tests (future)

```
1000 players
100 ticks/sec for 1 hour
Measure:
- CPU usage
- Memory growth
- Snapshot size
- Broadcast latency
```

## Summary

The tick system provides:

✓ **Deterministic simulation** — Same inputs always produce same results
✓ **Fixed time step** — No frame-rate dependency
✓ **Authoritative server** — All state owned by server
✓ **Scalability** — Foundation for distributed servers
✓ **Client prediction ready** — Room for local prediction and reconciliation
✓ **Production-grade** — Proven approach for multiplayer games
