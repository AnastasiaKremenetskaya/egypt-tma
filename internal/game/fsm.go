package game

import "fmt"

// Transition validates and applies a phase change.
func (r *Room) Transition(to Phase) error {
	if !validTransition(r.Phase, to) {
		return fmt.Errorf("invalid transition %s → %s", r.Phase, to)
	}
	r.StopTimer()
	r.Phase = to
	return nil
}

func validTransition(from, to Phase) bool {
	allowed := map[Phase][]Phase{
		PhaseLobby:    {PhaseQuestion},
		PhaseQuestion: {PhaseVoting, PhaseQuestion}, // voice answer skips to voting; timeout loops
		PhaseVoting:   {PhaseSeth, PhaseQuestion, PhaseFinished},
		PhaseSeth:     {PhaseQuestion, PhaseFinished},
		PhaseFinished: {},
	}
	for _, p := range allowed[from] {
		if p == to {
			return true
		}
	}
	return false
}
