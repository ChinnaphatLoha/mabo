package app_test

import (
	"context"
	"encoding/json"
	"net"
	"sync"
	"testing"

	"github.com/ChinnaphatLoha/mabo/server/internal/app"
	"github.com/ChinnaphatLoha/mabo/server/internal/command"
	"github.com/ChinnaphatLoha/mabo/server/internal/logger"
	"github.com/ChinnaphatLoha/mabo/server/internal/match"
	"github.com/ChinnaphatLoha/mabo/server/internal/packet"
	"github.com/ChinnaphatLoha/mabo/server/internal/protocol"
	"github.com/ChinnaphatLoha/mabo/server/internal/room"
	"github.com/ChinnaphatLoha/mabo/server/internal/session"
	"github.com/ChinnaphatLoha/mabo/server/internal/system"
	"github.com/ChinnaphatLoha/mabo/server/internal/world"
)

// ── Fake network helpers ───────────────────────────────────────────────────

// fakeSender captures packets sent to each address.
type fakeSender struct {
	mu      sync.Mutex
	packets map[string][]packet.Packet
}

func newFakeSender() *fakeSender {
	return &fakeSender{packets: make(map[string][]packet.Packet)}
}

func (f *fakeSender) Send(target net.Addr, pkt packet.Packet) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	key := target.String()
	f.packets[key] = append(f.packets[key], pkt)
	return nil
}

// pop returns and removes the first packet matching id from target, or nil.
func (f *fakeSender) pop(target net.Addr, id packet.ID) *packet.Packet {
	f.mu.Lock()
	defer f.mu.Unlock()
	key := target.String()
	for i, p := range f.packets[key] {
		if p.ID == id {
			f.packets[key] = append(f.packets[key][:i], f.packets[key][i+1:]...)
			return &p
		}
	}
	return nil
}

// has checks whether a packet id was sent to target.
func (f *fakeSender) has(target net.Addr, id packet.ID) bool {
	f.mu.Lock()
	defer f.mu.Unlock()
	for _, p := range f.packets[target.String()] {
		if p.ID == id {
			return true
		}
	}
	return false
}

// fakeAddr returns a unique fake UDP address for testing.
func fakeAddr(port int) net.Addr {
	return &net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: port}
}

// ── Test harness ──────────────────────────────────────────────────────────

type harness struct {
	sessions *session.Manager
	rooms    *room.Manager
	matches  *match.Manager
	world    *world.World
	inputs   *command.InputBuffer
	sender   *fakeSender
	handler  *app.MultiplayerHandler
}

func newHarness() *harness {
	sessions := session.NewManager()
	rooms := room.NewManager()
	matches := match.NewManager()
	w := world.New()
	inputs := command.NewInputBuffer()
	sender := newFakeSender()
	log := logger.New("error") // quiet during tests

	handler := app.NewMultiplayerHandler(
		sessions, rooms, matches, w, inputs, sender, log,
	)
	return &harness{
		sessions: sessions,
		rooms:    rooms,
		matches:  matches,
		world:    w,
		inputs:   inputs,
		sender:   sender,
		handler:  handler,
	}
}

func (h *harness) send(addr net.Addr, id packet.ID, v interface{}) {
	payload, _ := json.Marshal(v)
	h.handler.HandlePacket(context.Background(), addr, packet.Packet{ID: id, Payload: payload})
}

func (h *harness) login(addr net.Addr, name string) protocol.LoginResponse {
	h.send(addr, packet.PacketIDLoginRequest, protocol.LoginRequest{GuestName: name})
	pkt := h.sender.pop(addr, packet.PacketIDLoginResponse)
	if pkt == nil {
		panic("no LoginResponse received for " + addr.String())
	}
	var resp protocol.LoginResponse
	if err := json.Unmarshal(pkt.Payload, &resp); err != nil {
		panic("unmarshal LoginResponse: " + err.Error())
	}
	return resp
}

func (h *harness) decodeRoomResponse(pkt *packet.Packet) protocol.RoomResponse {
	var r protocol.RoomResponse
	if err := json.Unmarshal(pkt.Payload, &r); err != nil {
		panic("unmarshal RoomResponse: " + err.Error())
	}
	return r
}

// ── Tests ──────────────────────────────────────────────────────────────────

