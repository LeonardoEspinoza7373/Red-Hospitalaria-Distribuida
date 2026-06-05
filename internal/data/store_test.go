package data

import (
	"os"
	"path/filepath"
	"testing"
)

func TestUserStore(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "users.json")

	store := NewUserStore(path)
	if err := store.Load(); err != nil {
		t.Fatal(err)
	}

	u := &User{
		Username:    "admin",
		Password:    "hash",
		DisplayName: "Administrador",
		Role:        "admin",
		HospitalID:  4,
	}
	if err := store.Create(u); err != nil {
		t.Fatal(err)
	}
	if u.ID != 1 {
		t.Fatalf("expected id=1, got %d", u.ID)
	}

	u2 := &User{
		Username:    "doctor1",
		Password:    "hash2",
		DisplayName: "Doctor Uno",
		Role:        "doctor",
		HospitalID:  1,
	}
	if err := store.Create(u2); err != nil {
		t.Fatal(err)
	}
	if u2.ID != 2 {
		t.Fatalf("expected id=2, got %d", u2.ID)
	}

	// Test GetByUsername
	found, err := store.GetByUsername("admin")
	if err != nil {
		t.Fatal(err)
	}
	if found.Username != "admin" {
		t.Fatalf("expected admin, got %s", found.Username)
	}

	// Test List
	users, err := store.List()
	if err != nil {
		t.Fatal(err)
	}
	if len(users) != 2 {
		t.Fatalf("expected 2 users, got %d", len(users))
	}

	// Test persistence (reload from disk)
	store2 := NewUserStore(path)
	if err := store2.Load(); err != nil {
		t.Fatal(err)
	}
	users2, _ := store2.List()
	if len(users2) != 2 {
		t.Fatalf("expected 2 users after reload, got %d", len(users2))
	}

	// Test Update
	u.DisplayName = "Admin Actualizado"
	if err := store.Update(u); err != nil {
		t.Fatal(err)
	}
	updated, _ := store.GetByID(1)
	if updated.DisplayName != "Admin Actualizado" {
		t.Fatalf("expected updated display name")
	}

	// Test Delete
	if err := store.Delete(2); err != nil {
		t.Fatal(err)
	}
	users3, _ := store.List()
	if len(users3) != 1 {
		t.Fatalf("expected 1 user after delete, got %d", len(users3))
	}

	// Test GetByID not found
	_, err = store.GetByID(999)
	if err == nil {
		t.Fatal("expected error for non-existent id")
	}

	// Test GetByUsername not found
	_, err = store.GetByUsername("nonexistent")
	if err == nil {
		t.Fatal("expected error for non-existent username")
	}

	// Clean up
	os.Remove(path)
}
