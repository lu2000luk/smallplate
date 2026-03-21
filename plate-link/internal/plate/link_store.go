package plate

import (
	"context"
	crand "crypto/rand"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	linkIDChars        = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	linkBaseIDLength   = 5
	linkAttemptsPerLen = 10
)

var (
	linkIDPrefixPattern = regexp.MustCompile(`^[A-Za-z0-9_-]*$`)
	placeholderPattern  = regexp.MustCompile(`\{[^{}]*\}`)
)

type LinkRecord struct {
	ID          string         `json:"id"`
	PlateID     string         `json:"plate_id"`
	Destination string         `json:"destination,omitempty"`
	Template    string         `json:"template,omitempty"`
	DynamicMode string         `json:"dynamic_mode"`
	Enabled     bool           `json:"enabled"`
	ExpiresAt   int64          `json:"expires_at,omitempty"`
	MaxUses     int64          `json:"max_uses,omitempty"`
	Uses        int64          `json:"uses"`
	Metadata    map[string]any `json:"metadata,omitempty"`
	CreatedAt   int64          `json:"created_at"`
	UpdatedAt   int64          `json:"updated_at"`
	IDPrefix    string         `json:"id_prefix,omitempty"`
}

type CreateLinkInput struct {
	PlateID     string
	Destination string
	Template    string
	ExpiresAt   int64
	MaxUses     int64
	Metadata    map[string]any
	IDPrefix    string
}

type UpdateLinkInput struct {
	Destination *string
	Template    *string
	ExpiresAt   *int64
	MaxUses     *int64
	Enabled     *bool
	Metadata    *map[string]any
}

type ResolveResult struct {
	Record      *LinkRecord
	Destination string
	Status      string
}

type LinkStore struct {
	cfg   Config
	redis *redis.Client
}

func NewLinkStore(cfg Config, redisClient *redis.Client) *LinkStore {
	return &LinkStore{cfg: cfg, redis: redisClient}
}

func (s *LinkStore) Create(ctx context.Context, input CreateLinkInput) (*LinkRecord, error) {
	plateID := strings.TrimSpace(input.PlateID)
	if plateID == "" {
		return nil, NewAPIError(http.StatusBadRequest, "invalid_plate_id", "plate id is required")
	}

	idPrefix := strings.TrimSpace(input.IDPrefix)
	if !linkIDPrefixPattern.MatchString(idPrefix) {
		return nil, NewAPIError(http.StatusBadRequest, "invalid_id_prefix", "id_prefix can only contain letters, numbers, underscore, and dash")
	}

	destination := strings.TrimSpace(input.Destination)
	template := strings.TrimSpace(input.Template)
	if destination == "" && template == "" {
		return nil, NewAPIError(http.StatusBadRequest, "invalid_request", "destination or template is required")
	}
	if destination != "" {
		normalized, err := normalizeURLLike(destination)
		if err != nil {
			return nil, err
		}
		destination = normalized
	}
	if template != "" {
		normalized, err := normalizeTemplate(template)
		if err != nil {
			return nil, err
		}
		template = normalized
	}

	if input.ExpiresAt < 0 {
		return nil, NewAPIError(http.StatusBadRequest, "invalid_expiry", "expires_at must be >= 0")
	}
	if input.MaxUses < 0 {
		return nil, NewAPIError(http.StatusBadRequest, "invalid_max_uses", "max_uses must be >= 0")
	}

	metadata := input.Metadata
	if metadata == nil {
		metadata = map[string]any{}
	}
	metadataJSON, err := json.Marshal(metadata)
	if err != nil {
		return nil, NewAPIError(http.StatusBadRequest, "invalid_metadata", "metadata must be JSON-serializable")
	}

	id, err := s.generateUniqueID(ctx, idPrefix)
	if err != nil {
		return nil, err
	}
	now := time.Now().UnixMilli()

	record := &LinkRecord{
		ID:          id,
		PlateID:     plateID,
		Destination: destination,
		Template:    template,
		DynamicMode: "mixed",
		Enabled:     true,
		ExpiresAt:   input.ExpiresAt,
		MaxUses:     input.MaxUses,
		Uses:        0,
		Metadata:    metadata,
		CreatedAt:   now,
		UpdatedAt:   now,
		IDPrefix:    idPrefix,
	}

	linkKey := s.linkKey(plateID, id)
	globalKey := s.globalIDKey(id)
	setKey := s.plateLinksKey(plateID)

	pipe := s.redis.TxPipeline()
	pipe.HSet(ctx, linkKey,
		"id", record.ID,
		"plate_id", record.PlateID,
		"destination", record.Destination,
		"template", record.Template,
		"dynamic_mode", record.DynamicMode,
		"enabled", "1",
		"expires_at", strconv.FormatInt(record.ExpiresAt, 10),
		"max_uses", strconv.FormatInt(record.MaxUses, 10),
		"uses", "0",
		"metadata", string(metadataJSON),
		"created_at", strconv.FormatInt(record.CreatedAt, 10),
		"updated_at", strconv.FormatInt(record.UpdatedAt, 10),
		"id_prefix", record.IDPrefix,
	)
	pipe.Set(ctx, globalKey, plateID, 0)
	pipe.SAdd(ctx, setKey, id)
	if _, err := pipe.Exec(ctx); err != nil {
		return nil, err
	}

	return record, nil
}

