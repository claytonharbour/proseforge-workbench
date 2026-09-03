package main

import (
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"
)

// A clock the test drives, so a flap is fabricated rather than waited for.
func fixedAlarm() (*roomAlarm, func(time.Duration)) { return fixedAlarmAfter(1) }

func fixedAlarmAfter(n int) (*roomAlarm, func(time.Duration)) {
	at := time.Date(2026, 8, 22, 12, 0, 0, 0, time.UTC)
	a := newRoomAlarm(n)
	a.now = func() time.Time { return at }
	return a, func(d time.Duration) { at = at.Add(d) }
}

var boom = errors.New("connection refused")

// ① FIRES. A forced failure produces exactly one alarm — not zero.
//
// 🛑 This is the state my own shell threshold made unreachable. I required 3
// CONSECUTIVE failures, so my real rc=12 outage stayed silent at 1/3 of the
// threshold while @Wayland's identical outage alarmed correctly. Same class,
// same night, same machine: the alarm I depended on was the one my own
// configuration made hardest to ever see fire.
func TestAlarmFiresOnEntryToFailure(t *testing.T) {
	a, _ := fixedAlarm()
	line := a.observe(false, 12, boom)
	if line == "" {
		t.Fatal("a failing tick produced NO alarm — silence is the failure this whole area exists to prevent")
	}
	if !strings.Contains(line, "FAILING") || !strings.Contains(line, "rc=12") {
		t.Errorf("alarm %q must name the state and the code", line)
	}
	// ⚑ And it must say the cursor held. A reader seeing a failure needs to
	// know whether messages were LOST; a failed poll never advances the cursor.
	if !strings.Contains(line, "nothing lost") {
		t.Errorf("alarm %q does not tell the reader whether anything was lost", line)
	}
}

// ② QUIET. Healthy ticks produce nothing at all — the common case.
func TestAlarmIsSilentWhileHealthy(t *testing.T) {
	a, adv := fixedAlarm()
	for i := 0; i < 200; i++ {
		if line := a.observe(true, 0, nil); line != "" {
			t.Fatalf("healthy tick %d spoke: %q — routine noise trains the reader to ignore the channel", i, line)
		}
		adv(time.Minute)
	}
}

// ③ FALSE-ALARM RATE — the one nobody had measured, and the one that decides
// whether the alarm is believed at all.
//
// ⚠️ NOT reachable by running longer: 1,730 consecutive clean ticks proved
// nothing about a flap. It has to be fabricated, which is the entire reason
// this logic was lifted out of the loop body.
func TestATransientFlapProducesExactlyTwoLines(t *testing.T) {
	a, adv := fixedAlarm()

	var spoke []string
	say := func(line string) {
		if line != "" {
			spoke = append(spoke, line)
		}
	}

	say(a.observe(true, 0, nil)) // healthy
	adv(60 * time.Second)
	say(a.observe(false, 12, boom)) // blip starts
	adv(60 * time.Second)
	say(a.observe(false, 12, boom)) // still down
	adv(60 * time.Second)
	say(a.observe(true, 0, nil)) // back
	adv(60 * time.Second)
	say(a.observe(true, 0, nil)) // and stays back

	if len(spoke) != 2 {
		t.Fatalf("a 2-tick blip produced %d lines, want exactly 2 (one alarm, one recovery):\n%s",
			len(spoke), strings.Join(spoke, "\n"))
	}
	if !strings.Contains(spoke[0], "FAILING") {
		t.Errorf("first line %q is not the alarm", spoke[0])
	}
	if !strings.Contains(spoke[1], "RECOVERED") {
		t.Errorf("second line %q is not the recovery", spoke[1])
	}
	// The recovery must size the gap — "it came back" without "how long were we
	// blind" leaves the reader unable to judge whether to go looking.
	if !strings.Contains(spoke[1], "2 failed tick(s)") || !strings.Contains(spoke[1], "2m0s") {
		t.Errorf("recovery %q does not name how many ticks or how long", spoke[1])
	}
}

// ⚑ AT MOST 2 LINES PER OUTAGE REGARDLESS OF INTERVAL. Level-triggering makes
// the line count scale with duration ÷ interval, which quietly makes --interval
// load-bearing: copy the command with a shorter one and a well-behaved alarm
// becomes a flood without anybody changing the alarm.
func TestSustainedFailureDoesNotReAlarmEveryTick(t *testing.T) {
	a, adv := fixedAlarm()
	lines := 0
	for i := 0; i < alarmEscalateEvery-1; i++ {
		if a.observe(false, 12, boom) != "" {
			lines++
		}
		adv(time.Second)
	}
	if lines != 1 {
		t.Fatalf("%d ticks of one outage produced %d lines, want 1", alarmEscalateEvery-1, lines)
	}
}

