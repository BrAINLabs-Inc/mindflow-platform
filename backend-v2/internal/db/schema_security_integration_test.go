package db_test

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/BrAINLabs-Inc/mindflow-platform/backend-v2/internal/testutil"
)

// These tests exercise database/project_db.sql's RLS policies and the
// research_id guard trigger the way a participant holding the public anon
// key would hit them through PostgREST: as the `authenticated` role with a
// JWT `sub` claim set. See FINDINGS_AND_DEVELOPMENT.md §3.

// asUser runs fn in a transaction as the `authenticated` role with
// auth.uid() = sub (pass uuid.Nil for "no JWT", i.e. service role /
// backend). The transaction is always rolled back.
func asUser(t *testing.T, pool *pgxpool.Pool, sub uuid.UUID, fn func(tx pgx.Tx) error) error {
	t.Helper()
	ctx := context.Background()
	tx, err := pool.Begin(ctx)
	if err != nil {
		t.Fatalf("begin: %v", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	if sub != uuid.Nil {
		if _, err := tx.Exec(ctx, `SELECT set_config('request.jwt.claim.sub', $1, true)`, sub.String()); err != nil {
			t.Fatalf("setting claim: %v", err)
		}
		if _, err := tx.Exec(ctx, `SET LOCAL ROLE authenticated`); err != nil {
			t.Fatalf("set role: %v", err)
		}
	}
	return fn(tx)
}

func sqlState(err error) string {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code
	}
	return ""
}

func seedUser(t *testing.T, pool *pgxpool.Pool) uuid.UUID {
	t.Helper()
	id := uuid.New()
	if _, err := pool.Exec(context.Background(), `INSERT INTO auth.users (id) VALUES ($1)`, id); err != nil {
		t.Fatalf("seeding user: %v", err)
	}
	return id
}

func TestResearchIDGuard(t *testing.T) {
	pool, _ := testutil.NewTestDB(t)
	ctx := context.Background()
	participant := seedUser(t, pool)
	admin := seedUser(t, pool)
	if _, err := pool.Exec(ctx, `INSERT INTO admins (id, username, email) VALUES ($1, 'admin', 'admin@example.com')`, admin); err != nil {
		t.Fatalf("seeding admin: %v", err)
	}
	if _, err := pool.Exec(ctx, `UPDATE profiles SET research_id = 'MF-001.cg' WHERE id = $1`, participant); err != nil {
		t.Fatalf("backend assigning research_id: %v", err)
	}

	t.Run("participant cannot change own study arm", func(t *testing.T) {
		err := asUser(t, pool, participant, func(tx pgx.Tx) error {
			_, err := tx.Exec(ctx, `UPDATE profiles SET research_id = 'MF-001.ex' WHERE id = $1`, participant)
			return err
		})
		if sqlState(err) != "42501" {
			t.Fatalf("err = %v, want insufficient_privilege (42501)", err)
		}
	})

	t.Run("participant can still update own username", func(t *testing.T) {
		err := asUser(t, pool, participant, func(tx pgx.Tx) error {
			tag, err := tx.Exec(ctx, `UPDATE profiles SET username = 'new_name' WHERE id = $1`, participant)
			if err == nil && tag.RowsAffected() != 1 {
				t.Errorf("rows affected = %d, want 1", tag.RowsAffected())
			}
			return err
		})
		if err != nil {
			t.Fatalf("username update: %v", err)
		}
	})

	t.Run("admin can reassign research_id", func(t *testing.T) {
		err := asUser(t, pool, admin, func(tx pgx.Tx) error {
			tag, err := tx.Exec(ctx, `UPDATE profiles SET research_id = 'MF-001.ex' WHERE id = $1`, participant)
			if err == nil && tag.RowsAffected() != 1 {
				t.Errorf("rows affected = %d, want 1", tag.RowsAffected())
			}
			return err
		})
		if err != nil {
			t.Fatalf("admin update: %v", err)
		}
	})

	t.Run("backend/service connection (no JWT) is unaffected", func(t *testing.T) {
		err := asUser(t, pool, uuid.Nil, func(tx pgx.Tx) error {
			_, err := tx.Exec(ctx, `UPDATE profiles SET research_id = NULL WHERE id = $1`, participant)
			return err
		})
		if err != nil {
			t.Fatalf("backend update: %v", err)
		}
	})
}

func TestParticipantsCannotWriteResearchDataDirectly(t *testing.T) {
	pool, _ := testutil.NewTestDB(t)
	ctx := context.Background()
	participant := seedUser(t, pool)

	writes := map[string]string{
		"daily_sliders insert":    `INSERT INTO daily_sliders (user_id, stress_level) VALUES ($1, 3)`,
		"voice_recordings insert": `INSERT INTO voice_recordings (user_id, week_number, year, file_key) VALUES ($1, 1, 2026, 'k')`,
		"pss10 insert": `INSERT INTO questionnaire_pss10_responses (user_id, q1,q2,q3,q4,q5,q6,q7,q8,q9,q10)
			VALUES ($1, 1,1,1,1,1,1,1,1,1,1)`,
		"about_me update": `UPDATE about_me_profiles SET is_completed = true WHERE id = $1`,
	}
	for name, stmt := range writes {
		t.Run(name, func(t *testing.T) {
			err := asUser(t, pool, participant, func(tx pgx.Tx) error {
				tag, err := tx.Exec(ctx, stmt, participant)
				// RLS rejects INSERTs with an error, but silently filters
				// UPDATE targets to zero rows.
				if err == nil && tag.RowsAffected() != 0 {
					t.Errorf("rows affected = %d, want 0", tag.RowsAffected())
				}
				return err
			})
			if err != nil && sqlState(err) != "42501" {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}

	t.Run("participant can still read own rows", func(t *testing.T) {
		if _, err := pool.Exec(ctx, `INSERT INTO daily_sliders (user_id, stress_level) VALUES ($1, 3)`, participant); err != nil {
			t.Fatalf("backend insert: %v", err)
		}
		err := asUser(t, pool, participant, func(tx pgx.Tx) error {
			var n int
			if err := tx.QueryRow(ctx, `SELECT COUNT(*) FROM daily_sliders`).Scan(&n); err != nil {
				return err
			}
			if n != 1 {
				t.Errorf("visible rows = %d, want 1", n)
			}
			return nil
		})
		if err != nil {
			t.Fatalf("select: %v", err)
		}
	})
}

// TestSecurityHardeningPatchIsIdempotent applies the patch for existing
// databases on top of the already-hardened schema, twice.
func TestSecurityHardeningPatchIsIdempotent(t *testing.T) {
	pool, _ := testutil.NewTestDB(t)
	_, thisFile, _, _ := runtime.Caller(0)
	patch, err := os.ReadFile(filepath.Join(filepath.Dir(thisFile), "..", "..", "..", "database", "patches", "2026-11-18-security-hardening.sql"))
	if err != nil {
		t.Fatalf("reading patch: %v", err)
	}
	for i := 1; i <= 2; i++ {
		if _, err := pool.Exec(context.Background(), string(patch)); err != nil {
			t.Fatalf("applying patch (run %d): %v", i, err)
		}
	}
}
