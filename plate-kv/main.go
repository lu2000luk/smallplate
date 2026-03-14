package main

import (
	"encoding/json"
	"net/url"
	"os"
	"os/signal"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"
	"github.com/joho/godotenv"
)

const Version = "0.1.0"

const (
	pingInterval      = 30 * time.Second
	pongMissThreshold = 30 * time.Second
	writeWait         = 10 * time.Second
	retryDelay        = 1 * time.Second
	maxDialRetries    = 20
)

type PingPacket struct {
	Type string `json:"type"`
	Time int64  `json:"time,omitempty"`
}

func main() {
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	err := godotenv.Load()
	if err != nil {
		Warn("Error loading .env file (ignore if vars are set in the env)")
	}

	db_url := os.Getenv("DB_URL")
	service_id := os.Getenv("SERVICE_ID")
	service_key := os.Getenv("SERVICE_KEY")
	manager_url := os.Getenv("MANAGER_URL")

	if db_url == "" || service_id == "" || service_key == "" || manager_url == "" {
		Error("Missing required environment variables. Required variables: DB_URL, SERVICE_ID, SERVICE_KEY, MANAGER_URL")
		Log("Variables configuration guide:")
		Log("DB_URL: The URL of the Redis/Valkey (or compatible)")
		Log("SERVICE_ID: The unique identifier for this service (do not change this ever to avoid data loss)")
		Log("SERVICE_KEY: The secret key to connect to the manager")
		Log("MANAGER_URL: The URL of the manager service (e.g. manager.example.com), no protocol, no path")
		return
	}

	u := url.URL{
		Scheme: "ws",
		Host:   manager_url,
		Path:   "/__service",
		RawQuery: url.Values{
			"id": []string{service_id},
			"t":  []string{"kv"},
			"k":  []string{service_key},
		}.Encode(),
	}
	u_masked := url.URL{
		Scheme: "ws",
		Host:   u.Host,
		Path:   u.Path,
		RawQuery: url.Values{
			"id": []string{service_id},
			"t":  []string{"kv"},
			"k":  []string{"REDACTED"},
		}.Encode(),
	}

	Log("Manager endpoint:", u_masked.String())

	for {
		select {
		case <-interrupt:
			Log("Interrupt received, shutting down")
			return
		default:
		}

		conn, ok := connectWithRetry(u.String(), u_masked.String(), interrupt)
		if !ok {
			return
		}

		reconnect := runConnection(conn, interrupt)
		if !reconnect {
			return
		}

		Log("Reconnecting...")
	}
}

func connectWithRetry(rawURL string, maskedURL string, interrupt <-chan os.Signal) (*websocket.Conn, bool) {
	for attempt := 1; attempt <= maxDialRetries; attempt++ {
		Log("Connecting to", maskedURL, "( attempt", attempt, "of", maxDialRetries, ")")
		c, _, err := websocket.DefaultDialer.Dial(rawURL, nil)
		if err == nil {
			Log("Connected to manager")
			return c, true
		}

		Error("Dial failed:", err)

		if attempt >= maxDialRetries {
			Error("Giving up after", maxDialRetries, "failed attempts")
			return nil, false
		}

		Log("Reconnect in 1s...")
		select {
		case <-interrupt:
			Log("Interrupt received during backoff, shutting down")
			return nil, false
		case <-time.After(retryDelay):
		}
	}

	return nil, false
}

func runConnection(c *websocket.Conn, interrupt <-chan os.Signal) bool {
	defer c.Close()

	var writeMu sync.Mutex
	var closed int32

	var lastPingSent int64
	atomic.StoreInt64(&lastPingSent, 0)

	closeAndMark := func(reason string) {
		if atomic.CompareAndSwapInt32(&closed, 0, 1) {
			Warn("Closing socket:", reason)
			_ = c.Close()
		}
	}

	readDone := make(chan struct{})
	go func() {
		defer close(readDone)

		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				if atomic.LoadInt32(&closed) == 0 {
					Warn("Read loop ended:", err)
				}
				return
			}

			var packet PingPacket
			if err := json.Unmarshal(message, &packet); err != nil {
				continue
			}

			switch packet.Type {
			case "ping":
				resp := PingPacket{
					Type: "pong",
					Time: packet.Time,
				}
				payload, err := json.Marshal(resp)
				if err != nil {
					continue
				}

				writeMu.Lock()
				_ = c.SetWriteDeadline(time.Now().Add(writeWait))
				err = c.WriteMessage(websocket.TextMessage, payload)
				_ = c.SetWriteDeadline(time.Time{})
				writeMu.Unlock()
				if err != nil {
					Warn("Failed to write pong:", err)
					closeAndMark("ping write failure")
					return
				}
			case "pong":
				lastSent := atomic.LoadInt64(&lastPingSent)
				if lastSent != 0 {
					atomic.StoreInt64(&lastPingSent, 0)
				}
			}
		}
	}()

	pingTicker := time.NewTicker(pingInterval)
	defer pingTicker.Stop()

	watchdogTicker := time.NewTicker(1 * time.Second)
	defer watchdogTicker.Stop()

	for {
		select {
		case <-interrupt:
			Log("Interrupt received, closing connection")
			closeAndMark("interrupt")
			<-readDone
			return false

		case <-readDone:
			return true

		case <-pingTicker.C:
			if atomic.LoadInt64(&lastPingSent) != 0 {
				Warn("Previous ping still awaiting pong, forcing reconnect")
				closeAndMark("missed pong")
				<-readDone
				return true
			}

			ping := PingPacket{
				Type: "ping",
				Time: time.Now().UnixMilli(),
			}
			payload, err := json.Marshal(ping)
			if err != nil {
				Warn("Failed to encode ping:", err)
				continue
			}

			writeMu.Lock()
			_ = c.SetWriteDeadline(time.Now().Add(writeWait))
			err = c.WriteMessage(websocket.TextMessage, payload)
			_ = c.SetWriteDeadline(time.Time{})
			writeMu.Unlock()
			if err != nil {
				Warn("Failed to send ping:", err)
				closeAndMark("ping write failure")
				<-readDone
				return true
			}

			atomic.StoreInt64(&lastPingSent, time.Now().UnixNano())

		case <-watchdogTicker.C:
			lastSent := atomic.LoadInt64(&lastPingSent)
			if lastSent == 0 {
				continue
			}

			last := time.Unix(0, lastSent)
			if time.Since(last) > pongMissThreshold {
				Warn("Missed pong for over 30s, forcing reconnect")
				closeAndMark("missed pong")
				<-readDone
				return true
			}
		}
	}
}
