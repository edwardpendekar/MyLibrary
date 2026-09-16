//go:build integration

// Package integration runs the service layer against real Postgres and Redis
// containers (via testcontainers-go) instead of fakes. It exists specifically
// for the two paths that are too infrastructure-heavy to trust to in-memory
// fakes: the Excel/CSV import pipeline (raw pgx COPY + set-based merge) and
// full-text search (generated tsvector columns + GIN indexes only Postgres
// itself can execute). Everything else stays covered by the fast, fake-backed
// unit tests in internal/service.
//
// Run with: go test -tags=integration ./test/integration/...
// Requires a working Docker daemon; there is no other way to skip these other
// than not passing the build tag, since testcontainers has no fake mode.
package integration

import (
	"context"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/testcontainers/testcontainers-go"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	tcredis "github.com/testcontainers/testcontainers-go/modules/redis"
	"github.com/testcontainers/testcontainers-go/wait"
	gormpg "gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"

	"bookreader/backend/pkg/cache"
)

// testEnv bundles the same three handles main.go wires up in production
// (GORM for CRUD repos, a raw pgx pool for COPY/FTS, and Redis for caching),
// pointed at throwaway per-test containers.
type testEnv struct {
	DB    *gorm.DB
	Pool  *pgxpool.Pool
	Cache *cache.Cache
}

func setupTestEnv(t *testing.T) *testEnv {
	t.Helper()
	ctx := context.Background()

	pgContainer, err := tcpostgres.Run(ctx, "postgres:17-alpine",
		tcpostgres.WithDatabase("bookreader_test"),
		tcpostgres.WithUsername("postgres"),
		tcpostgres.WithPassword("postgres"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(60*time.Second)),
	)
	if err != nil {
		t.Fatalf("failed to start postgres container: %v", err)
	}
	t.Cleanup(func() {
		if err := testcontainers.TerminateContainer(pgContainer); err != nil {
			t.Logf("failed to terminate postgres container: %v", err)
		}
	})

	dsn, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Fatalf("failed to get postgres connection string: %v", err)
	}

	applyMigrations(t, ctx, dsn)

	db, err := gorm.Open(gormpg.Open(dsn), &gorm.Config{
		Logger:                                   gormlogger.Default.LogMode(gormlogger.Silent),
		DisableForeignKeyConstraintWhenMigrating: true,
	})
	if err != nil {
		t.Fatalf("failed to open gorm connection: %v", err)
	}
	t.Cleanup(func() {
		if sqlDB, err := db.DB(); err == nil {
			_ = sqlDB.Close()
		}
	})

	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Fatalf("failed to create pgx pool: %v", err)
	}
	t.Cleanup(pool.Close)

	redisContainer, err := tcredis.Run(ctx, "redis:7-alpine")
	if err != nil {
		t.Fatalf("failed to start redis container: %v", err)
	}
	t.Cleanup(func() {
		if err := testcontainers.TerminateContainer(redisContainer); err != nil {
			t.Logf("failed to terminate redis container: %v", err)
		}
	})

	redisURI, err := redisContainer.ConnectionString(ctx)
	if err != nil {
		t.Fatalf("failed to get redis connection string: %v", err)
	}
	redisAddr := strings.TrimPrefix(redisURI, "redis://")

	redisCache := cache.New(redisAddr, "", 0)
	t.Cleanup(func() { _ = redisCache.Close() })
	if err := redisCache.Ping(ctx); err != nil {
		t.Fatalf("failed to ping redis container: %v", err)
	}

	return &testEnv{DB: db, Pool: pool, Cache: redisCache}
}

// applyMigrations replays every migration/*.up.sql file against the fresh
// container in filename order (they're zero-padded, so lexicographic order is
// migration order) — the same files docker-compose's `migrate` service and CI
// apply, just executed directly instead of shelling out to golang-migrate.
func applyMigrations(t *testing.T, ctx context.Context, dsn string) {
	t.Helper()

	// The simple query protocol is required here (and nowhere else in this
	// codebase) because each migration file is a batch of several ;-separated
	// statements, which pgx's default extended-protocol Exec cannot run as one
	// prepared statement.
	cfg, err := pgx.ParseConfig(dsn)
	if err != nil {
		t.Fatalf("failed to parse postgres dsn: %v", err)
	}
	cfg.DefaultQueryExecMode = pgx.QueryExecModeSimpleProtocol

	conn, err := pgx.ConnectConfig(ctx, cfg)
	if err != nil {
		t.Fatalf("failed to connect for migrations: %v", err)
	}
	defer conn.Close(ctx)

	dir := migrationDir(t)
	files, err := filepath.Glob(filepath.Join(dir, "*.up.sql"))
	if err != nil || len(files) == 0 {
		t.Fatalf("no migration files found in %s: %v", dir, err)
	}
	sort.Strings(files)

	for _, f := range files {
		sqlBytes, err := os.ReadFile(f)
		if err != nil {
			t.Fatalf("reading migration %s: %v", f, err)
		}
		if _, err := conn.Exec(ctx, string(sqlBytes)); err != nil {
			t.Fatalf("applying migration %s: %v", filepath.Base(f), err)
		}
	}
}

func migrationDir(t *testing.T) string {
	t.Helper()
	dir, err := filepath.Abs(filepath.Join("..", "..", "migration"))
	if err != nil {
		t.Fatalf("resolving migration dir: %v", err)
	}
	if info, err := os.Stat(dir); err != nil || !info.IsDir() {
		t.Fatalf("migration dir not found at %s (run tests from backend/test/integration): %v", dir, err)
	}
	return dir
}

// tempStorageDir returns a fresh directory for storage.NewLocalDisk, isolated
// per test so parallel tests never collide on the same uploaded file.
func tempStorageDir(t *testing.T) string {
	t.Helper()
	return t.TempDir()
}