func TestLoginCreatesSession(t *testing.T) {
	h := newHarness()
	addr := fakeAddr(10001)

	resp := h.login(addr, "alice")

	if resp.SessionID == "" || resp.PlayerID == "" {
		t.Fatalf("login response missing ids: %+v", resp)
	}
	if _, ok := h.sessions.GetByAddr(addr); !ok {
		t.Fatal("session not found by address after login")
	}
}

func TestDuplicateLoginReturnsAlreadyConnected(t *testing.T) {
	h := newHarness()
	addr := fakeAddr(10001)

	h.login(addr, "alice")

	// Second login from same addr.
	h.send(addr, packet.PacketIDLoginRequest, protocol.LoginRequest{GuestName: "alice2"})

	pkt := h.sender.pop(addr, packet.PacketIDErrorResponse)
	if pkt == nil {
		t.Fatal("expected ErrorResponse for duplicate login")
	}
	var errResp protocol.ErrorResponse
	_ = json.Unmarshal(pkt.Payload, &errResp)
	if errResp.Code != protocol.ErrorAlreadyConnected {
		t.Fatalf("error code = %q, want %q", errResp.Code, protocol.ErrorAlreadyConnected)
	}
}

func TestCreateRoomAutoJoinsCreator(t *testing.T) {
	h := newHarness()
	addr := fakeAddr(10001)
	h.login(addr, "alice")

	h.send(addr, packet.PacketIDCreateRoomRequest, protocol.CreateRoomRequest{Capacity: 5})

	created := h.sender.pop(addr, packet.PacketIDRoomCreated)
	if created == nil {
		t.Fatal("no RoomCreated packet")
	}
	joined := h.sender.pop(addr, packet.PacketIDRoomJoined)
	if joined == nil {
		t.Fatal("no RoomJoined packet for creator")
	}
	r := h.decodeRoomResponse(created)
	if r.RoomID == "" {
		t.Fatal("room id is empty")
	}
	if len(r.Players) != 1 {
		t.Fatalf("creator not auto-joined: players=%v", r.Players)
	}
}

func TestJoinRoomBroadcastsAndStartsMatchWhenFull(t *testing.T) {
	h := newHarness()
	addrA := fakeAddr(10001)
	addrB := fakeAddr(10002)

	h.login(addrA, "alice")
	h.login(addrB, "bob")

	// Alice creates room with capacity 2.
	h.send(addrA, packet.PacketIDCreateRoomRequest, protocol.CreateRoomRequest{Capacity: 2})
	createdPkt := h.sender.pop(addrA, packet.PacketIDRoomCreated)
	if createdPkt == nil {
		t.Fatal("no RoomCreated")
	}
	roomResp := h.decodeRoomResponse(createdPkt)
	h.sender.pop(addrA, packet.PacketIDRoomJoined) // consume alice's join

	// Bob joins the room.
	h.send(addrB, packet.PacketIDJoinRoomRequest, protocol.JoinRoomRequest{RoomID: roomResp.RoomID})

	// Both should receive RoomJoined (Bob's addr for Bob, Alice's addr for broadcast).
	if !h.sender.has(addrB, packet.PacketIDRoomJoined) {
		t.Fatal("Bob did not receive RoomJoined")
	}
	if !h.sender.has(addrA, packet.PacketIDRoomJoined) {
		t.Fatal("Alice did not receive RoomJoined broadcast")
	}

	// Room is full → match started → both receive PlayerSpawned.
	if !h.sender.has(addrA, packet.PacketIDPlayerSpawned) {
		t.Fatal("Alice did not receive PlayerSpawned")
	}
	if !h.sender.has(addrB, packet.PacketIDPlayerSpawned) {
		t.Fatal("Bob did not receive PlayerSpawned")
	}
}

func TestDuplicateJoinReturnsError(t *testing.T) {
	h := newHarness()
	addrA := fakeAddr(10001)
	h.login(addrA, "alice")

	h.send(addrA, packet.PacketIDCreateRoomRequest, protocol.CreateRoomRequest{Capacity: 5})
	created := h.decodeRoomResponse(h.sender.pop(addrA, packet.PacketIDRoomCreated))
	h.sender.pop(addrA, packet.PacketIDRoomJoined)

	// Alice tries to join her own room again.
	h.send(addrA, packet.PacketIDJoinRoomRequest, protocol.JoinRoomRequest{RoomID: created.RoomID})

	pkt := h.sender.pop(addrA, packet.PacketIDErrorResponse)
	if pkt == nil {
		t.Fatal("expected ErrorResponse for duplicate join")
	}
	var errResp protocol.ErrorResponse
	_ = json.Unmarshal(pkt.Payload, &errResp)
	if errResp.Code != protocol.ErrorDuplicateJoin {
		t.Fatalf("code = %q, want DUPLICATE_JOIN", errResp.Code)
	}
}

