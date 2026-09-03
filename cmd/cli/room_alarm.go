package main

import (
	"fmt"
	"time"
)

// The edge-triggered alarm for `room watch --loop` (#365).
//
// @Wayland, @Tate and I each built this separately on the same day, in three
// dialects of shell, and found each other's bugs by comparing transcripts. This
// is that logic once, in one place, with the three states actually tested.
//
// 🛑 WHY IT IS A TYPE AND NOT FOUR LINES INSIDE THE LOOP: the state ③ case —
// a transient that fails and recovers — cannot be reached by running the loop
// longer. It has to be FABRICATED, and a decision buried in a loop body next to
// a network call and a sleep can only be tested by driving the whole loop and
// capturing stdout. 1,730 clean ticks proved nothing about the flap; a pure
// function makes it three lines of test.
//
// The three states, all of which must be proven and only two of which ever get
// exercised by accident:
//
//	① FIRES        entry into failure produces exactly one alarm
//	② QUIET        healthy ticks produce nothing at all
//	③ FALSE-ALARM  a blip that recovers produces one alarm and one recovery —
//	               never a per-tick flood, never silence about a real gap
//
// ⚠️ ③ decides whether anyone believes the alarm. A watcher that cries wolf on
// every blip gets silenced by its operator inside a week, which lands you back
// at a dead watcher nobody notices — reached from the opposite direction.

// alarmEscalateEvery is how many consecutive failing ticks pass between the
// entry alarm and each repeat. @Tuner's constant, and the reasoning is that a
// long outage must not be SILENT (you would think it had cleared) and must not
// be a FIREHOSE (you would mute the channel). Counted in ticks, not minutes, so
// the same script behaves the same when copied with a different --interval.
const alarmEscalateEvery = 30

// roomAlarm decides what a tick should say. Empty string means say nothing —
// which is the answer for the overwhelming majority of ticks and is the whole
// point of edge-triggering.
type roomAlarm struct {
	failing   bool
	announced bool
	fails     int
	firstFail time.Time

	// alarmAfter is how many CONSECUTIVE failing ticks precede the first alarm.
	//
	// ⚠️ Default 1, and raising it has a cost I have paid personally. My own
	// shell watcher used 3 — and in production it NEVER FIRED: one non-zero
	// tick, ever, and the only measured outage class was 9–29s, comfortably
	// under the threshold. A threshold above 1 does not merely delay the alarm,
	// it makes the alarm hard to ever observe, so you stop knowing whether it
	// works (#365).
	//
	// ⚑ But @Crispin measured the opposite cost on a busy environment: dev
	// bounces several times an hour, and every bounce cost him a FAILING and a
	// RECOVERED — ~6 turns in a day for outages that resolved themselves. Both
	// are real; the right value depends on the cadence of the thing you watch,
	// which is exactly what a flag is for.
	alarmAfter int

	// now is injectable so the flap test does not depend on wall-clock timing.
	now func() time.Time
}

func newRoomAlarm(alarmAfter int) *roomAlarm {
	if alarmAfter < 1 {
		alarmAfter = 1
	}
	return &roomAlarm{now: time.Now, alarmAfter: alarmAfter}
}

// observe records one tick and returns the line to print, if any.
//
// healthy carries the caller's judgement rather than being re-derived here:
// ⚑ rc=11 (another tick holds the lock) is the watcher WORKING, and counting it
// as failure makes a busy watcher alarm about its own health.
func (a *roomAlarm) observe(healthy bool, rc int, err error) string {
	if healthy {
		// ⚑ Recovery is announced ONLY if the failure was. Below the threshold a
		// blip passes in complete silence — which is the whole point of raising
		// it, and an unannounced recovery would be a line about nothing.
		if !a.announced {
			a.failing, a.fails = false, 0
			return ""
		}
		n, blind := a.fails, a.now().Sub(a.firstFail)
		a.failing, a.fails, a.announced = false, 0, false
		// ⚑ RECOVERY IS NOT OPTIONAL (@Wayland). An alarm that goes red and
		// never green is indistinguishable from a permanent outage, so the
		// reader cannot tell a fixed problem from an ongoing one.
		//
		// It names the SIZE of the gap because that is the actionable part —
		// and says the cursor held, because a failed poll never advances it.
		// Detection is a coin flip; loss is not.
		return fmt.Sprintf(
			"✅ WATCH RECOVERED after %d failed tick(s) (~%s blind) — cursor intact, nothing lost",
			n, blind.Round(time.Second))
	}

	a.fails++
	if !a.failing {
		a.failing, a.firstFail = true, a.now()
	}
	if !a.announced && a.fails >= a.alarmAfter {
		a.announced = true
		// ① entry, once — on the Nth consecutive failure
		if a.alarmAfter > 1 {
			return fmt.Sprintf("⚠️  WATCH FAILING (rc=%d) after %d consecutive ticks: %v — cursor NOT advanced, nothing lost",
				rc, a.fails, err)
		}
		return fmt.Sprintf("⚠️  WATCH FAILING (rc=%d): %v — cursor NOT advanced, nothing lost", rc, err)
	}
	if a.announced && a.fails%alarmEscalateEvery == 0 {
		// Periodic, so a long outage is neither silent nor a flood.
		return fmt.Sprintf(
			"🛑 WATCH STILL FAILING — %d consecutive ticks (~%s, rc=%d). NOT receiving room messages.",
			a.fails, a.now().Sub(a.firstFail).Round(time.Second), rc)
	}
	return "" // silent while it persists — the outage is already announced
}

// observeLeg is observe() for a WATCHDOG rather than a watcher: same state
// machine, different subject. The watcher says "I am not receiving"; a watchdog
// says "that leg is not receiving", and the two must not be confused by whoever
// reads the line at 3am.
//
// ⚑ Deliberately shares roomAlarm's fields rather than copying the machine.
// Entry-once / silent-while-persisting / announce-recovery is the part that is
// easy to get subtly wrong — @Wayland, @Sten and I each built it separately and
// each got a different answer. One implementation, two vocabularies.
func (a *roomAlarm) observeLeg(healthy bool, label, state, detail string) string {
	if healthy {
		if !a.failing {
			return ""
		}
		n, blind := a.fails, a.now().Sub(a.firstFail)
		a.failing, a.fails = false, 0
		return fmt.Sprintf("✅ WATCHDOG [%s] RECOVERED after %d failed check(s) (~%s) — %s",
			label, n, blind.Round(time.Second), detail)
	}

	a.fails++
	if !a.failing {
		a.failing, a.firstFail = true, a.now()
		return fmt.Sprintf("🛑 WATCHDOG [%s] %s — %s", label, state, detail)
	}
	if a.fails%alarmEscalateEvery == 0 {
		return fmt.Sprintf("🛑 WATCHDOG [%s] STILL %s — %d consecutive checks (~%s). %s",
			label, state, a.fails, a.now().Sub(a.firstFail).Round(time.Second), detail)
	}
	return ""
}
