package services

import (
	"context"
	"log"
	"sync"

	"github.com/jackc/pgx/v5/pgtype"
	"golang.org/x/sync/errgroup"

	"github.com/BrAINLabs-Inc/mindflow-platform/backend-v2/internal/db/queries"
)

// Port of backend/src/controllers/journeyController.ts.
type JourneyService struct {
	q       queries.Querier
	daily   *DailyService
	weekly  *WeeklyService
	thrive  *ThriveService
	stress  *StressService
	mindful *MindfulService
}

func NewJourneyService(q queries.Querier, daily *DailyService, weekly *WeeklyService, thrive *ThriveService, stress *StressService, mindful *MindfulService) *JourneyService {
	return &JourneyService{q: q, daily: daily, weekly: weekly, thrive: thrive, stress: stress, mindful: mindful}
}

type JourneyStatus struct {
	Daily   any `json:"daily"`
	Weekly  any `json:"weekly"`
	Thrive  any `json:"thrive"`
	Stress  any `json:"stress"`
	Mindful any `json:"mindful"`
}

// statusFallback mirrors getJourneyStatus's Promise.allSettled fallback:
// one roadmap step's status check failing must return this shape for
// that key only, never fail the whole endpoint.
func statusFallback() map[string]any {
	return map[string]any{"completed": false, "nextReset": nil, "error": true}
}

// GetJourneyStatus fans out to all 5 roadmap status checks concurrently.
// A sync.WaitGroup with disjoint per-goroutine result slots is the Go
// equivalent of Promise.allSettled used here -- see plan §5.
func (s *JourneyService) GetJourneyStatus(ctx context.Context, userID pgtype.UUID) JourneyStatus {
	var wg sync.WaitGroup
	var daily, weekly, thrive, stress, mindful any

	wg.Add(5)
	go func() {
		defer wg.Done()
		v, err := s.daily.GetDailyStatus(ctx, userID)
		if err != nil {
			log.Printf("getJourneyStatus: daily failed: %v", err)
			daily = statusFallback()
			return
		}
		daily = v
	}()
	go func() {
		defer wg.Done()
		v, err := s.weekly.GetWeeklyStatus(ctx, userID)
		if err != nil {
			log.Printf("getJourneyStatus: weekly failed: %v", err)
			weekly = statusFallback()
			return
		}
		weekly = v
	}()
	go func() {
		defer wg.Done()
		v, err := s.thrive.GetThriveStatus(ctx, userID)
		if err != nil {
			log.Printf("getJourneyStatus: thrive failed: %v", err)
			thrive = statusFallback()
			return
		}
		thrive = v
	}()
	go func() {
		defer wg.Done()
		v, err := s.stress.GetStressStatus(ctx, userID)
		if err != nil {
			log.Printf("getJourneyStatus: stress failed: %v", err)
			stress = statusFallback()
			return
		}
		stress = v
	}()
	go func() {
		defer wg.Done()
		v, err := s.mindful.GetMindfulStatus(ctx, userID)
		if err != nil {
			log.Printf("getJourneyStatus: mindful failed: %v", err)
			mindful = statusFallback()
			return
		}
		mindful = v
	}()
	wg.Wait()

	return JourneyStatus{Daily: daily, Weekly: weekly, Thrive: thrive, Stress: stress, Mindful: mindful}
}

type JourneyData struct {
	Daily    []queries.ListJourneyDailyRow  `json:"daily"`
	Weekly   []queries.ListJourneyWeeklyRow `json:"weekly"`
	Research JourneyResearch                `json:"research"`
}

type JourneyResearch struct {
	PSS10    []queries.ListJourneyStressRow  `json:"pss10"`
	FFMQ15   []queries.ListJourneyMindfulRow `json:"ffmq15"`
	WEMWBS14 []queries.ListJourneyThriveRow  `json:"wemwbs14"`
}

// GetJourneyData returns full journey history for admin/web dashboard
// consumption, capped at limit rows per feature domain.
func (s *JourneyService) GetJourneyData(ctx context.Context, userID pgtype.UUID, limit int32) (JourneyData, error) {
	// Five independent reads, run concurrently like the TS source's
	// Promise.all; any failure fails the request.
	var (
		daily    []queries.ListJourneyDailyRow
		weekly   []queries.ListJourneyWeeklyRow
		pss10    []queries.ListJourneyStressRow
		ffmq15   []queries.ListJourneyMindfulRow
		wemwbs14 []queries.ListJourneyThriveRow
	)
	g, gctx := errgroup.WithContext(ctx)
	g.Go(func() (err error) {
		daily, err = s.q.ListJourneyDaily(gctx, queries.ListJourneyDailyParams{UserID: userID, Limit: limit})
		return err
	})
	g.Go(func() (err error) {
		weekly, err = s.q.ListJourneyWeekly(gctx, queries.ListJourneyWeeklyParams{UserID: userID, Limit: limit})
		return err
	})
	g.Go(func() (err error) {
		pss10, err = s.q.ListJourneyStress(gctx, queries.ListJourneyStressParams{UserID: userID, Limit: limit})
		return err
	})
	g.Go(func() (err error) {
		ffmq15, err = s.q.ListJourneyMindful(gctx, queries.ListJourneyMindfulParams{UserID: userID, Limit: limit})
		return err
	})
	g.Go(func() (err error) {
		wemwbs14, err = s.q.ListJourneyThrive(gctx, queries.ListJourneyThriveParams{UserID: userID, Limit: limit})
		return err
	})
	if err := g.Wait(); err != nil {
		return JourneyData{}, err
	}
	// sqlc returns a nil slice for zero rows, which json.Marshal renders
	// as `null`; the TS source always falls back to `[]` (`data || []`),
	// so empty results are coalesced the same way here.
	if daily == nil {
		daily = []queries.ListJourneyDailyRow{}
	}
	if weekly == nil {
		weekly = []queries.ListJourneyWeeklyRow{}
	}
	if pss10 == nil {
		pss10 = []queries.ListJourneyStressRow{}
	}
	if ffmq15 == nil {
		ffmq15 = []queries.ListJourneyMindfulRow{}
	}
	if wemwbs14 == nil {
		wemwbs14 = []queries.ListJourneyThriveRow{}
	}

	return JourneyData{
		Daily:  daily,
		Weekly: weekly,
		Research: JourneyResearch{
			PSS10:    pss10,
			FFMQ15:   ffmq15,
			WEMWBS14: wemwbs14,
		},
	}, nil
}
