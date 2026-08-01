package storage

import (
	"path/filepath"
	"testing"
)

func TestUpdateClientAllowedIPs(t *testing.T) {
	stor, err := New(filepath.Join(t.TempDir(), "minion.db"))
	if err != nil {
		t.Fatalf("create storage: %v", err)
	}
	t.Cleanup(func() { _ = stor.DB.Close() })

	if err := stor.InsertClient("bootstrap", "127.0.0.1/32", "hash"); err != nil {
		t.Fatalf("insert bootstrap client: %v", err)
	}
	if err := stor.UpdateClientAllowedIPs("bootstrap", "192.0.2.10/32"); err != nil {
		t.Fatalf("update allowlist: %v", err)
	}

	clients, err := stor.GetClients()
	if err != nil {
		t.Fatalf("list clients: %v", err)
	}
	if len(clients) != 1 || len(clients[0].AllowedIPs) != 1 || clients[0].AllowedIPs[0] != "192.0.2.10/32" {
		t.Fatalf("unexpected clients after update: %#v", clients)
	}
}

func TestUpdateClientAllowedIPsRejectsMissingClient(t *testing.T) {
	stor, err := New(filepath.Join(t.TempDir(), "minion.db"))
	if err != nil {
		t.Fatalf("create storage: %v", err)
	}
	t.Cleanup(func() { _ = stor.DB.Close() })

	if err := stor.UpdateClientAllowedIPs("bootstrap", "192.0.2.10/32"); err == nil {
		t.Fatal("expected missing client error")
	}
}
