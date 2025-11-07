package state

import "testing"

func TestIsTransitionAllowed(t *testing.T) {
	testCases := []struct {
		name     string
		from     State
		to       State
		expected bool
	}{
		{name: "idle to buying search", from: StateIdle, to: StateBuyingSearch, expected: true},
		{name: "buying search to buying amount", from: StateBuyingSearch, to: StateBuyingAmount, expected: true},
		{name: "buying search back to idle", from: StateBuyingSearch, to: StateIdle, expected: true},
		{name: "buying amount to buying confirm", from: StateBuyingAmount, to: StateBuyingConfirm, expected: true},
		{name: "buying amount to buying search", from: StateBuyingAmount, to: StateBuyingSearch, expected: true},
		{name: "buying confirm to idle", from: StateBuyingConfirm, to: StateIdle, expected: true},
		{name: "idle to buying confirm invalid", from: StateIdle, to: StateBuyingConfirm, expected: false},
		{name: "buying confirm to buying amount invalid", from: StateBuyingConfirm, to: StateBuyingAmount, expected: false},
		{name: "unknown state to buying search invalid", from: State("unknown"), to: StateBuyingSearch, expected: false},
		{name: "any state to idle emergency", from: State("whatever"), to: StateIdle, expected: true},
		{name: "any state to error emergency", from: StateBuyingConfirm, to: StateError, expected: true},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			if actual := IsTransitionAllowed(tc.from, tc.to); actual != tc.expected {
				t.Errorf("IsTransitionAllowed(%s -> %s) = %t, expected %t", tc.from, tc.to, actual, tc.expected)
			}
		})
	}
}
