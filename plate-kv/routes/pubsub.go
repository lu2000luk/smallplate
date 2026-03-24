// Endpoints contained in this file:
// POST /{plateID}/publish/{channel}
// GET /{plateID}/subscribe/{channel}
// GET /{plateID}/ws/subscribe/{channel}
// POST /{plateID}/pubsub/{channel}/publish
// GET /{plateID}/pubsub/{channel}/subscribe
// GET /{plateID}/pubsub/{channel}/ws
// GET /{plateID}/pubsub/{channel}/client
// GET /{plateID}/s/{token}
package routes

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/redis/go-redis/v9"

	"plain/kv/internal/plate"
)

const (
	clientSocketTokenBytes     = 32
	clientSocketDefaultExpiry  = int64(5 * 60 * 1000)
	clientSocketMaxExpiry      = int64(10 * 60 * 1000)
	clientSocketJanitorTickDur = 30 * time.Second
)

var (
	pubSubUpgrader = websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}

	errClientSocketNotFound  = errors.New("client socket token not found")
	errClientSocketExpired   = errors.New("client socket token expired")
	errClientSocketExhausted = errors.New("client socket token exhausted")
)

type clientSocketTicket struct {
	PlateID  string
	Channel  string
	Pattern  bool
	MaxDurMS int64
	Expires  time.Time
	MaxUses  int64
	Uses     int64
}

type clientSocketStore struct {
	mu      sync.RWMutex
	tickets map[string]*clientSocketTicket
}

func newClientSocketStore() *clientSocketStore {
	return &clientSocketStore{tickets: make(map[string]*clientSocketTicket)}
}

func (s *clientSocketStore) startJanitor() {
	go func() {
		ticker := time.NewTicker(clientSocketJanitorTickDur)
		defer ticker.Stop()
		for now := range ticker.C {
			s.cleanupExpired(now)
		}
	}()
}

func (s *clientSocketStore) cleanupExpired(now time.Time) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for token, ticket := range s.tickets {
		if !now.Before(ticket.Expires) {
			delete(s.tickets, token)
		}
	}
}

func (s *clientSocketStore) issue(plateID string, channel string, pattern bool, maxDurMS int64, expiryMS int64, maxUses int64) (string, error) {
	expiresAt := time.Now().Add(time.Duration(expiryMS) * time.Millisecond)
	for range 8 {
		token, err := generateClientSocketToken()
		if err != nil {
			return "", err
		}
		s.mu.Lock()
		if _, exists := s.tickets[token]; !exists {
			s.tickets[token] = &clientSocketTicket{
				PlateID:  plateID,
				Channel:  channel,
				Pattern:  pattern,
				MaxDurMS: maxDurMS,
				Expires:  expiresAt,
				MaxUses:  maxUses,
			}
			s.mu.Unlock()
			return token, nil
		}
		s.mu.Unlock()
	}
	return "", errors.New("failed to generate unique client socket token")
}

func (s *clientSocketStore) consume(plateID string, token string) (clientSocketTicket, error) {
	now := time.Now()
	s.mu.Lock()
	defer s.mu.Unlock()
	ticket, ok := s.tickets[token]
	if !ok || ticket.PlateID != plateID {
		return clientSocketTicket{}, errClientSocketNotFound
	}
	if !now.Before(ticket.Expires) {
		delete(s.tickets, token)
		return clientSocketTicket{}, errClientSocketExpired
	}
	if ticket.MaxUses > 0 && ticket.Uses >= ticket.MaxUses {
		delete(s.tickets, token)
		return clientSocketTicket{}, errClientSocketExhausted
	}
	ticket.Uses++
	snapshot := *ticket
	if ticket.MaxUses > 0 && ticket.Uses >= ticket.MaxUses {
		delete(s.tickets, token)
	}
	return snapshot, nil
}

func generateClientSocketToken() (string, error) {
	buf := make([]byte, clientSocketTokenBytes)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
}

