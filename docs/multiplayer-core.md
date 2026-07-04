# Multiplayer Core

## Overview

The multiplayer core implements a complete game-server loop. Multiple Godot clients can:

1. Log in as guests to obtain a session and player ID.
2. Create or join rooms (lobby membership, capacity-based).
3. Automatically start a match when a room reaches capacity.
4. Receive deterministic spawn events with team assignment and spawn position.
5. Submit movement inputs; the server updates positions authoritatively.
6. Receive 20 TPS world snapshots covering all players in the match.
7. Broadcast a `PlayerDisconnected` event when a peer leaves.

---

## Architecture

```
Client Input → Network (UDP)
                   ↓
             Protocol Decode (packet/protocol)
                   ↓
        Application Validation (app/multiplayer_handler)
                   ↓
       ┌───────────┼──────────────────┐
  session.Manager room.Manager  match.Manager
                   ↓
              world.World (SpawnPlayer / ApplyInputs / Snapshots)
                   ↓
           system.InterestManager
                   ↓
             Snapshot Broadcast (20 TPS tick loop)
                   ↓
             UDP Send → Clients
```

Domain packages (`session`, `room`, `match`, `game`, `world`, `command`) never import network or UDP types. Application orchestration in `internal/app` is the only package that touches both.

---

## Packet IDs (canonical)

| Range | Purpose | IDs |
|---|---|---|
| 1–2 | Transport | Connect(1), Disconnect(2) |
| 10–11 | Auth | LoginRequest(10), LoginResponse(11) |
| 20–25 | Room | CreateRoomRequest(20), RoomCreated(21), JoinRoomRequest(22), RoomJoined(23), LeaveRoomRequest(24), RoomLeft(25) |
| 49 | Error | ErrorResponse(49) |
| 100–101 | Ping | Ping(100), Pong(101) |
| 150 | Input | MovementInput(150) |
| 200 | State | Snapshot(200) |
| 250–251 | Events | PlayerSpawned(250), PlayerDisconnected(251) |

---

## Wire Format

One-byte packet ID prefix (from Day 1), followed by JSON payload.

### Client → Server payloads

```json
LoginRequest      { "guest_name"?: "string" }
CreateRoomRequest { "capacity"?: 2..10, default 10 }
JoinRoomRequest   { "room_id": "string" }
LeaveRoomRequest  { "room_id": "string" }
MovementInput     { "sequence": uint, "move_x": float, "move_y": float, "rotation": float }
```

### Server → Client payloads

```json
LoginResponse   { "session_id": "string", "player_id": "string" }
RoomResponse    { "room_id", "player_id", "players": [], "capacity", "state" }
PlayerSpawned   { "match_id", "room_id", "entity_id", "player_id", "team", "position": {"x","y"}, "rotation", "hp", "animation_state" }
PlayerDisconnected { "room_id", "match_id", "player_id" }
Snapshot        { "tick", "match_id", "players": [SnapshotPlayer] }
SnapshotPlayer  { "player_id", "entity_id", "team", "position", "rotation", "velocity", "animation_state", "hp" }
ErrorResponse   { "request_packet_id", "code", "message" }
```

### Error codes

| Code | Meaning |
|---|---|
| `INVALID_ROOM` | Room ID not found |
| `ROOM_FULL` | Room at capacity or not joinable |
| `DUPLICATE_JOIN` | Player already in room |
| `ALREADY_CONNECTED` | Login from same address while session active |
| `DISCONNECTED` | Packet received from unauthenticated address |
| `INVALID_PACKET` | Malformed JSON or unknown packet ID |

---

## Room State Machine

```
[*] → Created → Waiting → Playing → Finished → Destroyed → [*]
```

- Room starts in `Waiting` when created.
- Transitions to `Playing` when capacity is reached (match starts).
- Room stays separate from Match: room = lobby/membership, match = running simulation.

---

## Match State Machine

```
[*] → Created → Running → Finished → Destroyed → [*]
```

- Match created when room reaches capacity.
- Teams balanced by current team counts (0 or 1).
- Spawn positions are deterministic: team 0 → `(index*2, 0)`, team 1 → `(100+index*2, 0)`.

---

## Movement & Tick Loop