// ...but NEVER SILENT FOREVER either. An outage that stops being mentioned
// reads as an outage that ended.
func TestLongOutageEscalatesPeriodically(t *testing.T) {
	a, adv := fixedAlarm()
	var spoke []string
	for i := 0; i < alarmEscalateEvery*2; i++ {
		if line := a.observe(false, 12, boom); line != "" {
			spoke = append(spoke, line)
		}
		adv(time.Second)
	}
	// entry + two escalations at 30 and 60
	if len(spoke) != 3 {
		t.Fatalf("%d failing ticks produced %d lines, want 3 (entry + 2 escalations):\n%s",
			alarmEscalateEvery*2, len(spoke), strings.Join(spoke, "\n"))
	}
	if !strings.Contains(spoke[1], "STILL FAILING") {
		t.Errorf("escalation %q must say the outage is ONGOING, not repeat the entry wording", spoke[1])
	}
	if !strings.Contains(spoke[1], fmt.Sprintf("%d consecutive", alarmEscalateEvery)) {
		t.Errorf("escalation %q does not say how long it has been going", spoke[1])
	}
}

// 🛑 rc=11 IS THE WATCHER WORKING — another tick holds the lock. Counting it as
// failure makes a busy watcher alarm about its own health, and the operator
// learns to distrust the one signal that matters.
func TestLockHeldNeverAlarms(t *testing.T) {
	a, _ := fixedAlarm()
	for i := 0; i < 50; i++ {
		if line := a.observe(true, 11, nil); line != "" {
			t.Fatalf("rc=11 alarmed: %q", line)
		}
	}
}

// Recovery fires ONCE, on the transition. A repeated green is as useless as a
// stuck red.
func TestRecoveryIsAnnouncedOnceNotEveryHealthyTick(t *testing.T) {
	a, adv := fixedAlarm()
	a.observe(false, 12, boom)
	adv(time.Minute)

	if line := a.observe(true, 0, nil); !strings.Contains(line, "RECOVERED") {
		t.Fatalf("no recovery announced: %q", line)
	}
	for i := 0; i < 10; i++ {
		if line := a.observe(true, 0, nil); line != "" {
			t.Fatalf("healthy tick after recovery spoke again: %q", line)
		}
	}
}

// A second outage after a recovery must alarm again — the counter has to reset,
// or the alarm works exactly once per process lifetime.
func TestASecondOutageAlarmsAgain(t *testing.T) {
	a, adv := fixedAlarm()
	a.observe(false, 12, boom)
	adv(time.Minute)
	a.observe(true, 0, nil)
	adv(time.Minute)

	line := a.observe(false, 12, boom)
	if !strings.Contains(line, "FAILING") {
		t.Fatalf("second outage did not alarm: %q — the alarm would work once per process", line)
	}
	adv(time.Minute)
	if rec := a.observe(true, 0, nil); !strings.Contains(rec, "1 failed tick(s)") {
		t.Errorf("recovery %q did not reset the count from the previous outage", rec)
	}
}

// @Crispin's ALARM_AFTER (#365). Dev bounces several times an hour, and every
// bounce cost him a FAILING and a RECOVERED — ~6 turns in a day for outages
// that resolved themselves.
//
// 🛑 A blip BELOW the threshold must pass in COMPLETE silence, recovery
// included. An unannounced failure followed by an announced recovery is a line
// about nothing, and it is the failure mode that makes a threshold feel broken.
func TestBlipBelowThresholdIsEntirelySilent(t *testing.T) {
	a, adv := fixedAlarmAfter(3)
	var spoke []string
	say := func(l string) {
		if l != "" {
			spoke = append(spoke, l)
		}
	}
	say(a.observe(false, 12, boom)) // 1
	adv(time.Second)
	say(a.observe(false, 12, boom)) // 2 — still under
	adv(time.Second)
	say(a.observe(true, 0, nil)) // recovered before the threshold
	if len(spoke) != 0 {
		t.Fatalf("a 2-tick blip under --alarm-after 3 spoke %d time(s):\n%s", len(spoke), strings.Join(spoke, "\n"))
	}
}

// ...and reaching the threshold still alarms, naming how many ticks it took —
// otherwise the reader cannot tell a 1-tick outage from a 3-tick one.
func TestThresholdReachedAlarmsAndNamesTheCount(t *testing.T) {
	a, adv := fixedAlarmAfter(3)
	for i := 0; i < 2; i++ {
		if l := a.observe(false, 12, boom); l != "" {
			t.Fatalf("spoke at tick %d, below threshold: %q", i+1, l)
		}
		adv(time.Second)
	}
	line := a.observe(false, 12, boom)
	if line == "" {
		t.Fatal("silent at the threshold — the alarm can never fire")
	}
	if !strings.Contains(line, "3 consecutive") {
		t.Errorf("alarm %q does not say how many ticks it took", line)
	}
	adv(time.Second)
	if rec := a.observe(true, 0, nil); !strings.Contains(rec, "RECOVERED") {
		t.Errorf("no recovery after an ANNOUNCED failure: %q", rec)
	}
}

// ⚠️ The default must stay 1. My own shell watcher used 3 and the alarm NEVER
// FIRED in production — one non-zero tick ever, against a measured outage class
// of 9–29s. A threshold that is never reached is an alarm you stop being able
// to test.
func TestDefaultThresholdIsOneSoTheAlarmStaysObservable(t *testing.T) {
	if got := newRoomAlarm(0).alarmAfter; got != 1 {
		t.Errorf("newRoomAlarm(0).alarmAfter = %d, want 1", got)
	}
	a, _ := fixedAlarm()
	if a.observe(false, 12, boom) == "" {
		t.Fatal("default threshold did not alarm on the first failing tick")
	}
}
