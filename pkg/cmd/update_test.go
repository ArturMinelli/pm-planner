package cmd

import (
	"bytes"
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	"pm-cli/pkg/message"
	"pm-cli/pkg/update"
)

func TestPrintStatusHumanReadable(t *testing.T) {
	cases := []struct {
		name   string
		status update.Status
		want   string
	}{
		{
			name:   "up to date",
			status: update.Status{Root: "/root", IsGit: true, CommitSHA: "a1b2c3d", Behind: 0},
			want:   "PM Planner está atualizado.",
		},
		{
			name:   "single update",
			status: update.Status{Root: "/root", IsGit: true, Behind: 1},
			want:   "1 atualização disponível.",
		},
		{
			name:   "several updates",
			status: update.Status{Root: "/root", IsGit: true, Behind: 4},
			want:   "4 atualizações disponíveis.",
		},
		{
			name:   "tarball install",
			status: update.Status{Root: "/root", IsGit: false},
			want:   "Versão atual: desconhecida",
		},
		{
			name:   "blocked",
			status: update.Status{Root: "/root", IsGit: true, Dirty: true, Blockers: []message.Message{
				message.New(message.KeyUpdateBlockersDirtyWorkingTree, nil),
			}},
			want:   "! Uncommitted local changes",
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			var out bytes.Buffer
			if err := printStatus(&out, &testCase.status, false); err != nil {
				t.Fatalf("printStatus: %v", err)
			}
			if !strings.Contains(out.String(), testCase.want) {
				t.Fatalf("output %q does not contain %q", out.String(), testCase.want)
			}
		})
	}
}

func TestPrintStatusJSONIsMachineReadable(t *testing.T) {
	var out bytes.Buffer
	status := update.Status{Root: "/root", IsGit: true, CommitSHA: "a1b2c3d", Behind: 2, UpdateAvailable: true}

	if err := printStatus(&out, &status, true); err != nil {
		t.Fatalf("printStatus: %v", err)
	}

	decoded := update.Status{}
	if err := json.Unmarshal(out.Bytes(), &decoded); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if !reflect.DeepEqual(decoded, status) {
		t.Fatalf("decoded %+v, want %+v", decoded, status)
	}
}
