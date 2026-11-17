package api

import (
	"context"
	"errors"
	"testing"
)

func TestRegisterPunchNeverHitsProduction(t *testing.T) {
	c := &Client{BaseURL: BaseURL, AuthHeaders: func() (map[string]string, error) {
		t.Fatal("must not authenticate against production")
		return nil, nil
	}}
	_, err := c.RegisterPunch(context.Background(), Location{Latitude: 1, Longitude: 1})
	if !errors.Is(err, errRealPunchBlocked) {
		t.Fatalf("want errRealPunchBlocked, got %v", err)
	}
	// Empty BaseURL also resolves to production.
	_, err = (&Client{}).RegisterPunch(context.Background(), Location{})
	if !errors.Is(err, errRealPunchBlocked) {
		t.Fatalf("empty BaseURL: want errRealPunchBlocked, got %v", err)
	}
}
