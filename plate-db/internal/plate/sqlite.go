package plate

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	lru "github.com/hashicorp/golang-lru/v2"
	_ "turso.tech/database/tursogo"
)

type Dependencies struct {
	Config       Config
	AuthCache    *AuthCache
	Manager      *ManagerClient
	DBs          *DBStore
	Transactions *TransactionManager
}

func NewDependencies(cfg Config) (*Dependencies, error) {
	cache, err := NewAuthCache(cfg.AuthCacheSize)
	if err != nil {
		return nil, err
	}
	if err := os.MkdirAll(cfg.DataDir, 0o755); err != nil {
		return nil, err
	}

	dbStore, err := NewDBStore(cfg)
	if err != nil {
		return nil, err
	}
	txManager := NewTransactionManager(cfg, dbStore)

	deps := &Dependencies{
		Config:       cfg,
		AuthCache:    cache,
		DBs:          dbStore,
		Transactions: txManager,
	}
	deps.Manager = NewManagerClient(cfg, cache, dbStore, txManager)
	return deps, nil
}

func (d *Dependencies) Close() error {
	if d.Transactions != nil {
		d.Transactions.Close()
	}
	if d.DBs != nil {
		return d.DBs.Close()
	}
	return nil
}

var dbIDPattern = regexp.MustCompile(`^[A-Za-z0-9_-]+$`)

type DBStore struct {
	cfg   Config
	mu    sync.Mutex
	cache *lru.Cache[string, *sql.DB]
}

func NewDBStore(cfg Config) (*DBStore, error) {
	store := &DBStore{cfg: cfg}
	cache, err := lru.NewWithEvict[string, *sql.DB](cfg.ConnectionCacheSize, func(_ string, value *sql.DB) {
		_ = value.Close()
	})
	if err != nil {
		return nil, err
	}
	store.cache = cache
	return store, nil
}

func (s *DBStore) validateID(id string) error {
	id = strings.TrimSpace(id)
	if id == "" {
		return NewAPIError(400, "invalid_id", "database id is required")
	}
	if !dbIDPattern.MatchString(id) {
		return NewAPIError(400, "invalid_id", "database id must contain only letters, numbers, underscore, and dash")
	}
	return nil
}

func (s *DBStore) dbPath(id string) (string, error) {
	if err := s.validateID(id); err != nil {
		return "", err
	}
	return filepath.Join(s.cfg.DataDir, id+".db"), nil
}

func (s *DBStore) getOrOpen(id string) (*sql.DB, error) {
	if err := s.validateID(id); err != nil {
		return nil, err
	}

	s.mu.Lock()
	if existing, ok := s.cache.Get(id); ok {
		s.mu.Unlock()
		return existing, nil
	}
	s.mu.Unlock()

	path, err := s.dbPath(id)
	if err != nil {
		return nil, err
	}
	db, err := sql.Open("turso", path)
	if err != nil {
		return nil, err
	}
	db.SetConnMaxLifetime(0)
	db.SetMaxIdleConns(1)
	db.SetMaxOpenConns(1)

	ctx, cancel := context.WithTimeout(context.Background(), s.cfg.OpTimeout)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, err
	}
	if err := s.initDB(ctx, db); err != nil {
		_ = db.Close()
		return nil, err
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	if existing, ok := s.cache.Get(id); ok {
		_ = db.Close()
		return existing, nil
	}
	s.cache.Add(id, db)
	return db, nil
}

func (s *DBStore) initDB(ctx context.Context, db *sql.DB) error {
	statements := []string{
		"PRAGMA foreign_keys=ON;",
		"PRAGMA busy_timeout=5000;",
		"CREATE TABLE IF NOT EXISTS __plate_meta (key TEXT PRIMARY KEY, value TEXT NOT NULL);",
		"INSERT OR IGNORE INTO __plate_meta (key, value) VALUES ('created_at', ?);",
	}
	for idx, statement := range statements {
		if idx == len(statements)-1 {
			if _, err := db.ExecContext(ctx, statement, time.Now().UTC().Format(time.RFC3339Nano)); err != nil {
				return err
			}
			continue
		}
		if _, err := db.ExecContext(ctx, statement); err != nil {
			return err
		}
	}
	return nil
}

func (s *DBStore) Ensure(ctx context.Context, id string) (bool, error) {
	path, err := s.dbPath(id)
	if err != nil {
		return false, err
	}
	_, statErr := os.Stat(path)
	created := errors.Is(statErr, os.ErrNotExist)
	if _, err := s.getOrOpen(id); err != nil {
		return false, err
	}
	if created {
		Info("Created database:", id)
	}
	return created, nil
}

func (s *DBStore) Delete(ctx context.Context, id string) error {
	_ = ctx
	path, err := s.dbPath(id)
	if err != nil {
		return err
	}

	s.mu.Lock()
	if existing, ok := s.cache.Peek(id); ok {
		s.cache.Remove(id)
		_ = existing.Close()
	}
	s.mu.Unlock()

	if err := os.Remove(path); err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	return nil
}

func (s *DBStore) ReplaceWithBytes(ctx context.Context, id string, payload []byte) error {
	_ = ctx
	path, err := s.dbPath(id)
	if err != nil {
		return err
	}

	s.mu.Lock()
	if existing, ok := s.cache.Peek(id); ok {
		s.cache.Remove(id)
		_ = existing.Close()
	}
	s.mu.Unlock()

	if err := os.WriteFile(path, payload, 0o644); err != nil {
		return err
	}
	_, err = s.getOrOpen(id)
	return err
}

func (s *DBStore) Open(ctx context.Context, id string) (*sql.DB, error) {
	_ = ctx
	return s.getOrOpen(id)
}

func (s *DBStore) FilePath(id string) (string, error) {
	return s.dbPath(id)
}

func (s *DBStore) CreatedAt(ctx context.Context, id string) (string, error) {
	db, err := s.Open(ctx, id)
	if err != nil {
		return "", err
	}
	var createdAt string
	if err := db.QueryRowContext(ctx, "SELECT value FROM __plate_meta WHERE key='created_at' LIMIT 1").Scan(&createdAt); err != nil {
		return "", err
	}
	return createdAt, nil
}

func (s *DBStore) TableCount(ctx context.Context, id string) (int, error) {
	db, err := s.Open(ctx, id)
	if err != nil {
		return 0, err
	}
	var count int
	err = db.QueryRowContext(ctx, "SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name NOT LIKE 'sqlite_%' AND name != '__plate_meta'").Scan(&count)
	return count, err
}

func (s *DBStore) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	var errs []error
	for _, key := range s.cache.Keys() {
		db, ok := s.cache.Peek(key)
		if !ok {
			continue
		}
		if err := db.Close(); err != nil {
			errs = append(errs, err)
		}
		s.cache.Remove(key)
	}
	if len(errs) == 0 {
		return nil
	}
	return fmt.Errorf("close errors: %v", errs)
}
