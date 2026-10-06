package internal_test

import (
	"bytes"
	"encoding/json"
	"io"
	"os/exec"
	"strings"
	"testing"
)

const modulePath = "github.com/fintech/kyc/"

// TestDependencyRule enforces AGENTS.md: domain imports no other layer, and
// application imports domain only. Checked over transitive deps.
func TestDependencyRule(t *testing.T) {
	rules := []struct {
		pattern string
		allowed string // repo-internal prefix this layer may depend on
	}{
		{"./internal/domain/...", modulePath + "internal/domain/"},
		{"./internal/application/...", modulePath + "internal/application/"},
	}
	for _, r := range rules {
		t.Run(r.pattern, func(t *testing.T) {
			cmd := exec.Command("go", "list", "-deps", "-json", r.pattern)
			cmd.Dir = ".."
			var stderr bytes.Buffer
			cmd.Stderr = &stderr
			out, err := cmd.Output()
			if err != nil {
				t.Fatalf("go list: %v\n%s", err, stderr.String())
			}
			dec := json.NewDecoder(bytes.NewReader(out))
			for {
				var pkg struct{ ImportPath string }
				if err := dec.Decode(&pkg); err == io.EOF {
					break
				} else if err != nil {
					t.Fatal(err)
				}
				p := pkg.ImportPath
				if !strings.HasPrefix(p, modulePath) {
					continue
				}
				if strings.HasPrefix(p, modulePath+"internal/domain/") ||
					strings.HasPrefix(p, r.allowed) {
					continue
				}
				t.Errorf("forbidden dependency: %s", p)
			}
		})
	}
}
