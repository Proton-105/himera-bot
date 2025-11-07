package state

// validTransitions contains the permitted non-emergency transitions in the FSM.
var validTransitions = map[State][]State{
	StateIdle: {
		StateBuyingSearch,
	},
	StateBuyingSearch: {
		StateBuyingAmount,
		StateIdle,
	},
	StateBuyingAmount: {
		StateBuyingConfirm,
		StateBuyingSearch,
	},
	StateBuyingConfirm: {
		StateIdle,
	},
}

// IsTransitionAllowed reports whether moving from one state to another is valid.
func IsTransitionAllowed(from, to State) bool {
	if to == StateError || to == StateIdle {
		return true
	}

	allowed, ok := validTransitions[from]
	if !ok {
		return false
	}

	for _, state := range allowed {
		if state == to {
			return true
		}
	}

	return false
}
