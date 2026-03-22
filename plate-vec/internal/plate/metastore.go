package plate

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

type EmbeddingProvider string

const (
	ProviderOpenAI     EmbeddingProvider = "openai"
	ProviderOpenRouter EmbeddingProvider = "openrouter"
)

type APIKeyMode string

const (
	APIKeyModeStored APIKeyMode = "stored"
	APIKeyModeHeader APIKeyMode = "header"
)

type EmbeddingProfile struct {
	ID           string            `json:"id"`
	Name         string            `json:"name"`
	Provider     EmbeddingProvider `json:"provider"`
	Model        string            `json:"model"`
	Dimensions   *int              `json:"dimensions,omitempty"`
	APIKeyMode   APIKeyMode        `json:"api_key_mode"`
	SecretCipher string            `json:"secret_cipher,omitempty"`
	CreatedAt    string            `json:"created_at"`
	UpdatedAt    string            `json:"updated_at"`
}

type EmbeddingDefaults struct {
	DefaultProfileID         string `json:"default_profile_id,omitempty"`
	FallbackToHeaderProvider bool   `json:"fallback_to_header_provider"`
}

type CollectionSettings struct {
	Dimension          *int   `json:"dimension,omitempty"`
	DistanceMetric     string `json:"distance_metric,omitempty"`
	EmbeddingProfileID string `json:"embedding_profile_id,omitempty"`
}

type plateMetadata struct {
	Profiles   map[string]*EmbeddingProfile   `json:"profiles"`
	Defaults   EmbeddingDefaults              `json:"defaults"`
	Collection map[string]*CollectionSettings `json:"collection"`
}

type MetaStore struct {
	mu      sync.Mutex
	baseDir string
	key     [32]byte
}

func NewMetaStore(baseDir string, serviceKey string) (*MetaStore, error) {
	trimmed := strings.TrimSpace(serviceKey)
	if trimmed == "" {
		return nil, errors.New("service key is required")
	}
	store := &MetaStore{baseDir: filepath.Join(baseDir, "meta"), key: sha256.Sum256([]byte(trimmed))}
	if err := os.MkdirAll(store.baseDir, 0o755); err != nil {
		return nil, err
	}
	return store, nil
}

func (m *MetaStore) Close() error {
	return nil
}

func (m *MetaStore) DeletePlate(plateID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	path := m.platePath(plateID)
	if err := os.Remove(path); err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	return nil
}

func (m *MetaStore) ListProfiles(plateID string) ([]EmbeddingProfile, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	meta, err := m.readLocked(plateID)
	if err != nil {
		return nil, err
	}
	result := make([]EmbeddingProfile, 0, len(meta.Profiles))
	for _, profile := range meta.Profiles {
		result = append(result, sanitizeProfile(*profile))
	}
	sort.Slice(result, func(i int, j int) bool { return result[i].ID < result[j].ID })
	return result, nil
}

func (m *MetaStore) GetProfile(plateID string, profileID string) (*EmbeddingProfile, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	meta, err := m.readLocked(plateID)
	if err != nil {
		return nil, err
	}
	profile, ok := meta.Profiles[profileID]
	if !ok {
		return nil, nil
	}
	copy := *profile
	return &copy, nil
}

func (m *MetaStore) SaveProfile(plateID string, profile *EmbeddingProfile, plainSecret string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	meta, err := m.readLocked(plateID)
	if err != nil {
		return err
	}
	if profile.ID == "" {
		profile.ID = fmt.Sprintf("prof_%d", time.Now().UnixNano())
	}
	now := time.Now().UTC().Format(time.RFC3339Nano)
	if existing, ok := meta.Profiles[profile.ID]; ok {
		profile.CreatedAt = existing.CreatedAt
	} else if profile.CreatedAt == "" {
		profile.CreatedAt = now
	}
	profile.UpdatedAt = now
	if plainSecret != "" {
		cipherText, encErr := m.encryptSecret(plainSecret)
		if encErr != nil {
			return encErr
		}
		profile.SecretCipher = cipherText
	} else if existing, ok := meta.Profiles[profile.ID]; ok {
		profile.SecretCipher = existing.SecretCipher
	}
	copy := *profile
	meta.Profiles[profile.ID] = &copy
	return m.writeLocked(plateID, meta)
}

func (m *MetaStore) DeleteProfile(plateID string, profileID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	meta, err := m.readLocked(plateID)
	if err != nil {
		return err
	}
	delete(meta.Profiles, profileID)
	if meta.Defaults.DefaultProfileID == profileID {
		meta.Defaults.DefaultProfileID = ""
	}
	for key, settings := range meta.Collection {
		if settings == nil {
			continue
		}
		if settings.EmbeddingProfileID == profileID {
			settings.EmbeddingProfileID = ""
			if settings.Dimension == nil && settings.DistanceMetric == "" {
				delete(meta.Collection, key)
			}
		}
	}
	return m.writeLocked(plateID, meta)
}

