package plan

import "time"
// BuiltinAnchors are default clock targets (HH:MM) for Entrada1 → Saída1 → Entrada2 → Saída2.
// Legacy shorthand "0008" means **08:00**, not midnight 00:08.
func BuiltinAnchors() [4]string {
	return [4]string{"08:00", "12:00", "13:30", "18:00"}
}

func absInt(value int) int {
	if value < 0 {
		return -value
	}
	return value
}

// parseHHMMToMinutes parses a "15:04" string and returns hours*60+minutes.
// Returns 0 for invalid input.
func parseHHMMToMinutes(hhmm string) int {
	parsed, err := time.Parse("15:04", hhmm)
	if err != nil {
		return 0
	}
	return parsed.Hour()*60 + parsed.Minute()
}

// assignStampsToPlannerSlots uses order-preserving DP to assign all stamps to the
// minimum-cost subset of slots. All stamps must be assigned (slotCount >= len(stamps)).
// Returns a []string of length len(anchors): each element is the assigned stamp time or "".
func assignStampsToPlannerSlots(stamps []string, anchors []string) []string {
	numSlots := len(anchors)
	result := make([]string, numSlots)

	if len(stamps) == 0 || numSlots == 0 {
		return result
	}

	numStamps := len(stamps)

	stampMinutes := make([]int, numStamps)
	for stampIndex, stamp := range stamps {
		stampMinutes[stampIndex] = parseHHMMToMinutes(stamp)
	}

	anchorMinutes := make([]int, numSlots)
	for slotIndex, anchor := range anchors {
		anchorMinutes[slotIndex] = parseHHMMToMinutes(anchor)
	}

	const maxCost = int(^uint(0) >> 1)

	// dp[i][j] = min cost to assign stamps[0..i-1] to i slots chosen from anchors[0..j-1].
	// All i stamps must be assigned; slots may be skipped (left empty).
	dp := make([][]int, numStamps+1)
	for rowIndex := range dp {
		dp[rowIndex] = make([]int, numSlots+1)
		for colIndex := range dp[rowIndex] {
			dp[rowIndex][colIndex] = maxCost
		}
	}
	for colIndex := 0; colIndex <= numSlots; colIndex++ {
		dp[0][colIndex] = 0
	}

	for stampIndex := 1; stampIndex <= numStamps; stampIndex++ {
		for slotIndex := 1; slotIndex <= numSlots; slotIndex++ {
			// Option 1: skip slot slotIndex-1 (leave it empty).
			skipSlotCost := dp[stampIndex][slotIndex-1]

			// Option 2: assign stamp stampIndex-1 to slot slotIndex-1.
			assignCost := maxCost
			if dp[stampIndex-1][slotIndex-1] < maxCost {
				assignCost = dp[stampIndex-1][slotIndex-1] + absInt(stampMinutes[stampIndex-1]-anchorMinutes[slotIndex-1])
			}

			if assignCost < skipSlotCost {
				dp[stampIndex][slotIndex] = assignCost
			} else {
				dp[stampIndex][slotIndex] = skipSlotCost
			}
		}
	}

	// Backtrack to reconstruct the assignment.
	stampIndex := numStamps
	slotIndex := numSlots
	for stampIndex > 0 && slotIndex > 0 {
		assignCost := maxCost
		if dp[stampIndex-1][slotIndex-1] < maxCost {
			assignCost = dp[stampIndex-1][slotIndex-1] + absInt(stampMinutes[stampIndex-1]-anchorMinutes[slotIndex-1])
		}

		if dp[stampIndex][slotIndex] == assignCost {
			result[slotIndex-1] = stamps[stampIndex-1]
			stampIndex--
			slotIndex--
			continue
		}
		slotIndex--
	}

	return result
}
