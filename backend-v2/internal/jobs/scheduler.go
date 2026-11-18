// Package jobs runs backend-v2's background cron schedule.
package jobs

import (
	"context"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/robfig/cron/v3"

	"github.com/BrAINLabs-Inc/mindflow-platform/backend-v2/internal/services"
)

// studyTimezone is fixed to the study cohort's local time (Sri Lanka),
// not the deploy host's ambient timezone -- a host provisioned with
// TZ=UTC (the common default) would otherwise fire these hours off from
// the intended 8am/7pm local. 8am morning greeting, 7pm pending-task
// nudge -- see reminderScheduler.ts. cmd/server embeds time/tzdata so
// this resolves even on images without a system tz database.
const studyTimezone = "Asia/Colombo"

// Advisory-lock keys, one per job, so that when backend-v2 runs as
// several replicas only one of them sends each reminder batch -- every
// replica's cron fires at the same instant, and without this each
// participant would get one notification per replica.
const (
	morningGreetingLockKey int64 = 0x4d46_0001 // "MF" + job id
	pendingReminderLockKey int64 = 0x4d46_0002
)

// StartReminderScheduler registers the two Asia/Colombo-pinned cron jobs
// and starts the scheduler in the background. Returns the cron.Cron so
// main can Stop() it during graceful shutdown.
func StartReminderScheduler(notificationSvc *services.NotificationService, pool *pgxpool.Pool) (*cron.Cron, error) {
	loc, err := time.LoadLocation(studyTimezone)
	if err != nil {
		return nil, err
	}
	c := cron.New(cron.WithLocation(loc))
	if _, err := c.AddFunc("0 8 * * *", func() {
		runExclusive(pool, morningGreetingLockKey, "Morning greeting", func(ctx context.Context) {
			sent, err := notificationSvc.SendMorningGreetings(ctx)
			if err != nil {
				log.Printf("[Reminders] Morning greeting job failed: %v", err)
				return
			}
			log.Printf("[Reminders] Morning greetings sent: %d", sent)
		})
	}); err != nil {
		return nil, err
	}
	if _, err := c.AddFunc("0 19 * * *", func() {
		runExclusive(pool, pendingReminderLockKey, "Pending-task reminder", func(ctx context.Context) {
			sent, err := notificationSvc.SendPendingTaskReminders(ctx)
			if err != nil {
				log.Printf("[Reminders] Pending-task reminder job failed: %v", err)
				return
			}
			log.Printf("[Reminders] Pending-task reminders sent: %d", sent)
		})
	}); err != nil {
		return nil, err
	}
	c.Start()
	return c, nil
}

// runExclusive runs job only if this instance wins a transaction-scoped
// Postgres advisory lock for key. A transaction-scoped lock (rather than a
// session lock) is used because DATABASE_URL may point at Supavisor in
// transaction mode, where a session lock could be released or leaked as
// the pooler hands the backend connection to someone else; an open
// transaction pins one backend connection until it ends. The lock is held
// for the job's duration and released on commit/rollback.
func runExclusive(pool *pgxpool.Pool, key int64, name string, job func(context.Context)) {
	ctx := context.Background()
	tx, err := pool.Begin(ctx)
	if err != nil {
		log.Printf("[Reminders] %s job skipped: could not begin lock transaction: %v", name, err)
		return
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var acquired bool
	if err := tx.QueryRow(ctx, "SELECT pg_try_advisory_xact_lock($1)", key).Scan(&acquired); err != nil {
		log.Printf("[Reminders] %s job skipped: lock query failed: %v", name, err)
		return
	}
	if !acquired {
		log.Printf("[Reminders] %s job skipped: another instance is running it", name)
		return
	}
	job(ctx)
}