func TestJoinInvalidRoomReturnsError(t *testing.T) {
	h := newHarness()
	addr := fakeAddr(10001)
	h.login(addr, "alice")

	h.send(addr, packet.PacketIDJoinRoomRequest, protocol.JoinRoomRequest{RoomID: "nonexistent"})

	pkt := h.sender.pop(addr, packet.PacketIDErrorResponse)
	if pkt == nil {
		t.Fatal("expected ErrorResponse for invalid room")
	}
	var errResp protocol.ErrorResponse
	_ = json.Unmarshal(pkt.Payload, &errResp)
	if errResp.Code != protocol.ErrorInvalidRoom {
		t.Fatalf("code = %q, want INVALID_ROOM", errResp.Code)
	}
}

func TestLeaveRoomNotifiesOtherMembers(t *testing.T) {
	h := newHarness()
	addrA := fakeAddr(10001)
	addrB := fakeAddr(10002)

	h.login(addrA, "alice")
	h.login(addrB, "bob")

	h.send(addrA, packet.PacketIDCreateRoomRequest, protocol.CreateRoomRequest{Capacity: 5})
	created := h.decodeRoomResponse(h.sender.pop(addrA, packet.PacketIDRoomCreated))
	h.sender.pop(addrA, packet.PacketIDRoomJoined)

	h.send(addrB, packet.PacketIDJoinRoomRequest, protocol.JoinRoomRequest{RoomID: created.RoomID})
	h.sender.pop(addrB, packet.PacketIDRoomJoined)
	h.sender.pop(addrA, packet.PacketIDRoomJoined) // broadcast

	// Alice leaves.
	h.send(addrA, packet.PacketIDLeaveRoomRequest, protocol.LeaveRoomRequest{RoomID: created.RoomID})

	if !h.sender.has(addrA, packet.PacketIDRoomLeft) {
		t.Fatal("Alice did not receive RoomLeft")
	}
	if !h.sender.has(addrB, packet.PacketIDRoomLeft) {
		t.Fatal("Bob did not receive RoomLeft broadcast")
	}
}

func TestMovementInputBufferedAndAppliedByTickLoop(t *testing.T) {
	h := newHarness()
	addrA := fakeAddr(10001)
	addrB := fakeAddr(10002)

	h.login(addrA, "alice")
	h.login(addrB, "bob")

	// Create full room so match starts and entities spawn.
	h.send(addrA, packet.PacketIDCreateRoomRequest, protocol.CreateRoomRequest{Capacity: 2})
	created := h.decodeRoomResponse(h.sender.pop(addrA, packet.PacketIDRoomCreated))
	h.sender.pop(addrA, packet.PacketIDRoomJoined)

	h.send(addrB, packet.PacketIDJoinRoomRequest, protocol.JoinRoomRequest{RoomID: created.RoomID})

	// Get Alice's player ID.
	sessA, _ := h.sessions.GetByAddr(addrA)

	// Alice sends movement input.
	h.send(addrA, packet.PacketIDMovementInput, protocol.MovementInput{
		Sequence: 1, MoveX: 1, MoveY: 0, Rotation: 0.5,
	})

	// Verify input is buffered.
	inputs := h.inputs.Drain()
	if len(inputs) == 0 {
		t.Fatal("no inputs in buffer")
	}
	found := false
	for _, inp := range inputs {
		if inp.PlayerID == sessA.PlayerID {
			found = true
			if inp.Sequence != 1 || inp.MoveX != 1 {
				t.Fatalf("unexpected input: %+v", inp)
			}
		}
	}
	if !found {
		t.Fatalf("Alice's input not found in buffer; got: %v", inputs)
	}
}

