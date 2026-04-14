package model

import "testing"

func TestAdminPasswordLifecycle(t *testing.T) {
	admin := &Admin{Username: "admin"}

	if err := admin.SetPassword("secret123"); err != nil {
		t.Fatalf("SetPassword() error = %v", err)
	}

	if admin.PasswordHash == "" {
		t.Fatal("expected password hash to be generated")
	}

	if admin.PasswordHash == "secret123" {
		t.Fatal("expected password to be hashed")
	}

	if !admin.CheckPassword("secret123") {
		t.Fatal("expected password check to succeed")
	}

	if admin.CheckPassword("wrong-password") {
		t.Fatal("expected password check to fail for wrong password")
	}
}
