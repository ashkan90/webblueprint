package api

import (
	"golang.org/x/crypto/bcrypt"
	"testing"
)

func TestX(t *testing.T) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("default_password"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatal(err)
	}

	t.Log(string(hashedPassword))
}