func (s *LinkStore) Update(ctx context.Context, plateID string, id string, input UpdateLinkInput) (*LinkRecord, error) {
	record, err := s.GetByPlateAndID(ctx, plateID, id)
	if err != nil {
		return nil, err
	}

	updates := map[string]any{}
	now := time.Now().UnixMilli()

	if input.Destination != nil {
		normalized, err := normalizeURLLike(*input.Destination)
		if err != nil {
			return nil, err
		}
		updates["destination"] = normalized
		record.Destination = normalized
	}
	if input.Template != nil {
		normalized, err := normalizeTemplate(*input.Template)
		if err != nil {
			return nil, err
		}
		updates["template"] = normalized
		record.Template = normalized
	}
	record.DynamicMode = "mixed"
	if input.ExpiresAt != nil {
		if *input.ExpiresAt < 0 {
			return nil, NewAPIError(http.StatusBadRequest, "invalid_expiry", "expires_at must be >= 0")
		}
		updates["expires_at"] = strconv.FormatInt(*input.ExpiresAt, 10)
		record.ExpiresAt = *input.ExpiresAt
	}
	if input.MaxUses != nil {
		if *input.MaxUses < 0 {
			return nil, NewAPIError(http.StatusBadRequest, "invalid_max_uses", "max_uses must be >= 0")
		}
		updates["max_uses"] = strconv.FormatInt(*input.MaxUses, 10)
		record.MaxUses = *input.MaxUses
	}
	if input.Enabled != nil {
		if *input.Enabled {
			updates["enabled"] = "1"
		} else {
			updates["enabled"] = "0"
		}
		record.Enabled = *input.Enabled
	}
	if input.Metadata != nil {
		payload, err := json.Marshal(*input.Metadata)
		if err != nil {
			return nil, NewAPIError(http.StatusBadRequest, "invalid_metadata", "metadata must be JSON-serializable")
		}
		updates["metadata"] = string(payload)
		record.Metadata = *input.Metadata
	}

	if len(updates) == 0 {
		return record, nil
	}

	updates["updated_at"] = strconv.FormatInt(now, 10)
	record.UpdatedAt = now

	if err := s.redis.HSet(ctx, s.linkKey(plateID, id), updates).Err(); err != nil {
		return nil, err
	}

	return record, nil
}