- `MovementInput` carries **move vector and rotation only** (no client position).
- Server normalizes the move vector and applies `velocity = normalized * PlayerMoveSpeed (5.0)`.
- The `command.InputBuffer` stores the **latest** input per player by sequence number; older sequences are discarded.
- Every 50 ms (20 TPS): drain inputs → `world.ApplyInputs` → `world.Snapshots` → broadcast via `system.InterestManager`.
- `BroadcastInterestManager` sends every snapshot to every player in the match. Future implementations can replace this with frustum/distance culling without touching room, match, or world code.

---

## New Files

| File | Purpose |
|---|---|
| `internal/app/multiplayer_handler.go` | `PacketHandler` implementation, wires all managers |
| `internal/app/tick_loop.go` | 20 TPS simulation and snapshot broadcast |
| `internal/app/multiplayer_test.go` | Integration tests (fake UDP sender) |
| `internal/system/visibility.go` | `InterestManager` interface + `BroadcastInterestManager` |
| `internal/command/buffer.go` | `InputBuffer` – latest-input-per-player, thread-safe |
| `internal/world/world.go` | Authoritative world: spawn, apply inputs, snapshot |
| `internal/protocol/messages.go` | All JSON DTOs |
| `internal/packet/packet.go` | Canonical packet IDs |

### Modified files

| File | Change |
|---|---|
| `internal/room/manager.go` | Added `List() []*Room` for disconnect cleanup scan |
| `internal/session/session.go` | Canonical `Session` type extracted |
| `internal/room/room.go` | Canonical `Room` type extracted |
| `internal/match/manager.go` | Match + team assignment + lifecycle |
| `cmd/server/main.go` | Full multiplayer wiring |

---

## Test Coverage

### Unit tests (all green)

| Package | Tests |
|---|---|
| `session` | Create, duplicate address rejection, remove by addr |
| `room` | Create, duplicate join, full room, playing not joinable, leave, destroy |
| `match` | Create, get by room, remove, team assignment balance |
| `game` | Spawn determinism by team and slot |
| `command` | Buffer add (sequence filtering), drain clears buffer |
| `world` | Spawn initial state, movement normalization, remove player |
| `system` | BroadcastInterestManager returns all players |
| `packet` | All packet IDs in correct ranges |
| `protocol` | DTO round-trip, error codes |

### Integration tests (`internal/app`)

| Test | Scenario |
|---|---|
| `TestLoginCreatesSession` | Login creates session, response has IDs |
| `TestDuplicateLoginReturnsAlreadyConnected` | Second login from same addr → `ALREADY_CONNECTED` |
| `TestCreateRoomAutoJoinsCreator` | Room created + creator auto-joined |
| `TestJoinRoomBroadcastsAndStartsMatchWhenFull` | Full room → match starts, both clients get `PlayerSpawned` |
| `TestDuplicateJoinReturnsError` | Player rejoins own room → `DUPLICATE_JOIN` |
| `TestJoinInvalidRoomReturnsError` | Unknown room ID → `INVALID_ROOM` |
| `TestLeaveRoomNotifiesOtherMembers` | Leave broadcasts `RoomLeft` to remaining members |
| `TestMovementInputBufferedAndAppliedByTickLoop` | Input buffered with correct fields |
| `TestDisconnectBroadcastsPlayerDisconnected` | Disconnect → peer receives `PlayerDisconnected`, session removed |
| `TestUnloggedClientGetsDisconnectedError` | Unauthenticated packet → `DISCONNECTED` |
| `TestMalformedPacketGetsInvalidPacketError` | Bad JSON → `INVALID_PACKET` |
| `TestSnapshotBroadcastViaTickLoop` | Manual tick → both clients receive `Snapshot` |

---

## Manual Verification Checklist

1. Start server: `cd server && go run ./cmd/server`
2. Client A sends `LoginRequest` → receives `LoginResponse` with session/player IDs.
3. Client A sends `CreateRoomRequest{capacity:2}` → receives `RoomCreated` + `RoomJoined`.
4. Client B logs in, sends `JoinRoomRequest{room_id}` → both clients receive `RoomJoined` + `PlayerSpawned`.
5. Client A sends `MovementInput` → next tick snapshot shows updated position for Client A.
6. Client A sends `Disconnect` → Client B receives `PlayerDisconnected`.
7. Error cases: invalid room, duplicate join, unauthenticated packet → correct `ErrorResponse.code`.