func registerPubSub(mux *http.ServeMux, deps *plate.Dependencies) {
	clientSockets := newClientSocketStore()
	clientSockets.startJanitor()

	publish := plate.Authenticated(deps, func(w http.ResponseWriter, r *http.Request, plateID string) error {
		channel, err := plate.PathValue(r, "key")
		if err != nil {
			return err
		}
		var request struct {
			Message any `json:"message"`
		}
		if err := plate.DecodeJSON(r, &request); err != nil {
			return err
		}
		payload, err := json.Marshal(request.Message)
		if err != nil {
			return err
		}
		receivers, err := deps.Redis.Publish(r.Context(), plate.PrefixKey(plateID, channel), payload).Result()
		if err != nil {
			return err
		}
		plate.WriteOK(w, http.StatusOK, map[string]any{"receivers": receivers, "channel": channel})
		return nil
	})
	subscribe := plate.Authenticated(deps, func(w http.ResponseWriter, r *http.Request, plateID string) error {
		channel, err := plate.PathValue(r, "key")
		if err != nil {
			return err
		}
		prefixed := plate.PrefixKey(plateID, channel)
		pubsub := deps.PubSub.Subscribe(r.Context(), prefixed)
		defer pubsub.Close()
		if plate.QueryBool(r, "pattern") {
			_ = pubsub.Close()
			pubsub = deps.PubSub.PSubscribe(r.Context(), plate.PrefixPattern(plateID, channel))
			defer pubsub.Close()
		}
		if _, err := pubsub.Receive(r.Context()); err != nil {
			return err
		}
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")
		flusher, ok := w.(http.Flusher)
		if !ok {
			return plate.NewAPIError(http.StatusInternalServerError, "streaming_unsupported", "response writer does not support streaming")
		}
		messageCh := pubsub.Channel(redis.WithChannelSize(deps.Config.PubSubBufferSize))
		heartbeat := time.NewTicker(15 * time.Second)
		defer heartbeat.Stop()
		for {
			select {
			case <-r.Context().Done():
				return nil
			case <-heartbeat.C:
				_, _ = w.Write([]byte(": keepalive\n\n"))
				flusher.Flush()
			case message, ok := <-messageCh:
				if !ok {
					return nil
				}
				payload, _ := json.Marshal(map[string]any{
					"channel": plate.UnprefixKey(plateID, message.Channel),
					"message": decodePubSubPayload(message.Payload),
				})
				_, _ = w.Write([]byte("event: message\n"))
				_, _ = w.Write([]byte("data: "))
				_, _ = w.Write(payload)
				_, _ = w.Write([]byte("\n\n"))
				flusher.Flush()
			}
		}
	})
	createClientSocket := plate.Authenticated(deps, func(w http.ResponseWriter, r *http.Request, plateID string) error {
		channel, err := plate.PathValue(r, "key")
		if err != nil {
			return err
		}

		maxDurMS, err := queryInt64(r, "max_dur", 0)
		if err != nil {
			return err
		}
		if maxDurMS < 0 {
			return plate.NewAPIError(http.StatusBadRequest, "invalid_query", "max_dur must be zero or greater")
		}

		expiryMS, err := parseClientSocketExpiry(r)
		if err != nil {
			return err
		}

		maxUses, err := queryInt64(r, "max_uses", 1)
		if err != nil {
			return err
		}
		if maxUses < 0 {
			return plate.NewAPIError(http.StatusBadRequest, "invalid_query", "max_uses must be zero or greater")
		}

		token, err := clientSockets.issue(plateID, channel, plate.QueryBool(r, "pattern"), maxDurMS, expiryMS, maxUses)
		if err != nil {
			return err
		}

		plate.WriteOK(w, http.StatusOK, map[string]any{
			"path":       fmt.Sprintf("/%s/s/%s", plateID, token),
			"plate_id":   plateID,
			"channel":    channel,
			"pattern":    plate.QueryBool(r, "pattern"),
			"max_dur_ms": maxDurMS,
			"expiry_ms":  expiryMS,
			"max_uses":   maxUses,
		})
		return nil
	})
	mux.HandleFunc("POST /{plateID}/publish/{key}", publish)
	mux.HandleFunc("POST /{plateID}/pubsub/{key}/publish", publish)
	mux.HandleFunc("GET /{plateID}/subscribe/{key}", subscribe)
	mux.HandleFunc("GET /{plateID}/pubsub/{key}/subscribe", subscribe)
	mux.HandleFunc("GET /{plateID}/pubsub/{key}/client", createClientSocket)
	mux.HandleFunc("GET /{plateID}/ws/subscribe/{key}", wsSubscribeHandler(deps))
	mux.HandleFunc("GET /{plateID}/pubsub/{key}/ws", wsSubscribeHandler(deps))
	mux.HandleFunc("GET /{plateID}/s/{token}", clientSocketSubscribeHandler(deps, clientSockets))
}

