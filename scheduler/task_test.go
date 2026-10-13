package scheduler_test

import (
	"testing"
	"time"

	"github.com/matthiasharzer/livebuffer/scheduler"
	"github.com/stretchr/testify/assert"
)

func TestScheduleTask(t *testing.T) {
	t.Run("should execute task after specified duration", func(t *testing.T) {
		startedAt := time.Now()
		done := make(chan time.Time, 1)
		scheduler.In(250 * time.Millisecond).Run(func() error { done <- time.Now(); return nil })
		select {
		case executedAt := <-done:
			assert.WithinDuration(t, startedAt.Add(250*time.Millisecond), executedAt, 50*time.Millisecond)
		case <-time.After(time.Second):
			t.Fatal("task not executed")
		}
	})
	t.Run("should execute task at specified time", func(t *testing.T) {
		startedAt := time.Now()
		scheduler.At(startedAt.Add(time.Millisecond * 250)).Run(func() error {
			executedAt := time.Now()
			assert.WithinDuration(t, startedAt.Add(time.Millisecond*250), executedAt, time.Millisecond*50)
			return nil
		})
	})
}

func TestShutdown(t *testing.T) {
	t.Run("should cancel all unfinished tasks", func(t *testing.T) {
		scheduler.In(time.Millisecond * 150).Run(func() error {
			assert.Fail(t, "task should not have been executed")
			return nil
		})
		scheduler.Shutdown()
		time.Sleep(time.Millisecond * 500)
		// If the task was executed, the test would have failed. Since we are not asserting anything here, we just ensure that the test completes without any failures.
	})
}