func (m *MetaStore) GetDefaults(plateID string) (EmbeddingDefaults, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	meta, err := m.readLocked(plateID)
	if err != nil {
		return EmbeddingDefaults{}, err
	}
	return meta.Defaults, nil
}

func (m *MetaStore) SetDefaults(plateID string, defaults EmbeddingDefaults) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	meta, err := m.readLocked(plateID)
	if err != nil {
		return err
	}
	meta.Defaults = defaults
	return m.writeLocked(plateID, meta)
}

func (m *MetaStore) GetCollectionSettings(plateID string, database string, collection string) (*CollectionSettings, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	meta, err := m.readLocked(plateID)
	if err != nil {
		return nil, err
	}
	settings, ok := meta.Collection[m.collectionKey(database, collection)]
	if !ok || settings == nil {
		return nil, nil
	}
	copy := *settings
	return &copy, nil
}

func (m *MetaStore) SaveCollectionSettings(plateID string, database string, collection string, settings CollectionSettings) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	meta, err := m.readLocked(plateID)
	if err != nil {
		return err
	}
	if settings.Dimension == nil && settings.DistanceMetric == "" && settings.EmbeddingProfileID == "" {
		delete(meta.Collection, m.collectionKey(database, collection))
	} else {
		copy := settings
		meta.Collection[m.collectionKey(database, collection)] = &copy
	}
	return m.writeLocked(plateID, meta)
}

func (m *MetaStore) DeleteCollectionSettings(plateID string, database string, collection string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	meta, err := m.readLocked(plateID)
	if err != nil {
		return err
	}
	delete(meta.Collection, m.collectionKey(database, collection))
	return m.writeLocked(plateID, meta)
}

func (m *MetaStore) DecryptProfileSecret(profile *EmbeddingProfile) (string, error) {
	if profile == nil || profile.SecretCipher == "" {
		return "", nil
	}
	return m.decryptSecret(profile.SecretCipher)
}

func sanitizeProfile(profile EmbeddingProfile) EmbeddingProfile {
	profile.SecretCipher = ""
	return profile
}

func (m *MetaStore) platePath(plateID string) string {
	return filepath.Join(m.baseDir, safeFileName(plateID)+".json")
}

func (m *MetaStore) collectionKey(database string, collection string) string {
	return strings.TrimSpace(database) + "/" + strings.TrimSpace(collection)
}

func (m *MetaStore) readLocked(plateID string) (*plateMetadata, error) {
	path := m.platePath(plateID)
	content, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return &plateMetadata{Profiles: map[string]*EmbeddingProfile{}, Collection: map[string]*CollectionSettings{}}, nil
		}
		return nil, err
	}
	var meta plateMetadata
	if err := json.Unmarshal(content, &meta); err != nil {
		return nil, err
	}
	if meta.Profiles == nil {
		meta.Profiles = map[string]*EmbeddingProfile{}
	}
	if meta.Collection == nil {
		meta.Collection = map[string]*CollectionSettings{}
	}
	return &meta, nil
}

func (m *MetaStore) writeLocked(plateID string, meta *plateMetadata) error {
	path := m.platePath(plateID)
	if meta.Profiles == nil {
		meta.Profiles = map[string]*EmbeddingProfile{}
	}
	if meta.Collection == nil {
		meta.Collection = map[string]*CollectionSettings{}
	}
	encoded, err := json.MarshalIndent(meta, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, encoded, 0o600)
}

func safeFileName(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return "plate"
	}
	replacer := strings.NewReplacer("/", "_", "\\", "_", ":", "_", "*", "_", "?", "_", "\"", "_", "<", "_", ">", "_", "|", "_")
	return replacer.Replace(value)
}

func (m *MetaStore) encryptSecret(secret string) (string, error) {
	block, err := aes.NewCipher(m.key[:])
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}
	sealed := gcm.Seal(nonce, nonce, []byte(secret), nil)
	return base64.StdEncoding.EncodeToString(sealed), nil
}

func (m *MetaStore) decryptSecret(cipherText string) (string, error) {
	raw, err := base64.StdEncoding.DecodeString(cipherText)
	if err != nil {
		return "", err
	}
	block, err := aes.NewCipher(m.key[:])
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonceSize := gcm.NonceSize()
	if len(raw) < nonceSize {
		return "", errors.New("invalid ciphertext")
	}
	nonce := raw[:nonceSize]
	data := raw[nonceSize:]
	plain, err := gcm.Open(nil, nonce, data, nil)
	if err != nil {
		return "", err
	}
	return string(plain), nil
}
