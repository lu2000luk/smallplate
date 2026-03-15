// Endpoints contained in this file:
// POST /{plateID}/publish/{channel}
// GET /{plateID}/subscribe/{channel}
// GET /{plateID}/ws/subscribe/{channel}
package routes

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"github.com/redis/go-redis/v9"

	"plain/kv/internal/plate"
)

func registerPubSub(mux *http.ServeMux, deps *plate.Dependencies) {
	mux.HandleFunc("POST /{plateID}/publish/{key}", plate.Authenticated(deps, func(w http.ResponseWriter, r *http.Request, plateID string) error {
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
	}))
	mux.HandleFunc("GET /{plateID}/subscribe/{key}", plate.Authenticated(deps, func(w http.ResponseWriter, r *http.Request, plateID string) error {
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
	}))
	mux.HandleFunc("GET /{plateID}/ws/subscribe/{key}", wsSubscribeHandler(deps))
}

func wsSubscribeHandler(deps *plate.Dependencies) http.HandlerFunc {
	upgrader := websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}
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
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			plate.WriteError(w, err)
			return
		}
		defer conn.Close()
		pubsub := deps.PubSub.Subscribe(r.Context(), plate.PrefixKey(plateID, channel))
		if plate.QueryBool(r, "pattern") {
			_ = pubsub.Close()
			pubsub = deps.PubSub.PSubscribe(r.Context(), plate.PrefixPattern(plateID, channel))
		}
		defer pubsub.Close()
		if _, err := pubsub.Receive(r.Context()); err != nil {
			return
		}
		messageCh := pubsub.Channel(redis.WithChannelSize(deps.Config.PubSubBufferSize))
		for {
			select {
			case <-r.Context().Done():
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
}

func decodePubSubPayload(payload string) any {
	var decoded any
	if err := json.Unmarshal([]byte(payload), &decoded); err == nil {
		return decoded
	}
	return payload
}