func wsSubscribeHandler(deps *plate.Dependencies) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		plateID := r.PathValue("plateID")
		authKey := plate.NormalizeAuthorizationHeader(r.Header.Get("Authorization"))
		if plateID == "" || authKey == "" {
			plate.WriteError(w, plate.NewAPIError(http.StatusUnauthorized, "missing_authorization", "authorization header is required"))
			return
		}
		authCtx, cancel := context.WithTimeout(r.Context(), deps.Config.RedisOpTimeout)
		defer cancel()
		allowed, err := deps.Manager.Authorize(authCtx, plateID, authKey)
		if err != nil || !allowed {
			if err == nil {
				err = plate.NewAPIError(http.StatusUnauthorized, "invalid_authorization", "authorization denied")
			}
			plate.WriteError(w, err)
			return
		}
		channel, err := plate.PathValue(r, "key")
		if err != nil {
			plate.WriteError(w, err)
			return
		}
		streamPubSubWebSocket(w, r, deps, plateID, channel, plate.QueryBool(r, "pattern"), 0)
	}
}

func clientSocketSubscribeHandler(deps *plate.Dependencies, store *clientSocketStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !websocket.IsWebSocketUpgrade(r) {
			plate.WriteError(w, plate.NewAPIError(http.StatusBadRequest, "websocket_required", "this route only supports websocket connections"))
			return
		}
		plateID, err := plate.PathValue(r, "plateID")
		if err != nil {
			plate.WriteError(w, err)
			return
		}
		token, err := plate.PathValue(r, "token")
		if err != nil {
			plate.WriteError(w, err)
			return
		}
		ticket, err := store.consume(plateID, token)
		if err != nil {
			switch {
			case errors.Is(err, errClientSocketNotFound):
				plate.WriteError(w, plate.NewAPIError(http.StatusUnauthorized, "invalid_client_socket", "client socket token is invalid"))
			case errors.Is(err, errClientSocketExpired):
				plate.WriteError(w, plate.NewAPIError(http.StatusGone, "expired_client_socket", "client socket token has expired"))
			case errors.Is(err, errClientSocketExhausted):
				plate.WriteError(w, plate.NewAPIError(http.StatusGone, "exhausted_client_socket", "client socket token has no uses left"))
			default:
				plate.WriteError(w, err)
			}
			return
		}
		streamPubSubWebSocket(w, r, deps, ticket.PlateID, ticket.Channel, ticket.Pattern, ticket.MaxDurMS)
	}
}

func streamPubSubWebSocket(w http.ResponseWriter, r *http.Request, deps *plate.Dependencies, plateID string, channel string, pattern bool, maxDurMS int64) {
	conn, err := pubSubUpgrader.Upgrade(w, r, nil)
	if err != nil {
		plate.WriteError(w, err)
		return
	}
	defer conn.Close()

	streamCtx := r.Context()
	cancel := func() {}
	if maxDurMS > 0 {
		streamCtx, cancel = context.WithTimeout(streamCtx, time.Duration(maxDurMS)*time.Millisecond)
	}
	defer cancel()

	pubsub := deps.PubSub.Subscribe(streamCtx, plate.PrefixKey(plateID, channel))
	if pattern {
		_ = pubsub.Close()
		pubsub = deps.PubSub.PSubscribe(streamCtx, plate.PrefixPattern(plateID, channel))
	}
	defer pubsub.Close()
	if _, err := pubsub.Receive(streamCtx); err != nil {
		return
	}
	messageCh := pubsub.Channel(redis.WithChannelSize(deps.Config.PubSubBufferSize))
	for {
		select {
		case <-streamCtx.Done():
			return
		case message, ok := <-messageCh:
			if !ok {
				return
			}
			payload := map[string]any{
				"channel": plate.UnprefixKey(plateID, message.Channel),
				"message": decodePubSubPayload(message.Payload),
			}
			if err := conn.WriteJSON(payload); err != nil {
				return
			}
		}
	}
}

func parseClientSocketExpiry(r *http.Request) (int64, error) {
	query := r.URL.Query()
	raw := strings.TrimSpace(query.Get("expiry"))
	if raw == "" {
		raw = strings.TrimSpace(query.Get("ttl"))
	}
	if raw == "" {
		return clientSocketDefaultExpiry, nil
	}
	value, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		return 0, plate.NewAPIError(http.StatusBadRequest, "invalid_query", "invalid expiry/ttl")
	}
	if value < 0 || value > clientSocketMaxExpiry {
		return 0, plate.NewAPIError(http.StatusBadRequest, "invalid_query", "expiry/ttl must be between 0 and 600000")
	}
	return value, nil
}

func decodePubSubPayload(payload string) any {
	var decoded any
	if err := json.Unmarshal([]byte(payload), &decoded); err == nil {
		return decoded
	}
	return payload
}
