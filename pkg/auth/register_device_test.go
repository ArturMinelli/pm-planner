package auth

import (
	"testing"
	"time"
)

func TestSessionToHeadersIncludesTokenMetadata(t *testing.T) {
	s := &session{
		Token:     "token",
		Uid:       "user@example.com",
		Client:    "client",
		TokenType: "Bearer",
		Expiry:    2000000000,
		UUID:      "device-1",
	}
	headers := sessionToHeaders(s)
	if headers["Token-Type"] != "Bearer" {
		t.Fatalf("Token-Type: %q", headers["Token-Type"])
	}
	if headers["Expiry"] != "2000000000" {
		t.Fatalf("Expiry: %q", headers["Expiry"])
	}
}

func TestBuildRegisterDeviceIncludesSessionUUID(t *testing.T) {
	setupAuthTestHome(t)
	s := session{
		Token:         "token",
		Uid:           "user@example.com",
		Client:        "client",
		TokenType:     "Bearer",
		Expiry:        time.Now().Add(time.Hour).Unix(),
		UUID:          "device-1",
		SignInSuccess: "ok",
		SignInCount:   2,
		CachedAt:      time.Now(),
	}
	if err := writeCachedSession(&s); err != nil {
		t.Fatal(err)
	}
	device, err := BuildRegisterDevice()
	if err != nil {
		t.Fatal(err)
	}
	uuidObj, ok := device["uuid"].(map[string]any)
	if !ok || uuidObj["uuid"] != "device-1" {
		t.Fatalf("device uuid: %#v", device["uuid"])
	}
}