func (s *LinkStore) Delete(ctx context.Context, plateID string, id string) error {
	linkKey := s.linkKey(plateID, id)
	exists, err := s.redis.Exists(ctx, linkKey).Result()
	if err != nil {
		return err
	}
	if exists == 0 {
		return NewAPIError(http.StatusNotFound, "not_found", "link not found")
	}

	pipe := s.redis.TxPipeline()
	pipe.Del(ctx, linkKey)
	pipe.Del(ctx, s.globalIDKey(id))
	pipe.SRem(ctx, s.plateLinksKey(plateID), id)
	_, err = pipe.Exec(ctx)
	return err
}

func (s *LinkStore) List(ctx context.Context, plateID string) ([]*LinkRecord, error) {
	ids, err := s.redis.SMembers(ctx, s.plateLinksKey(plateID)).Result()
	if err != nil {
		return nil, err
	}
	out := make([]*LinkRecord, 0, len(ids))
	for _, id := range ids {
		record, err := s.GetByPlateAndID(ctx, plateID, id)
		if err != nil {
			continue
		}
		out = append(out, record)
	}
	return out, nil
}

func (s *LinkStore) GetByPlateAndID(ctx context.Context, plateID string, id string) (*LinkRecord, error) {
	raw, err := s.redis.HGetAll(ctx, s.linkKey(plateID, id)).Result()
	if err != nil {
		return nil, err
	}
	if len(raw) == 0 {
		return nil, NewAPIError(http.StatusNotFound, "not_found", "link not found")
	}
	return parseLinkRecord(raw)
}

func (s *LinkStore) Resolve(ctx context.Context, plateID string, id string, tail []string, query url.Values) (*ResolveResult, error) {
	key := s.linkKey(plateID, id)
	result, err := resolveScript.Run(ctx, s.redis, []string{key}, strconv.FormatInt(time.Now().UnixMilli(), 10)).Result()
	if err != nil {
		return nil, err
	}

	parts, ok := result.([]any)
	if !ok || len(parts) < 2 {
		return nil, NewAPIError(http.StatusInternalServerError, "invalid_resolve_result", "unexpected resolve result")
	}
	status := toString(parts[0])
	if status != "ok" {
		switch status {
		case "not_found":
			return nil, NewAPIError(http.StatusNotFound, "not_found", "link not found")
		case "disabled":
			return nil, NewAPIError(http.StatusGone, "disabled", "link is disabled")
		case "expired":
			return nil, NewAPIError(http.StatusGone, "expired", "link expired")
		case "max_uses_reached":
			return nil, NewAPIError(http.StatusGone, "max_uses_reached", "link max uses reached")
		default:
			return nil, NewAPIError(http.StatusGone, "unavailable", "link unavailable")
		}
	}

	if len(parts) < 11 {
		return nil, NewAPIError(http.StatusInternalServerError, "invalid_resolve_result", "incomplete resolve payload")
	}

	record := &LinkRecord{
		ID:          id,
		PlateID:     plateID,
		Destination: toString(parts[2]),
		Template:    toString(parts[3]),
		DynamicMode: "mixed",
		Enabled:     true,
		ExpiresAt:   toInt64(parts[5]),
		MaxUses:     toInt64(parts[6]),
		Uses:        toInt64(parts[1]),
		Metadata:    parseMetadata(toString(parts[7])),
		CreatedAt:   toInt64(parts[8]),
		UpdatedAt:   toInt64(parts[9]),
		IDPrefix:    toString(parts[10]),
	}

	destination, err := renderDestination(record, tail, query)
	if err != nil {
		return nil, err
	}

	return &ResolveResult{Record: record, Destination: destination, Status: "ok"}, nil
}

func (s *LinkStore) ResolveByPublicID(ctx context.Context, id string, tail []string, query url.Values) (*ResolveResult, error) {
	plateID, err := s.redis.Get(ctx, s.globalIDKey(id)).Result()
	if err == redis.Nil {
		return nil, NewAPIError(http.StatusNotFound, "not_found", "link not found")
	}
	if err != nil {
		return nil, err
	}
	return s.Resolve(ctx, plateID, id, tail, query)
}

