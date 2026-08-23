package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// Characterisation tests for #377, written BEFORE the refactor and pinning what
// the commands do TODAY.
//
// 🛑 Why these exist and why they came first (@Clayton's call, and he was right):
// the refactor deletes one of three identical-looking validators and merges two
// loops. Nothing fails loudly if the wrong copy goes, or if a merged loop drops
// a behaviour one caller relied on. These are the thing the refactor can be
// WRONG against.
//
// ⚠️ If a later commit needs to change an expectation here, that is a BEHAVIOUR
// CHANGE and must be stated as one — not absorbed into a "cleanup" diff.

// builtBin is compiled once per `go test` run, from the CURRENT source.
var builtBin string

// TestMain builds the binary these tests exec, instead of borrowing
// build/bin/pfw.
//
// 🛑 Borrowing it made this suite LIE IN THE REASSURING DIRECTION (@Sten). The
// old helper skipped when the binary was ABSENT but ran happily when it was
// STALE, so the same source tree gave opposite verdicts depending on whether you
// had rebuilt:
//
//	go test                 PASS   ← graded a binary from some earlier commit
//	make build && go test   FAIL   ← graded the code you actually wrote
//
// A test that passes against an image nobody is shipping is worse than no test:
// it certifies the wrong artifact and reports success. @Sten hit it landing a
// fix for a defect these very tests were meant to pin.
//
// ⚑ And building to a TEMP path is load-bearing for a second reason: every bench
// shares ~/.local/bin/pfw → build/bin/pfw, and `room watch --loop` re-execs when
// that image changes (#383). Compiling into build/bin as part of `go test` would
// push a test binary onto ~28 live watchers. Running the tests must not deploy.
func TestMain(m *testing.M) {
	dir, err := os.MkdirTemp("", "pfw-contract-")
	if err != nil {
		fmt.Fprintf(os.Stderr, "contract tests: %v\n", err)
		os.Exit(1)
	}
	defer os.RemoveAll(dir)

	builtBin = filepath.Join(dir, "pfw")
	build := exec.Command("go", "build", "-o", builtBin, ".")
	build.Stderr = os.Stderr
	if err := build.Run(); err != nil {
		// ⛔ Fail, never skip. A skip here reads as "no binary available" —
		// which is what the old helper said — and a suite that silently
		// declines to run is indistinguishable from one that passed.
		fmt.Fprintf(os.Stderr, "contract tests: cannot build the binary under test: %v\n", err)
		os.Exit(1)
	}

	code := m.Run()
	os.RemoveAll(dir)
	os.Exit(code)
}

func pfwBin(t *testing.T) string {
	t.Helper()
	if builtBin == "" {
		t.Fatal("binary under test was never built — TestMain did not run")
	}
	return builtBin
}

// run captures stdout and stderr SEPARATELY. Merging them is the single most
// repeated measurement error in this codebase's history — it makes a stderr line
// indistinguishable from a stdout one, and the whole wake/telemetry contract is
// exactly that distinction.
func run(t *testing.T, bin string, args ...string) (stdout, stderr string, code int) {
	t.Helper()
	cmd := exec.Command(bin, args...)
	var out, errb strings.Builder
	cmd.Stdout, cmd.Stderr = &out, &errb
	err := cmd.Run()
	if ee, ok := err.(*exec.ExitError); ok {
		code = ee.ExitCode()
	} else if err != nil {
		t.Fatalf("%v: %v", args, err)
	}
	return out.String(), errb.String(), code
}

// Both commands refuse --loop without --interval, LOUDLY and on STDOUT, and do
// so without needing any credentials. Three copies of this validation exist
// today; after the refactor there must be one, and this must still hold for both.
func TestLoopWithoutIntervalRefusesOnStdout(t *testing.T) {
	bin := pfwBin(t)
	for _, tc := range []struct {
		name string
		args []string
	}{
		{"room watch", []string{"room", "watch", "some-id", "--loop"}},
		{"room health", []string{"room", "health", "/tmp/nowhere", "--loop"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			stdout, _, code := run(t, bin, tc.args...)
			if !strings.Contains(stdout, "requires --interval") {
				t.Errorf("refusal not on STDOUT (got %q) — stderr reaches nobody when a Monitor is the consumer", stdout)
			}
			if code == 0 {
				t.Error("exit 0 on misuse: a caller checking $? would proceed as though it were looping")
			}
		})
	}
}

