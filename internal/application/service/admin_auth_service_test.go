package service

import (
	"sync"
	"testing"
)

func TestAdminAuth_LoginSuccess(t *testing.T) {
	svc := NewAdminAuthService("admin", "secret")
	token, err := svc.Login("admin", "secret")
	if err != nil {
		t.Fatalf("Login: %v", err)
	}
	if token == "" {
		t.Error("Login returned empty token")
	}
}

func TestAdminAuth_LoginWrongLogin(t *testing.T) {
	svc := NewAdminAuthService("admin", "secret")
	_, err := svc.Login("wrong", "secret")
	if err != ErrForbidden {
		t.Errorf("Login(wrong login) error = %v, want ErrForbidden", err)
	}
}

func TestAdminAuth_LoginWrongPassword(t *testing.T) {
	svc := NewAdminAuthService("admin", "secret")
	_, err := svc.Login("admin", "wrong")
	if err != ErrForbidden {
		t.Errorf("Login(wrong password) error = %v, want ErrForbidden", err)
	}
}

func TestAdminAuth_LoginEmpty(t *testing.T) {
	svc := NewAdminAuthService("admin", "secret")
	_, err := svc.Login("", "")
	if err != ErrForbidden {
		t.Errorf("Login(empty) error = %v, want ErrForbidden", err)
	}
}

func TestAdminAuth_ValidateValid(t *testing.T) {
	svc := NewAdminAuthService("admin", "secret")
	token, _ := svc.Login("admin", "secret")
	if !svc.Validate(token) {
		t.Error("Validate(valid token) = false, want true")
	}
}

func TestAdminAuth_ValidateInvalid(t *testing.T) {
	svc := NewAdminAuthService("admin", "secret")
	if svc.Validate("nonexistent") {
		t.Error("Validate(invalid token) = true, want false")
	}
}

func TestAdminAuth_ValidateEmpty(t *testing.T) {
	svc := NewAdminAuthService("admin", "secret")
	if svc.Validate("") {
		t.Error("Validate(empty) = true, want false")
	}
}

func TestAdminAuth_Logout(t *testing.T) {
	svc := NewAdminAuthService("admin", "secret")
	token, _ := svc.Login("admin", "secret")
	svc.Logout(token)
	if svc.Validate(token) {
		t.Error("Validate after Logout = true, want false")
	}
}

func TestAdminAuth_UniqueTokens(t *testing.T) {
	svc := NewAdminAuthService("admin", "secret")
	t1, _ := svc.Login("admin", "secret")
	t2, _ := svc.Login("admin", "secret")
	if t1 == t2 {
		t.Error("two Login calls returned the same token")
	}
}

func TestAdminAuth_concurrent(t *testing.T) {
	svc := NewAdminAuthService("admin", "secret")
	var wg sync.WaitGroup
	const n = 50

	tokens := make([]string, n)
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			tok, err := svc.Login("admin", "secret")
			if err != nil {
				t.Errorf("goroutine %d Login error: %v", idx, err)
				return
			}
			tokens[idx] = tok
		}(i)
	}
	wg.Wait()

	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			if tokens[idx] == "" {
				return
			}
			if !svc.Validate(tokens[idx]) {
				t.Errorf("goroutine %d Validate failed", idx)
			}
			svc.Logout(tokens[idx])
		}(i)
	}
	wg.Wait()
}
