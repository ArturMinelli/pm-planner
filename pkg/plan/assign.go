package plan

// BuiltinAnchors are default clock targets (HH:MM) for Entrada1 → Saída1 → Entrada2 → Saída2.
// Legacy shorthand "0008" means **08:00**, not midnight 00:08.
func BuiltinAnchors() [4]string {
	return [4]string{"08:00", "12:00", "13:30", "18:00"}
}

// assignStampsToPlannerSlots fills slots in punch order: stamp i goes to slot i.
// Extra slots stay empty. Extra stamps beyond len(anchors) are ignored.
func assignStampsToPlannerSlots(stamps []string, anchors []string) []string {
	result := make([]string, len(anchors))
	limit := len(stamps)
	if limit > len(anchors) {
		limit = len(anchors)
	}
	copy(result, stamps[:limit])
	return result
}