// The mirror case. --interval without --loop is a no-op that LOOKS configured,
// which is the failure class this whole epic is about.
func TestIntervalWithoutLoopRefusesOnStdout(t *testing.T) {
	bin := pfwBin(t)
	for _, tc := range []struct {
		name string
		args []string
	}{
		{"room watch", []string{"room", "watch", "some-id", "--interval", "30m"}},
		{"room health", []string{"room", "health", "/tmp/nowhere", "--interval", "5m"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			stdout, _, code := run(t, bin, tc.args...)
			if !strings.Contains(stdout, "no effect without --loop") {
				t.Errorf("silent no-op or wrong stream (stdout %q)", stdout)
			}
			if code == 0 {
				t.Error("exit 0: the caller believes it is polling repeatedly and it is not")
			}
		})
	}
}

// ⚠️ Misuse must be validated BEFORE config. An operator with a half-set-up
// environment must get the flag error, not "PROSEFORGE_URL is required" — then
// fix the config and hit the real error on a second run.
func TestFlagMisuseIsReportedBeforeMissingConfig(t *testing.T) {
	bin := pfwBin(t)
	cmd := exec.Command(bin, "room", "watch", "some-id", "--loop")
	cmd.Env = []string{"PATH=/usr/bin:/bin"} // no PROSEFORGE_URL, no token
	out, _ := cmd.Output()
	if !strings.Contains(string(out), "requires --interval") {
		t.Errorf("config error preempted the flag error: %q", string(out))
	}
}

// A single-leg `room health` prints ONE line: "<state> rc=<n> <detail>". Benches
// grep this. The multi-leg path must not have changed it, and the refactor must
// not either.
func TestSingleLegHealthOutputContract(t *testing.T) {
	bin := pfwBin(t)
	dir := t.TempDir()
	stdout, _, code := run(t, bin, "room", "health", dir, "--max-age", "5m")

	first := strings.SplitN(strings.TrimSpace(stdout), "\n", 2)[0]
	fields := strings.Fields(first)
	if len(fields) < 2 {
		t.Fatalf("line one has too few fields: %q", first)
	}
	if !strings.HasPrefix(fields[1], "rc=") {
		t.Errorf("field 2 = %q, want rc=… — the exit code must travel IN the payload, "+
			"because `| head -1` returns head's status and not ours", fields[1])
	}
	if code != 3 {
		t.Errorf("exit = %d, want 3 for a dir with no watcher state", code)
	}
}

// Multi-leg prints a VERDICT first, then one line per leg. `| head -1` must give
// the answer for the SET, not the first leg's state.
func TestMultiLegVerdictComesFirst(t *testing.T) {
	bin := pfwBin(t)
	a, b := t.TempDir(), t.TempDir()
	stdout, _, _ := run(t, bin, "room", "health", a, b, "--max-age", "5m")

	first := strings.SplitN(strings.TrimSpace(stdout), "\n", 2)[0]
	if !strings.Contains(first, "unhealthy") || !strings.Contains(first, "rc=") {
		t.Errorf("line one = %q, want the set verdict with rc= — a reader piping to head "+
			"must get the answer, not leg one", first)
	}
	if strings.Count(stdout, a) != 1 || strings.Count(stdout, b) != 1 {
		t.Errorf("every leg must be named exactly once; got:\n%s", stdout)
	}
}

// Per-leg thresholds (#372). A colon inside a path must not be eaten as a
// separator, and an unparseable suffix stays part of the path.
func TestPerLegThresholdIsAppliedAndEchoed(t *testing.T) {
	bin := pfwBin(t)
	a, b := t.TempDir(), t.TempDir()
	stdout, _, _ := run(t, bin, "room", "health", a+":90m", b+":30s", "--max-age", "5m")

	for _, want := range []string{"1h30m0s", "30s"} {
		if !strings.Contains(stdout, want) {
			t.Errorf("applied threshold %q not echoed — an override that is silently "+
				"ignored looks identical to one that worked:\n%s", want, stdout)
		}
	}
	if strings.Contains(stdout, "5m0s") {
		t.Errorf("the global --max-age leaked onto a leg that overrode it:\n%s", stdout)
	}
}

