package session

import (
	"errors"
	"net"
	"testing"
)

func TestCreateStoresGuestSessionByAddressAndPlayer(t *testing.T) {
	manager := NewManager()
	addr := &net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: 10001}

	created, err := manager.Create(addr, "guest-a")
	if err != nil {
		t.Fatalf("create session: %v", err)
	}

	byAddr, ok := manager.GetByAddr(addr)
	if !ok {
		t.Fatal("session not found by address")
	}
	if byAddr.ID == "" || byAddr.PlayerID == "" || byAddr.GuestName != "guest-a" {
		t.Fatalf("unexpected session: %#v", byAddr)
	}
	if byAddr != created {
		t.Fatal("address lookup returned different session pointer")
	}

	byPlayer, ok := manager.GetByPlayerID(created.PlayerID)
	if !ok || byPlayer != created {
		t.Fatalf("session not found by player id: ok=%v session=%#v", ok, byPlayer)
	}
}

func TestCreateRejectsDuplicateAddress(t *testing.T) {
	manager := NewManager()
	addr := &net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: 10001}

	if _, err := manager.Create(addr, "guest-a"); err != nil {
		t.Fatalf("create first session: %v", err)
	}

	if _, err := manager.Create(addr, "guest-b"); !errors.Is(err, ErrAlreadyConnected) {
		t.Fatalf("duplicate create error = %v, want %v", err, ErrAlreadyConnected)
	}
}

func TestRemoveByAddrDeletesSessionIndexes(t *testing.T) {
	manager := NewManager()
	addr := &net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: 10001}
	created, err := manager.Create(addr, "guest-a")
	if err != nil {
		t.Fatalf("create session: %v", err)
	}

	removed, ok := manager.RemoveByAddr(addr)
	if !ok {
		t.Fatal("expected session removal")
	}
	if removed.PlayerID != created.PlayerID {
		t.Fatalf("removed player id = %q, want %q", removed.PlayerID, created.PlayerID)
	}

	if _, ok := manager.GetByAddr(addr); ok {
		t.Fatal("session still found by address after removal")
	}
	if _, ok := manager.GetByPlayerID(created.PlayerID); ok {
		t.Fatal("session still found by player id after removal")
	}
}
