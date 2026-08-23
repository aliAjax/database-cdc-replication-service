package failoverstate

func Active(states map[string]State) []string {
	out := make([]string, 0, len(states))
	for id, state := range states {
		if state == StateStreaming || state == StateDegraded || state == StateRecovering {
			out = append(out, id)
		}
	}
	return out
}