// The healthy path says NOTHING on stdout in loop mode. Silence is the output;
// a checker that speaks every tick is one its operator learns to skip.
func TestHealthLoopIsSilentWhileEverythingIsHealthy(t *testing.T) {
	bin := pfwBin(t)
	dir := t.TempDir()
	state := filepath.Join(dir, "state")
	if err := exec.Command("mkdir", "-p", state).Run(); err != nil {
		t.Fatal(err)
	}
	stamp := time.Now().UTC().Format(time.RFC3339)
	if err := exec.Command("sh", "-c",
		"printf '%s\\n' "+stamp+" > "+filepath.Join(state, "health-poll")).Run(); err != nil {
		t.Fatal(err)
	}

	cmd := exec.Command(bin, "room", "health", state+":5m", "--loop", "--interval", "1s")
	var out, errb strings.Builder
	cmd.Stdout, cmd.Stderr = &out, &errb
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	time.Sleep(2500 * time.Millisecond)
	_ = cmd.Process.Kill()
	_ = cmd.Wait()

	if strings.TrimSpace(out.String()) != "" {
		t.Errorf("loop spoke on stdout while healthy — every line here is a wake:\n%s", out.String())
	}
	if !strings.Contains(errb.String(), "wd ") {
		t.Errorf("no per-check telemetry on stderr; a running-but-silent checker is "+
			"indistinguishable from a stopped one:\n%s", errb.String())
	}
}

// 🛑 <dir>:<max-age> must work with ONE leg, not only with several.
//
// I applied parseLegSpec to the multi-leg path only, so `room health <dir>:5m`
// with a single leg treated the whole string as a directory name and reported
// `unknown` — the most alarming state the command has — about a healthy watcher.
// That is #361's exact shape, which I fixed in the morning and reintroduced in a
// different function the same afternoon.
//
// ⚠️ The characterisation suite missed it because every existing row passed a
// PLAIN directory. The syntax I had just documented was the one path nothing
// exercised — a test suite is only as good as the inputs someone thought to use.
func TestPerLegThresholdWorksWithASingleLeg(t *testing.T) {
	bin := pfwBin(t)
	dir := t.TempDir()
	state := filepath.Join(dir, "state")
	if err := exec.Command("mkdir", "-p", state).Run(); err != nil {
		t.Fatal(err)
	}
	stamp := time.Now().UTC().Format(time.RFC3339)
	if err := exec.Command("sh", "-c",
		"printf '%s\\n' "+stamp+" > "+filepath.Join(state, "health-poll")).Run(); err != nil {
		t.Fatal(err)
	}

	plain, _, plainCode := run(t, bin, "room", "health", state, "--max-age", "5m")
	suffix, _, suffixCode := run(t, bin, "room", "health", state+":5m")

	plainState := strings.Fields(plain)[0]
	suffixState := strings.Fields(suffix)[0]
	if plainState != suffixState || plainCode != suffixCode {
		t.Errorf("the two spellings of one threshold disagree:\n  plain  %q rc=%d\n  suffix %q rc=%d\n"+
			"a suffix parsed as part of the path reports a LIVE watcher dead",
			plainState, plainCode, suffixState, suffixCode)
	}

	// And the suffix must actually be APPLIED, not merely stripped.
	//
	// ⚠️ The stamp is BACKDATED. My first version wrote it `now` and checked it
	// against a 1s threshold in the same instant — which is legitimately `alive`,
	// so the test failed and briefly read as "the fix does not work." A threshold
	// test needs an age it can actually exceed.
	old := time.Now().Add(-30 * time.Second).UTC().Format(time.RFC3339)
	if err := exec.Command("sh", "-c",
		"printf '%s\\n' "+old+" > "+filepath.Join(state, "health-poll")).Run(); err != nil {
		t.Fatal(err)
	}
	tight, _, tightCode := run(t, bin, "room", "health", state+":1s")
	if strings.Fields(tight)[0] != "stale" || tightCode == 0 {
		t.Errorf("single-leg threshold parsed but not honoured: got %q rc=%d, want stale/non-zero",
			strings.Fields(tight)[0], tightCode)
	}
}