func (s *LinkStore) DeletePlateData(ctx context.Context, plateID string) error {
	ids, err := s.redis.SMembers(ctx, s.plateLinksKey(plateID)).Result()
	if err != nil && err != redis.Nil {
		return err
	}

	pipe := s.redis.TxPipeline()
	for _, id := range ids {
		pipe.Del(ctx, s.globalIDKey(id))
	}
	pipe.Del(ctx, s.plateLinksKey(plateID))

	match := PrefixPattern(plateID, "link:*")
	var cursor uint64
	for {
		keys, next, err := s.redis.Scan(ctx, cursor, match, 500).Result()
		if err != nil {
			return err
		}
		if len(keys) > 0 {
			pipe.Unlink(ctx, keys...)
		}
		cursor = next
		if cursor == 0 {
			break
		}
	}
	_, err = pipe.Exec(ctx)
	return err
}

func (s *LinkStore) generateUniqueID(ctx context.Context, prefix string) (string, error) {
	length := linkBaseIDLength
	for {
		for attempt := 0; attempt < linkAttemptsPerLen; attempt++ {
			randPart, err := randomIDPart(length)
			if err != nil {
				return "", err
			}
			id := prefix + randPart
			exists, err := s.redis.Exists(ctx, s.globalIDKey(id)).Result()
			if err != nil {
				return "", err
			}
			if exists == 0 {
				return id, nil
			}
		}
		length++
	}
}

func randomIDPart(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := crand.Read(bytes); err != nil {
		return "", err
	}
	buffer := make([]byte, length)
	for i, b := range bytes {
		buffer[i] = linkIDChars[int(b)%len(linkIDChars)]
	}
	return string(buffer), nil
}

func (s *LinkStore) linkKey(plateID string, id string) string {
	return PrefixKey(plateID, "link:"+id)
}

func (s *LinkStore) plateLinksKey(plateID string) string {
	return PrefixKey(plateID, "links")
}

func (s *LinkStore) globalIDKey(id string) string {
	return "link:id:" + id
}

func parseLinkRecord(raw map[string]string) (*LinkRecord, error) {
	record := &LinkRecord{
		ID:          raw["id"],
		PlateID:     raw["plate_id"],
		Destination: raw["destination"],
		Template:    raw["template"],
		DynamicMode: "mixed",
		Enabled:     raw["enabled"] != "0",
		ExpiresAt:   parseInt64(raw["expires_at"]),
		MaxUses:     parseInt64(raw["max_uses"]),
		Uses:        parseInt64(raw["uses"]),
		Metadata:    parseMetadata(raw["metadata"]),
		CreatedAt:   parseInt64(raw["created_at"]),
		UpdatedAt:   parseInt64(raw["updated_at"]),
		IDPrefix:    raw["id_prefix"],
	}
	if record.ID == "" || record.PlateID == "" {
		return nil, NewAPIError(http.StatusInternalServerError, "invalid_record", "link record is malformed")
	}
	return record, nil
}

func normalizeURLLike(raw string) (string, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return "", NewAPIError(http.StatusBadRequest, "invalid_destination", "destination is required")
	}
	if !strings.HasPrefix(strings.ToLower(trimmed), "http://") && !strings.HasPrefix(strings.ToLower(trimmed), "https://") {
		trimmed = "https://" + trimmed
	}
	if _, err := url.ParseRequestURI(trimmed); err != nil {
		return "", NewAPIError(http.StatusBadRequest, "invalid_destination", "destination must be a valid URL")
	}
	return trimmed, nil
}

