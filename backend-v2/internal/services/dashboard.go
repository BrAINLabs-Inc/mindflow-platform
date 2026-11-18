package services

import (
	"context"
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"golang.org/x/sync/errgroup"

	"github.com/BrAINLabs-Inc/mindflow-platform/backend-v2/internal/db/queries"
	"github.com/BrAINLabs-Inc/mindflow-platform/backend-v2/internal/researchgroup"
	"github.com/BrAINLabs-Inc/mindflow-platform/backend-v2/internal/timeutil"
)

// Port of backend/src/services/dashboardService.ts.
type DashboardService struct {
	q queries.Querier
}

func NewDashboardService(q queries.Querier) *DashboardService {
	return &DashboardService{q: q}
}

type DashboardStatus struct {
	DailyDone   bool `json:"dailyDone"`
	WeeklyDone  bool `json:"weeklyDone"`
	MonthlyDone bool `json:"monthlyDone"`
}

type DashboardSummary struct {
	Streak         int                 `json:"streak"`
	Consistency    int                 `json:"consistency"`
	WeeklyProgress int                 `json:"weeklyProgress"`
	TotalCompleted int64               `json:"totalCompleted"`
	Status         DashboardStatus     `json:"status"`
	Group          researchgroup.Group `json:"group"`
}

// GetUserSummary is a deliberate, faithful port of dashboardService.ts
// INCLUDING its timezone inconsistency: unlike every other feature, the
// streak/consistency math here uses the server process's local time
// (time.Now() / time.Local), NOT the Sri-Lanka-pinned timeutil helpers
// used everywhere else. This looks like a bug and may be one, but
// silently "fixing" it during the rewrite would shift real participants'
// streak numbers with no product sign-off -- see plan §5. weeklyProgress
// below DOES use the SLT-pinned ISO week helper, exactly like the TS
// source -- that split is intentional, not an oversight.
func (s *DashboardService) GetUserSummary(ctx context.Context, userID pgtype.UUID) (DashboardSummary, error) {
	today := time.Now()
	sixMonthsAgo := today.AddDate(0, -6, 0)
	thirtyDaysAgo := today.AddDate(0, 0, -30)
	startOfToday := time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, today.Location())
	startOfMonth := time.Date(today.Year(), today.Month(), 1, 0, 0, 0, 0, today.Location())
	streakSince := today.AddDate(0, 0, -StreakLookbackDays)

	// The eight reads are independent, so they run concurrently (as
	// Promise.all does in the TS source); any failure fails the request.
	var (
		streakData                             []pgtype.Timestamptz
		voiceRecordings                        []pgtype.Timestamptz
		totalCount, recentCount                int64
		pss10Count, ffmq15Count, wemwbs14Count int64
		researchID                             pgtype.Text
	)
	since := func(t time.Time) pgtype.Timestamptz { return pgtype.Timestamptz{Time: t, Valid: true} }
	g, gctx := errgroup.WithContext(ctx)
	g.Go(func() (err error) {
		streakData, err = s.q.ListDailyEntriesSince(gctx, queries.ListDailyEntriesSinceParams{UserID: userID, CreatedAt: since(streakSince)})
		return err
	})
	g.Go(func() (err error) {
		totalCount, err = s.q.CountDailyEntriesSince(gctx, queries.CountDailyEntriesSinceParams{UserID: userID, CreatedAt: since(sixMonthsAgo)})
		return err
	})
	g.Go(func() (err error) {
		recentCount, err = s.q.CountDailyEntriesSince(gctx, queries.CountDailyEntriesSinceParams{UserID: userID, CreatedAt: since(thirtyDaysAgo)})
		return err
	})
	g.Go(func() (err error) {
		voiceRecordings, err = s.q.ListVoiceRecordingsSince(gctx, queries.ListVoiceRecordingsSinceParams{UserID: userID, CreatedAt: since(sixMonthsAgo)})
		return err
	})
	g.Go(func() (err error) {
		pss10Count, err = s.q.CountStressResponsesSince(gctx, queries.CountStressResponsesSinceParams{UserID: userID, CreatedAt: since(startOfMonth)})
		return err
	})
	g.Go(func() (err error) {
		ffmq15Count, err = s.q.CountMindfulResponsesSince(gctx, queries.CountMindfulResponsesSinceParams{UserID: userID, CreatedAt: since(startOfMonth)})
		return err
	})
	g.Go(func() (err error) {
		wemwbs14Count, err = s.q.CountThriveResponsesSince(gctx, queries.CountThriveResponsesSinceParams{UserID: userID, CreatedAt: since(startOfMonth)})
		return err
	})
	g.Go(func() error {
		var err error
		researchID, err = s.q.GetResearchID(gctx, userID)
		if errors.Is(err, pgx.ErrNoRows) {
			return nil
		}
		return err
	})
	if err := g.Wait(); err != nil {
		return DashboardSummary{}, err
	}
	var group researchgroup.Group
	if researchID.Valid {
		group = researchgroup.Derive(&researchID.String)
	}

	// Streak calculation.
	var currentStreak int
	var isDailyDone bool
	if len(streakData) > 0 {
		lastEntry := streakData[0].Time // ListDailyEntriesSince orders DESC
		isDailyDone = !lastEntry.Before(startOfToday)

		entryDates := make(map[int64]struct{}, len(streakData))
		for _, e := range streakData {
			local := e.Time.In(today.Location())
			d := time.Date(local.Year(), local.Month(), local.Day(), 0, 0, 0, 0, today.Location())
			entryDates[d.Unix()] = struct{}{}
		}

		checkDate := startOfToday
		if !isDailyDone {
			checkDate = checkDate.AddDate(0, 0, -1)
		}
		for {
			if _, ok := entryDates[checkDate.Unix()]; !ok {
				break
			}
			currentStreak++
			checkDate = checkDate.AddDate(0, 0, -1)
		}
	}

	// Consistency: % of last 30 days with entries.
	consistency := 0
	if recentCount > 0 {
		consistency = min(100, int(math.Round(float64(recentCount)/30*100)))
	}

	// Weekly progress: unique ISO weeks with voice recordings out of 26.
	uniqueWeeks := make(map[string]struct{})
	for _, r := range voiceRecordings {
		year, week := timeutil.GetISOWeekNumber(r.Time)
		uniqueWeeks[isoWeekKey(year, week)] = struct{}{}
	}
	weeklyProgress := 0
	if len(uniqueWeeks) > 0 {
		weeklyProgress = min(100, int(math.Round(float64(len(uniqueWeeks))/26*100)))
	}

	currentYear, currentWeek := timeutil.GetISOWeekNumber(time.Now())
	_, isWeeklyDone := uniqueWeeks[isoWeekKey(currentYear, currentWeek)]

	isMonthlyDone := (pss10Count + ffmq15Count + wemwbs14Count) > 0

	return DashboardSummary{
		Streak:         currentStreak,
		Consistency:    consistency,
		WeeklyProgress: weeklyProgress,
		TotalCompleted: totalCount,
		Status: DashboardStatus{
			DailyDone:   isDailyDone,
			WeeklyDone:  isWeeklyDone,
			MonthlyDone: isMonthlyDone,
		},
		Group: group,
	}, nil
}

func isoWeekKey(year, week int) string {
	return fmt.Sprintf("%d-W%02d", year, week)
}
