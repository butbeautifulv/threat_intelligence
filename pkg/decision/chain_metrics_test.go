package decision

import "testing"

func TestStepSuccessProbability(t *testing.T) {
	got := StepSuccessProbability(0.95, 0.8)
	want := 0.76
	if got != want {
		t.Fatalf("got %v want %v", got, want)
	}
}

func TestExecutionTimeEstimate_default(t *testing.T) {
	if ExecutionTimeEstimate("unknown-tool") != 180 {
		t.Fatal("expected default 180s")
	}
}
