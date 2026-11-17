package api

import (
	"errors"
	"sync/atomic"
	"testing"
)

var errRealPunchBlocked = errors.New("real clock-in blocked: dev/test run is not using the fake PontoMais server")

var realPunchBlocked atomic.Bool

// BlockRealPunches makes every punch aimed at the real PontoMais fail. Dev binaries
// (cmd/pm-dev) call it at startup; Go tests are blocked automatically.
func BlockRealPunches() { realPunchBlocked.Store(true) }

// guardRealPunch refuses to write a punch to production from tests or dev runs.
func guardRealPunch(c *Client) error {
	if (testing.Testing() || realPunchBlocked.Load()) && c.baseURL() == BaseURL {
		return errRealPunchBlocked
	}
	return nil
}