func normalizeTemplate(raw string) (string, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return "", NewAPIError(http.StatusBadRequest, "invalid_template", "template is required")
	}
	if !strings.HasPrefix(strings.ToLower(trimmed), "http://") && !strings.HasPrefix(strings.ToLower(trimmed), "https://") {
		trimmed = "https://" + trimmed
	}
	if _, err := url.Parse(trimmed); err != nil {
		return "", NewAPIError(http.StatusBadRequest, "invalid_template", "template must be URL-like")
	}
	return trimmed, nil
}

func renderDestination(record *LinkRecord, tail []string, query url.Values) (string, error) {
	if record.Template == "" {
		if record.Destination == "" {
			return "", NewAPIError(http.StatusInternalServerError, "invalid_link", "link has no destination")
		}
		return record.Destination, nil
	}

	index := 0
	unnamedIndex := 0
	rendered := placeholderPattern.ReplaceAllStringFunc(record.Template, func(match string) string {
		name := strings.TrimSuffix(strings.TrimPrefix(match, "{"), "}")
		if strings.TrimSpace(name) == "" {
			if index < len(tail) {
				value := tail[index]
				index++
				unnamedIndex++
				return value
			}
			unnamedIndex++
			fallback := query.Get(fmt.Sprintf("p%d", unnamedIndex))
			if fallback != "" {
				return fallback
			}
			return match
		}
		if value := query.Get(name); value != "" {
			return value
		}
		return match
	})

	if _, err := url.ParseRequestURI(rendered); err != nil {
		return "", NewAPIError(http.StatusBadRequest, "unresolved_template", "template could not be resolved to a valid URL")
	}
	return rendered, nil
}

func parseMetadata(value string) map[string]any {
	if strings.TrimSpace(value) == "" {
		return map[string]any{}
	}
	var out map[string]any
	if err := json.Unmarshal([]byte(value), &out); err != nil || out == nil {
		return map[string]any{}
	}
	return out
}

func parseInt64(value string) int64 {
	parsed, err := strconv.ParseInt(strings.TrimSpace(value), 10, 64)
	if err != nil {
		return 0
	}
	return parsed
}

func toString(value any) string {
	switch typed := value.(type) {
	case string:
		return typed
	case []byte:
		return string(typed)
	default:
		return fmt.Sprint(value)
	}
}

func toInt64(value any) int64 {
	switch typed := value.(type) {
	case int64:
		return typed
	case int:
		return int64(typed)
	case string:
		return parseInt64(typed)
	case []byte:
		return parseInt64(string(typed))
	default:
		return parseInt64(fmt.Sprint(value))
	}
}

var resolveScript = redis.NewScript(`
local key = KEYS[1]
local now = tonumber(ARGV[1])

if redis.call("EXISTS", key) == 0 then
  return {"not_found"}
end

local enabled = redis.call("HGET", key, "enabled")
if enabled == "0" then
  return {"disabled"}
end

local expires_at = tonumber(redis.call("HGET", key, "expires_at") or "0")
if expires_at > 0 and now > expires_at then
  return {"expired"}
end

local max_uses = tonumber(redis.call("HGET", key, "max_uses") or "0")
local uses = tonumber(redis.call("HGET", key, "uses") or "0")
if max_uses > 0 and uses >= max_uses then
  return {"max_uses_reached"}
end

uses = redis.call("HINCRBY", key, "uses", 1)
redis.call("HSET", key, "updated_at", tostring(now))

local destination = redis.call("HGET", key, "destination") or ""
local template = redis.call("HGET", key, "template") or ""
local dynamic_mode = redis.call("HGET", key, "dynamic_mode") or "mixed"
local metadata = redis.call("HGET", key, "metadata") or "{}"
local created_at = redis.call("HGET", key, "created_at") or "0"
local updated_at = redis.call("HGET", key, "updated_at") or tostring(now)
local id_prefix = redis.call("HGET", key, "id_prefix") or ""

return {"ok", tostring(uses), destination, template, dynamic_mode, tostring(expires_at), tostring(max_uses), metadata, created_at, updated_at, id_prefix}
`)
