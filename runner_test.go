package outrunner

import "testing"

func TestRunnerPhaseString(t *testing.T) {
	tests := []struct {
		phase RunnerPhase
		want  string
	}{
		{RunnerProvisioning, "provisioning"},
		{RunnerIdle, "idle"},
		{RunnerRunning, "running"},
		{RunnerStopping, "stopping"},
		{RunnerPhase(99), "unknown(99)"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := tt.phase.String()
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestSignalDoneIdempotent(t *testing.T) {
	state := &RunnerState{
		done: make(chan struct{}),
	}

	// First call should close the channel
	state.SignalDone()

	select {
	case <-state.done:
		// ok, channel is closed
	default:
		t.Fatal("done channel not closed after SignalDone")
	}

	// Second call should not panic
	state.SignalDone()
}

func TestSignalDoneUnblocks(t *testing.T) {
	state := &RunnerState{
		done: make(chan struct{}),
	}

	done := make(chan struct{})
	go func() {
		<-state.done
		close(done)
	}()

	state.SignalDone()

	<-done // should not block
}
