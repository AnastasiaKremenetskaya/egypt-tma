package game

import (
	"math/rand"

	"github.com/anterekhova/egypt-tma/internal/questions"
)

// NextTurn advances the active player index and resets per-turn state.
func (r *Room) NextTurn() {
	r.ActiveIdx = (r.ActiveIdx + 1) % len(r.Players)
	if r.ActiveIdx == 0 {
		r.Round++
	}
	r.CurrentQuestion = nil
	r.CurrentAnswer = nil
	r.ResetVotes()
}

// DrawCard picks a random Maat or Seth question (50/50) and stores it.
func (r *Room) DrawCard(bank *questions.Bank) *QuestionRef {
	if rand.Intn(2) == 0 && len(bank.Maat) > 0 {
		q := bank.Maat[rand.Intn(len(bank.Maat))]
		ref := &QuestionRef{ID: q.ID, Type: "maat", Text: q.Text}
		r.CurrentQuestion = ref
		return ref
	}
	return r.DrawSethCard(bank)
}

// DrawSethCard picks a Seth question and builds shuffled options.
func (r *Room) DrawSethCard(bank *questions.Bank) *QuestionRef {
	if len(bank.Seth) == 0 {
		return nil
	}
	q := bank.Seth[rand.Intn(len(bank.Seth))]

	// collect up to 3 distractors from other Seth answers
	distractors := make([]string, 0, 3)
	seen := map[string]bool{q.Answer: true}
	shuffled := rand.Perm(len(bank.Seth))
	for _, i := range shuffled {
		a := bank.Seth[i].Answer
		if !seen[a] {
			distractors = append(distractors, a)
			seen[a] = true
		}
		if len(distractors) == 3 {
			break
		}
	}

	// build options: correct + distractors, then shuffle
	opts := append([]string{q.Answer}, distractors...)
	rand.Shuffle(len(opts), func(i, j int) { opts[i], opts[j] = opts[j], opts[i] })

	correctIdx := 0
	for i, o := range opts {
		if o == q.Answer {
			correctIdx = i
			break
		}
	}

	ref := &QuestionRef{
		ID:            q.ID,
		Type:          "seth",
		Text:          q.Text,
		Answer:        q.Answer,
		Options:       opts,
		CorrectOptIdx: correctIdx,
	}
	r.CurrentQuestion = ref
	return ref
}
