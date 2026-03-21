package plate

import (
	"context"
	"database/sql"
	"fmt"
	"math/rand"
	"sync"
	"time"
)

type ManagedTransaction struct {
	ID        string
	DBID      string
	Tx        *sql.Tx
	ExpiresAt time.Time
	LastUsed  time.Time
}

type TransactionManager struct {
	cfg   Config
	store *DBStore

	mu   sync.Mutex
	txns map[string]*ManagedTransaction
}

func NewTransactionManager(cfg Config, store *DBStore) *TransactionManager {
	return &TransactionManager{
		cfg:   cfg,
		store: store,
		txns:  make(map[string]*ManagedTransaction),
	}
}

func (m *TransactionManager) Run(ctx context.Context) {
	ticker := time.NewTicker(m.cfg.TransactionCleanupInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			m.Close()
			return
		case <-ticker.C:
			m.cleanupExpired()
		}
	}
}

func (m *TransactionManager) Start(ctx context.Context, dbID string) (*ManagedTransaction, error) {
	db, err := m.store.Open(ctx, dbID)
	if err != nil {
		return nil, err
	}
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	managed := &ManagedTransaction{
		ID:        makeTxnID(),
		DBID:      dbID,
		Tx:        tx,
		ExpiresAt: now.Add(m.cfg.TransactionTTL),
		LastUsed:  now,
	}
	m.mu.Lock()
	m.txns[managed.ID] = managed
	m.mu.Unlock()
	return managed, nil
}

func (m *TransactionManager) Get(dbID string, txnID string) (*ManagedTransaction, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	txn, ok := m.txns[txnID]
	if !ok {
		return nil, NewAPIError(404, "transaction_not_found", "transaction not found")
	}
	if txn.DBID != dbID {
		return nil, NewAPIError(404, "transaction_not_found", "transaction not found")
	}
	now := time.Now().UTC()
	if now.After(txn.ExpiresAt) {
		_ = txn.Tx.Rollback()
		delete(m.txns, txnID)
		return nil, NewAPIError(410, "transaction_expired", "transaction expired")
	}
	txn.LastUsed = now
	txn.ExpiresAt = now.Add(m.cfg.TransactionTTL)
	return txn, nil
}

func (m *TransactionManager) Commit(dbID string, txnID string) error {
	m.mu.Lock()
	txn, ok := m.txns[txnID]
	if ok {
		delete(m.txns, txnID)
	}
	m.mu.Unlock()
	if !ok || txn.DBID != dbID {
		return NewAPIError(404, "transaction_not_found", "transaction not found")
	}
	return txn.Tx.Commit()
}

func (m *TransactionManager) Rollback(dbID string, txnID string) error {
	m.mu.Lock()
	txn, ok := m.txns[txnID]
	if ok {
		delete(m.txns, txnID)
	}
	m.mu.Unlock()
	if !ok || txn.DBID != dbID {
		return NewAPIError(404, "transaction_not_found", "transaction not found")
	}
	return txn.Tx.Rollback()
}

func (m *TransactionManager) cleanupExpired() {
	now := time.Now().UTC()
	var expired []*ManagedTransaction
	m.mu.Lock()
	for id, txn := range m.txns {
		if now.After(txn.ExpiresAt) {
			expired = append(expired, txn)
			delete(m.txns, id)
		}
	}
	m.mu.Unlock()
	for _, txn := range expired {
		_ = txn.Tx.Rollback()
	}
}

func (m *TransactionManager) DeleteDB(dbID string) {
	var doomed []*ManagedTransaction
	m.mu.Lock()
	for id, txn := range m.txns {
		if txn.DBID == dbID {
			doomed = append(doomed, txn)
			delete(m.txns, id)
		}
	}
	m.mu.Unlock()
	for _, txn := range doomed {
		_ = txn.Tx.Rollback()
	}
}

func (m *TransactionManager) Close() {
	m.mu.Lock()
	all := make([]*ManagedTransaction, 0, len(m.txns))
	for id, txn := range m.txns {
		all = append(all, txn)
		delete(m.txns, id)
	}
	m.mu.Unlock()
	for _, txn := range all {
		_ = txn.Tx.Rollback()
	}
}

func makeTxnID() string {
	return fmt.Sprintf("txn_%d%x", time.Now().UnixNano(), rand.Uint32())
}
