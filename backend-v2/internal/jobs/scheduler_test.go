package jobs

import (
	"context"
	"testing"

	"github.com/BrAINLabs-Inc/mindflow-platform/backend-v2/internal/testutil"
)

// TestRunExclusive simulates two replicas firing the same cron tick: while
// one holds the job's advisory lock, the other must skip rather than send
// a duplicate batch; once it is released, the next run proceeds.
func TestRunExclusive(t *testing.T) {
	pool, _ := testutil.NewTestDB(t)
	ctx := context.Background()

	ran := 0
	job := func(context.Context) { ran++ }

	runExclusive(pool, morningGreetingLockKey, "test", job)
	if ran != 1 {
		t.Fatalf("uncontended run: ran = %d, want 1", ran)
	}

	// Another "instance" is mid-job holding the lock.
	holder, err := pool.Begin(ctx)
	if err != nil {
		t.Fatalf("begin: %v", err)
	}
	if _, err := holder.Exec(ctx, "SELECT pg_advisory_xact_lock($1)", morningGreetingLockKey); err != nil {
		t.Fatalf("taking lock: %v", err)
	}
	runExclusive(pool, morningGreetingLockKey, "test", job)
	if ran != 1 {
		t.Fatalf("contended run: ran = %d, want still 1 (must skip)", ran)
	}
	// A different job's lock is independent.
	runExclusive(pool, pendingReminderLockKey, "test", job)
	if ran != 2 {
		t.Fatalf("other job: ran = %d, want 2", ran)
	}

	if err := holder.Rollback(ctx); err != nil {
		t.Fatalf("release: %v", err)
	}
	runExclusive(pool, morningGreetingLockKey, "test", job)
	if ran != 3 {
		t.Fatalf("after release: ran = %d, want 3", ran)
	}
}
