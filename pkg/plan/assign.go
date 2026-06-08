package plan

import "time"

// BuiltinAnchors are default clock targets (HH:MM) for Entrada1 → Saída1 → Entrada2 → Saída2.
// Legacy shorthand "0008" means **08:00**, not midnight 00:08.
func BuiltinAnchors() [4]string {
	return [4]string{"08:00", "12:00", "13:30", "18:00"}
}

func absDur(d time.Duration) time.Duration {
	if d < 0 {
		return -d
	}
	return d
}

func anchorCost(stamp, anchor time.Time) time.Duration {
	return absDur(stamp.Sub(anchor))
}

// AssignStampsToPlannerSlots maps chronological stamps to the four planner slots.
// Returned slice has length 4: each slot is the stamp index in sorted stamps, or -1.
func AssignStampsToPlannerSlots(sorted []time.Time, date time.Time, anchorsHM [4]string) [4]int {
	return assignStampsToPlannerSlots(sorted, date, anchorsHM)
}

// anchorCombinations returns all C(4,k) increasing index tuples drawn from {0,1,2,3}.
func anchorCombinations(k int) [][]int {
	if k <= 0 || k > 4 {
		return nil
	}
	var out [][]int
	var build func(start int, cur []int)
	build = func(start int, cur []int) {
		if len(cur) == k {
			cp := make([]int, k)
			copy(cp, cur)
			out = append(out, cp)
			return
		}
		for i := start; i < 4; i++ {
			cur = append(cur, i)
			build(i+1, cur)
			cur = cur[:len(cur)-1]
		}
	}
	build(0, nil)
	return out
}

// assignStampsToPlannerSlots selects up to four increasing stamp indices and maps each to one
// of four planner anchors (order-preserving), minimizing summed |stamp−anchor|.
// Returned slice has length 4: each slot is the stamp index in sorted chronological stamps, or -1 if none.
func assignStampsToPlannerSlots(sorted []time.Time, date time.Time, anchorsHM [4]string) [4]int {
	fail := func() [4]int {
		var r [4]int
		for i := range r {
			r[i] = -1
		}
		return r
	}

	n := len(sorted)
	var anchors [4]time.Time
	for i, hhmm := range anchorsHM {
		at, err := parseClock(hhmm, date)
		if err != nil {
			return fail()
		}
		anchors[i] = at
	}

	var best [4]int
	for i := range best {
		best[i] = -1
	}
	found := false
	var bestCost time.Duration

	improveIfBetter := func(anchorIx []int, stampIx []int) {
		cost := time.Duration(0)
		for j := range anchorIx {
			cost += anchorCost(sorted[stampIx[j]], anchors[anchorIx[j]])
		}
		if !found || cost < bestCost {
			found = true
			bestCost = cost
			for i := range best {
				best[i] = -1
			}
			for j := range anchorIx {
				best[anchorIx[j]] = stampIx[j]
			}
		}
	}

	if n >= 4 {
		combos := anchorCombinations(4)
		if len(combos) == 0 || len(combos[0]) != 4 {
			return fail()
		}
		anc := combos[0]
		var walk func(startIx int, picks []int)
		walk = func(startIx int, picks []int) {
			if len(picks) == 4 {
				improveIfBetter(anc, picks)
				return
			}
			need := 4 - len(picks)
			for ix := startIx; ix <= n-need; ix++ {
				walk(ix+1, append(picks, ix))
			}
		}
		walk(0, nil)
	} else {
		k := n
		for _, anc := range anchorCombinations(k) {
			var walk func(startIx int, picks []int)
			walk = func(startIx int, picks []int) {
				if len(picks) == k {
					improveIfBetter(anc, picks)
					return
				}
				need := k - len(picks)
				for ix := startIx; ix <= n-need; ix++ {
					walk(ix+1, append(picks, ix))
				}
			}
			walk(0, nil)
		}
	}

	used := make([]bool, n)
	for _, ix := range best {
		if ix >= 0 {
			used[ix] = true
		}
	}
	var loose []int
	for i := 0; i < n; i++ {
		if !used[i] {
			loose = append(loose, i)
		}
	}
	li := 0
	for si := 0; si < 4 && li < len(loose); si++ {
		if best[si] >= 0 {
			continue
		}
		best[si] = loose[li]
		li++
	}

	return best
}
