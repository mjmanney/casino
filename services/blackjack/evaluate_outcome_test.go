package blackjack

import "testing"

func TestEvaluateOutcome_PlayerWins(t *testing.T) {
	got := EvaluateOutcome(20, Qualified, 19, Qualified)
	want := Win
	if got != want {
		t.Errorf("got %v, want %v", got, want)
	}
	got = EvaluateOutcome(20, Qualified, 22, Busted)
	want = Win
	if got != want {
		t.Errorf("got %v, want %v", got, want)
	}

	got = EvaluateOutcome(21, Blackjack, 21, Qualified)
	want = Win
	if got != want {
		t.Errorf("got %v, want %v", got, want)
	}

	got = EvaluateOutcome(18, Qualified, 22, Qualified)
	want = Win
	if got != want {
		t.Errorf("got %v, want %v", got, want)
	}

}

func TestEvaluateOutcome_PlayerLoses(t *testing.T) {
	got := EvaluateOutcome(18, Qualified, 19, Qualified)
	want := Loss
	if got != want {
		t.Errorf("got %v, want %v", got, want)
	}

	got = EvaluateOutcome(22, Busted, 22, Busted)
	want = Loss
	if got != want {
		t.Errorf("got %v, want %v", got, want)
	}

	got = EvaluateOutcome(21, Qualified, 21, Blackjack)
	want = Loss
	if got != want {
		t.Errorf("got %v, want %v", got, want)
	}

	got = EvaluateOutcome(20, Surrendered, 22, Busted)
	want = Loss
	if got != want {
		t.Errorf("got %v, want %v", got, want)
	}

	got = EvaluateOutcome(22, Qualified, 18, Qualified)
	want = Loss
	if got != want {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestEvaluateOutcome_PlayerPushes(t *testing.T) {
	got := EvaluateOutcome(21, Qualified, 21, Qualified)
	want := Push
	if got != want {
		t.Errorf("got %v, want %v", got, want)
	}

	got = EvaluateOutcome(21, Blackjack, 21, Blackjack)
	want = Push
	if got != want {
		t.Errorf("got %v, want %v", got, want)
	}
}
