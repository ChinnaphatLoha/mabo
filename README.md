# Mabo

A multiplayer game backend foundation built with Go and Godot. This project demonstrates server-authoritative architecture, networking abstraction, and clean code principles suitable for a senior-level portfolio.

## Project Overview

Mabo is a MOBA-style game server designed to support 5v5 multiplayer gameplay with future scalability to 1000 concurrent players. The project prioritizes **networking quality** and **clean architecture** over graphics fidelity.

### Core Goals

- Server-authoritative architecture
- Tick-based simulation (20 TPS)
- UDP networking with abstraction layers
- Dedicated server model
- Reconnection support
- Extensible foundation for matchmaking, replay, ranking, chat, AI, spectator, and analytics

## Project Structure

```
mabo/
├── client/                  # Godot client assets
│   └── godot/              
├── server/                  # Go backend service
│   ├── cmd/
│   │   └── server/         # Application entrypoint
│   ├── internal/            # Private packages
│   │   ├── app/            # Application lifecycle
│   │   ├── config/         # Configuration loading
│   │   ├── logger/         # Structured logging
│   │   ├── network/        # UDP transport layer
│   │   ├── packet/         # Packet definitions
│   │   ├── transport/      # Transport abstractions
│   │   ├── player/         # Player state
│   │   ├── room/           # Room management
│   │   ├── world/          # World simulation
│   │   ├── session/        # Session lifecycle
│   │   ├── command/        # Input handling
│   │   └── system/         # Gameplay systems
│   ├── configs/            # Runtime configuration
│   ├── Dockerfile          # Container image
│   ├── Makefile            # Build targets
│   ├── go.mod              # Go module definition
│   └── go.sum              # Dependency lock
├── docker/                  # Container orchestration
│   └── docker-compose.yml  # Local development setup
├── docs/                    # Architecture and design
│   ├── architecture.md     # System architecture
│   ├── networking.md       # Networking design
│   ├── tick-system.md      # Tick loop design
│   ├── protocol.md         # Packet protocol
│   ├── config.md           # Configuration strategy
│   └── logging.md          # Logging strategy
├── scripts/                 # Helper scripts
└── shared/                  # Shared models and contracts
```

## Getting Started

### Prerequisites
- Go 1.26.4+
- Docker & Docker Compose (optional)

### Local Development

1. **Start the server**
   ```bash
   cd server
   make run
   ```

2. **Build the binary**
   ```bash
   cd server
   make build
   ```

3. **Run tests**
   ```bash
   cd server
   make test
   ```

4. **Format code**
   ```bash
   cd server
   make fmt
   ```

### Docker Deployment

```bash
cd server
make docker
```

This starts the server container on port 9000.

## Architecture Highlights

### Server-Authoritative Design
- All game state authority resides on the server
- Clients submit input; server processes and broadcasts authoritative state
- Prevents cheating and ensures consistency

### Layered Architecture
1. **Transport Layer** — UDP networking, packet I/O
2. **Protocol Layer** — Packet definitions and serialization
3. **Application Layer** — Business logic, rooms, players, simulation
4. **Domain Layer** — Core gameplay systems

### Networking Abstraction
- Packet handling is decoupled from game logic
- Transport protocol can be swapped (UDP → ENet, WebSocket)
- Clear interfaces between layers

### Configuration-Driven
- YAML configuration file: `server/configs/server.yml`
- Environment variable overrides: `SERVER_PORT`, `LOG_LEVEL`, etc.
- Production-ready deployment strategy

### Structured Logging
- JSON-based structured logs
- Configurable log levels (debug, info, warn, error)
- Context tracking with player IDs, request IDs, and correlation fields

## Configuration

Server configuration is loaded from `server/configs/server.yml`:

```yaml
port: 9000
bind_address: "0.0.0.0:9000"
tick_rate: 20
max_players: 1000
max_rooms: 100
log_level: debug
```

Override with environment variables:
- `SERVER_PORT` — server port
- `SERVER_BIND_ADDRESS` — bind address
- `SERVER_TICK_RATE` — simulation tick rate (Hz)
- `SERVER_MAX_PLAYERS` — maximum concurrent players
- `SERVER_MAX_ROOMS` — maximum concurrent rooms
- `LOG_LEVEL` — logging level

## Design Principles

1. **Game logic never depends on networking** — Gameplay is independent of transport
2. **Networking is replaceable** — Abstract transport allows protocol swaps
3. **Configuration is external** — No hardcoded settings
4. **Everything is testable** — Dependency injection and interfaces throughout
5. **No global state** — Dependency injection preferred
6. **Composition over inheritance** — Go's interface-based design
7. **Small packages** — Single responsibility principle
8. **Readable code** — Code reads like English
9. **Portfolio quality** — Production-grade architecture and documentation

## Networking Model

### Packet Flow

```
Client UDP → Network Layer → Packet Layer → Handler → Application Layer → World State
                                                           ↓
Response broadcasts state snapshots back to connected clients
```

### Server Authority

1. Client sends input packet to server
2. Server validates input in the current game context
3. Server simulates state change deterministically
4. Server broadcasts authoritative state to all players
5. Client receives state and updates local world

### Tick-Based Simulation

- Fixed 20 ticks per second (50ms per tick)
- Deterministic input processing
- Server broadcasts snapshot each tick
- Client-side prediction and reconciliation ready for implementation

## Packet Protocol

### Current Packet IDs

| ID | Name | Direction | Purpose |
|---|---|---|---|
| 1 | Connect | Client → Server | Initiate connection |
| 2 | Disconnect | Either direction | Terminate session |
| 100 | Ping | Client → Server | Latency measurement |
| 101 | Pong | Server → Client | Ping response |

Future packets will include input commands, state snapshots, and gameplay events.

## Deployment

### Local Development
```bash
make docker
```

### Cloud Deployment

The foundation supports future Kubernetes deployment:
- Stateless servers (sessions stored externally)
- Horizontal scaling through load balancing
- Per-server room/match isolation
- Session persistence via databases

## Future Extensions

The architecture is designed to accommodate without major refactoring:
- **Matchmaking** — Room selection and player pairing
- **Authentication** — Login and session management
- **Replay System** — Record and playback of matches
- **Ranking** — Player rating and leaderboards
- **Chat** — In-game messaging
- **AI Bots** — Non-player characters
- **Spectator Mode** — Observe matches
- **Analytics** — Performance and gameplay metrics

## Current Status

**Foundation Complete**
- Server bootstrap and lifecycle ✓
- Configuration management ✓
- Structured logging ✓
- UDP networking ✓
- Packet abstraction ✓
- Transport abstraction ✓
- Docker deployment ✓

**Not Yet Implemented**
- Room and match systems
- Player state management
- Authoritative tick loop
- Simulation and physics
- Packet serialization/deserialization
- Gameplay systems
- Client implementation

## Testing

Run the test suite:
```bash
cd server
go test ./...
```

## Code Quality

- **Format**: `make fmt`
- **Lint**: Use `golangci-lint` (optional)
- **Test**: `make test`

## Contributing

This is a portfolio project. Contributions should maintain:
- Clean architecture principles
- Clear package separation
- Production-grade code quality
- Comprehensive design documentation

## License

See LICENSE file.

## References

- [Architecture Documentation](docs/architecture.md)
- [Networking Design](docs/networking.md)
- [Tick System Design](docs/tick-system.md)
- [Packet Protocol](docs/protocol.md)
- [Configuration Strategy](docs/config.md)
- [Logging Strategy](docs/logging.md)