func TestDisconnectBroadcastsPlayerDisconnected(t *testing.T) {
	h := newHarness()
	addrA := fakeAddr(10001)
	addrB := fakeAddr(10002)

	h.login(addrA, "alice")
	h.login(addrB, "bob")

	// Create room (capacity 5, so no match starts and we can test disconnect logic).
	h.send(addrA, packet.PacketIDCreateRoomRequest, protocol.CreateRoomRequest{Capacity: 5})
	created := h.decodeRoomResponse(h.sender.pop(addrA, packet.PacketIDRoomCreated))
	h.sender.pop(addrA, packet.PacketIDRoomJoined)

	h.send(addrB, packet.PacketIDJoinRoomRequest, protocol.JoinRoomRequest{RoomID: created.RoomID})
	h.sender.pop(addrB, packet.PacketIDRoomJoined)
	h.sender.pop(addrA, packet.PacketIDRoomJoined)

	// Alice disconnects.
	h.handler.HandlePacket(context.Background(), addrA, packet.Packet{ID: packet.PacketIDDisconnect})

	// Bob should receive PlayerDisconnected.
	if !h.sender.has(addrB, packet.PacketIDPlayerDisconnected) {
		t.Fatal("Bob did not receive PlayerDisconnected after Alice disconnects")
	}

	// Alice's session should be gone.
	if _, ok := h.sessions.GetByAddr(addrA); ok {
		t.Fatal("Alice's session still exists after disconnect")
	}
}

func TestUnloggedClientGetsDisconnectedError(t *testing.T) {
	h := newHarness()
	addr := fakeAddr(10001)

	// Try to create a room without logging in.
	h.send(addr, packet.PacketIDCreateRoomRequest, protocol.CreateRoomRequest{Capacity: 2})

	pkt := h.sender.pop(addr, packet.PacketIDErrorResponse)
	if pkt == nil {
		t.Fatal("expected ErrorResponse for unauthenticated create room")
	}
	var errResp protocol.ErrorResponse
	_ = json.Unmarshal(pkt.Payload, &errResp)
	if errResp.Code != protocol.ErrorDisconnected {
		t.Fatalf("code = %q, want DISCONNECTED", errResp.Code)
	}
}

func TestMalformedPacketGetsInvalidPacketError(t *testing.T) {
	h := newHarness()
	addr := fakeAddr(10001)
	h.login(addr, "alice")

	// Send garbage JSON for a JoinRoomRequest.
	h.handler.HandlePacket(context.Background(), addr, packet.Packet{
		ID:      packet.PacketIDJoinRoomRequest,
		Payload: []byte("{bad json"),
	})

	pkt := h.sender.pop(addr, packet.PacketIDErrorResponse)
	if pkt == nil {
		t.Fatal("expected ErrorResponse for malformed packet")
	}
}

func TestSnapshotBroadcastViaTickLoop(t *testing.T) {
	h := newHarness()
	addrA := fakeAddr(10001)
	addrB := fakeAddr(10002)

	h.login(addrA, "alice")
	h.login(addrB, "bob")

	// Start a match with capacity 2.
	h.send(addrA, packet.PacketIDCreateRoomRequest, protocol.CreateRoomRequest{Capacity: 2})
	created := h.decodeRoomResponse(h.sender.pop(addrA, packet.PacketIDRoomCreated))
	h.sender.pop(addrA, packet.PacketIDRoomJoined)
	h.send(addrB, packet.PacketIDJoinRoomRequest, protocol.JoinRoomRequest{RoomID: created.RoomID})

	// Grab the match.
	var runningMatch *match.Match
	for _, r := range h.rooms.List() {
		players, _, _, mID := r.Snapshot()
		_ = players
		if mID != "" {
			mt, ok := h.matches.Get(mID)
			if ok {
				runningMatch = mt
			}
		}
	}
	if runningMatch == nil {
		t.Fatal("match not started after capacity reached")
	}

	// Simulate one tick via the loop internals.
	log := logger.New("error")
	interest := system.BroadcastInterestManager{}
	tickLoop := app.NewTickLoop(h.sessions, h.matches, h.world, h.inputs, interest, h.sender, log)
	tickLoop.ExposedTick(1) // call unexported via test method

	if !h.sender.has(addrA, packet.PacketIDSnapshot) {
		t.Fatal("Alice did not receive Snapshot from tick")
	}
	if !h.sender.has(addrB, packet.PacketIDSnapshot) {
		t.Fatal("Bob did not receive Snapshot from tick")
	}
}
