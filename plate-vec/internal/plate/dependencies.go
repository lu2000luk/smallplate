package plate

import (
	"fmt"
	"os"
)

type Dependencies struct {
	Config    Config
	AuthCache *AuthCache
	Manager   *ManagerClient
	Meta      *MetaStore
	Chroma    *ChromaClient
}

func NewDependencies(cfg Config) (*Dependencies, error) {
	cache, err := NewAuthCache(cfg.AuthCacheSize)
	if err != nil {
		return nil, err
	}
	if err := os.MkdirAll(cfg.DataDir, 0o755); err != nil {
		return nil, err
	}

	meta, err := NewMetaStore(cfg.DataDir, cfg.ServiceKey)
	if err != nil {
		return nil, err
	}

	deps := &Dependencies{
		Config:    cfg,
		AuthCache: cache,
		Meta:      meta,
		Chroma:    NewChromaClient(cfg.ChromaURL, cfg.OpTimeout),
	}
	deps.Manager = NewManagerClient(cfg, cache, meta)
	return deps, nil
}

func (d *Dependencies) Close() error {
	var errs []error
	if d.Meta != nil {
		if err := d.Meta.Close(); err != nil {
			errs = append(errs, err)
		}
	}
	if len(errs) == 0 {
		return nil
	}
	return fmt.Errorf("close errors: %v", errs)
}
