// Command api is the Book Reader backend's HTTP entrypoint. It loads config,
// wires the clean-architecture layers (repository -> service -> handler) by
// hand (no reflection-based DI container — explicit constructors are easier to
// trace and test), and serves until an OS signal requests graceful shutdown.
//
//	@title			Book Reader API
//	@version		1.0
//	@description	REST API for the Book Reader platform: public reading endpoints
//	@description	(books, chapters, verses, search, favorites, bookmarks, notes)
//	@description	and role-gated admin endpoints (books CRUD, Excel/CSV import,
//	@description	reference data, users, audit log).
//	@BasePath		/api/v1
//	@securityDefinitions.apikey	CookieAuth
//	@in							cookie
//	@name						access_token
//	@description				JWT access token, set as an httpOnly cookie by /auth/login.
package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/jackc/pgx/v5/pgxpool"

	"bookreader/backend/internal/config"
	adminh "bookreader/backend/internal/handler/admin"
	publich "bookreader/backend/internal/handler/public"
	"bookreader/backend/internal/repository/postgres"
	"bookreader/backend/internal/router"
	"bookreader/backend/internal/service"
	"bookreader/backend/pkg/cache"
	"bookreader/backend/pkg/jwtutil"
	"bookreader/backend/pkg/logger"
	"bookreader/backend/pkg/storage"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		panic(err)
	}
	log := logger.New(cfg.Env, cfg.Server.LogLevel)

	db, err := postgres.NewDB(cfg)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to connect to postgres (gorm)")
	}
	sqlDB, _ := db.DB()
	defer sqlDB.Close()

	pgxPool, err := pgxpool.New(context.Background(), cfg.Postgres.DSN)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to connect to postgres (pgx pool)")
	}
	defer pgxPool.Close()

	redisCache := cache.New(cfg.Redis.Addr, cfg.Redis.Password, cfg.Redis.DB)
	if err := redisCache.Ping(context.Background()); err != nil {
		log.Fatal().Err(err).Msg("failed to connect to redis")
	}
	defer redisCache.Close()
	cacheSvc := service.NewCacheService(redisCache)

	storageProvider := buildStorageProvider(cfg)
	issuer := jwtutil.NewIssuer(cfg.JWT.AccessSecret, cfg.JWT.AccessTTL, cfg.JWT.Issuer)

	// --- repositories ---
	roleRepo := postgres.NewRoleRepository(db)
	userRepo := postgres.NewUserRepository(db)
	refreshRepo := postgres.NewRefreshTokenRepository(db)
	sessionRepo := postgres.NewSessionRepository(db)
	languageRepo := postgres.NewLanguageRepository(db)
	categoryRepo := postgres.NewCategoryRepository(db)
	fileRepo := postgres.NewFileRepository(db)
	bookRepo := postgres.NewBookRepository(db)
	chapterRepo := postgres.NewChapterRepository(db)
	sectionRepo := postgres.NewSectionRepository(db)
	verseRepo := postgres.NewVerseRepository(db, pgxPool)
	pdfRepo := postgres.NewPDFRepository(db)
	favoriteRepo := postgres.NewFavoriteRepository(db)
	bookmarkRepo := postgres.NewBookmarkRepository(db)
	noteRepo := postgres.NewNoteRepository(db)
	importLogRepo := postgres.NewImportLogRepository(db)
	auditLogRepo := postgres.NewAuditLogRepository(db)
	statsRepo := postgres.NewStatsRepository(db)
	searchRepo := postgres.NewSearchRepository(db)

	// --- services ---
	authSvc := service.NewAuthService(userRepo, roleRepo, refreshRepo, sessionRepo, issuer, cfg.JWT.RefreshTTL)
	userSvc := service.NewUserService(userRepo, roleRepo)
	languageSvc := service.NewLanguageService(languageRepo)
	categorySvc := service.NewCategoryService(categoryRepo)
	bookSvc := service.NewBookService(bookRepo, chapterRepo, sectionRepo, verseRepo, cacheSvc)
	fileSvc := service.NewFileService(fileRepo, pdfRepo, storageProvider)
	favoriteSvc := service.NewFavoriteService(favoriteRepo)
	bookmarkSvc := service.NewBookmarkService(bookmarkRepo)
	noteSvc := service.NewNoteService(noteRepo)
	auditSvc := service.NewAuditService(auditLogRepo)
	statsSvc := service.NewStatsService(statsRepo)
	searchSvc := service.NewSearchService(searchRepo, cacheSvc)
	importSvc := service.NewImportService(fileRepo, bookRepo, chapterRepo, sectionRepo, verseRepo, importLogRepo, storageProvider, cacheSvc)

	// --- handlers ---
	h := &router.Handlers{
		Auth:      publich.NewAuthHandler(authSvc, cfg),
		Me:        publich.NewMeHandler(userSvc),
		Books:     publich.NewBookHandler(bookSvc, favoriteSvc, fileSvc, pdfRepo),
		Favorites: publich.NewFavoriteHandler(favoriteSvc, fileSvc),
		Bookmarks: publich.NewBookmarkHandler(bookmarkSvc),
		Notes:     publich.NewNoteHandler(noteSvc),
		Search:    publich.NewSearchHandler(searchSvc),

		AdminBooks:      adminh.NewBookHandler(bookSvc, fileSvc, pdfRepo),
		AdminFiles:      adminh.NewFileHandler(fileRepo, bookRepo, fileSvc),
		AdminLanguages:  adminh.NewLanguageHandler(languageSvc),
		AdminCategories: adminh.NewCategoryHandler(categorySvc),
		AdminUsers:      adminh.NewUserHandler(userSvc),
		AdminStats:      adminh.NewStatsHandler(statsSvc),
		AdminAudit:      adminh.NewAuditHandler(auditSvc),
		AdminImport:     adminh.NewImportHandler(importSvc, fileSvc),
	}

	engine := router.New(cfg, log, issuer, redisCache, h, &router.Services{Audit: auditSvc})

	srv := &http.Server{
		Addr:         ":" + cfg.Server.Port,
		Handler:      engine,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
	}

	go func() {
		log.Info().Str("port", cfg.Server.Port).Str("env", cfg.Env).Msg("starting server")
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal().Err(err).Msg("server failed to start")
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info().Msg("shutting down gracefully")
	ctx, cancel := context.WithTimeout(context.Background(), cfg.Server.ShutdownTimeout)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Error().Err(err).Msg("graceful shutdown failed")
	}
}

func buildStorageProvider(cfg *config.Config) storage.Provider {
	if cfg.Storage.Provider == "s3" {
		s3, err := storage.NewS3(storage.S3Config{
			Bucket: cfg.Storage.S3Bucket, Region: cfg.Storage.S3Region, Endpoint: cfg.Storage.S3Endpoint,
			AccessKey: cfg.Storage.S3AccessKey, SecretKey: cfg.Storage.S3SecretKey,
			UsePathStyle: cfg.Storage.S3PathStyle, PublicURL: cfg.Storage.S3PublicURL,
		})
		if err != nil {
			panic(err)
		}
		return s3
	}
	_ = os.MkdirAll(cfg.Storage.LocalBaseDir, 0o755)
	return storage.NewLocalDisk(cfg.Storage.LocalBaseDir, cfg.Storage.LocalPublic)
}
