package plate

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"
)

type managerCheckRequest struct {
	Type      string `json:"type"`
	RequestID string `json:"request_id"`
	PlateID   string `json:"plate_id"`
	Key       string `json:"key"`
	Service   string `json:"service"`
}

type managerCheckResponse struct {
	Type      string `json:"type"`
	RequestID string `json:"request_id"`
	PlateID   string `json:"plate_id"`
	Key       string `json:"key"`
	Valid     bool   `json:"valid"`
}

type managerInvalidateEvent struct {
	Type string `json:"type"`
	Key  string `json:"key"`
}

type managerDeleteEvent struct {
	Type string `json:"type"`
	ID   int    `json:"id"`
}

type managerPingPacket struct {
	Type string `json:"type"`
	Time int64  `json:"time,omitempty"`
}

type managerResult struct {
	decision AuthDecision
	err      error
}

type pendingAuthRequest struct {
	plateID string
	key     string
	ch      chan managerResult
}

type ManagerClient struct {
	cfg       Config
	authCache *AuthCache
	meta      *MetaStore

	connMu  sync.RWMutex
	conn    *websocket.Conn
	writeMu sync.Mutex
	closed  atomic.Int32

	pendingMu sync.Mutex
	pending   map[string]pendingAuthRequest
	requestID atomic.Uint64
}

func NewManagerClient(cfg Config, authCache *AuthCache, meta *MetaStore) *ManagerClient {
	return &ManagerClient{
		cfg:       cfg,
		authCache: authCache,
		meta:      meta,
		pending:   make(map[string]pendingAuthRequest),
	}
}

func (m *ManagerClient) Run(ctx context.Context) {
	managerURL, maskedURL, err := m.cfg.ManagerWSURL()
	if err != nil {
		Error("failed to build manager url:", err)
		return
	}

	Log("Manager endpoint:", maskedURL)
	for {
		if ctx.Err() != nil {
			m.closeConnection("shutdown")
			m.failPending(errors.New("manager shutdown"))
			return
		}

		conn, ok := m.connectWithRetry(ctx, managerURL, maskedURL)
		if !ok {
			return
		}

		reconnect := m.runConnection(ctx, conn)
		if !reconnect {
			return
		}

		select {
		case <-ctx.Done():
			return
		case <-time.After(m.cfg.ReconnectDelay):
		}
	}
}

func (m *ManagerClient) Authorize(ctx context.Context, plateID string, key string) (bool, error) {
	if decision, ok := m.authCache.Get(plateID, key, ServiceType); ok {
		return decision.Valid, nil
	}

	requestID := m.nextRequestID()
	resultCh := make(chan managerResult, 1)
	m.pendingMu.Lock()
	m.pending[requestID] = pendingAuthRequest{plateID: plateID, key: key, ch: resultCh}
	m.pendingMu.Unlock()

	request := managerCheckRequest{
		Type:      "check",
		RequestID: requestID,
		PlateID:   plateID,
		Key:       key,
		Service:   ServiceType,
	}

	if err := m.writeJSON(request); err != nil {
		m.pendingMu.Lock()
		delete(m.pending, requestID)
		m.pendingMu.Unlock()
		return false, err
	}

	select {
	case <-ctx.Done():
		m.pendingMu.Lock()
		delete(m.pending, requestID)
		m.pendingMu.Unlock()
		return false, ctx.Err()
	case result := <-resultCh:
		if result.err != nil {
			return false, result.err
		}
		m.authCache.Add(result.decision)
		return result.decision.Valid, nil
	}
}

func (m *ManagerClient) connectWithRetry(ctx context.Context, rawURL string, maskedURL string) (*websocket.Conn, bool) {
	for attempt := 1; attempt <= m.cfg.MaxDialRetries; attempt++ {
		Log("Connecting to", maskedURL, "( attempt", attempt, "of", m.cfg.MaxDialRetries, ")")
		conn, _, err := websocket.DefaultDialer.DialContext(ctx, rawURL, nil)
		if err == nil {
			Log("Connected to manager")
			return conn, true
		}
		Error("Dial failed:", err)
		if attempt >= m.cfg.MaxDialRetries {
			Error("Giving up after", m.cfg.MaxDialRetries, "failed attempts")
			return nil, false
		}
		select {
		case <-ctx.Done():
			return nil, false
		case <-time.After(m.cfg.ReconnectDelay):
		}
	}
	return nil, false
}

func (m *ManagerClient) runConnection(ctx context.Context, conn *websocket.Conn) bool {
	m.connMu.Lock()
	m.conn = conn
	m.connMu.Unlock()
	m.closed.Store(0)

	defer func() {
		m.closeConnection("reconnect")
		m.failPending(errors.New("manager connection dropped"))
	}()

	var lastPingSent atomic.Int64
	readDone := make(chan struct{})

	go func() {
		defer close(readDone)
		for {
			_, message, err := conn.ReadMessage()
			if err != nil {
				if m.closed.Load() == 0 {
					Warn("Read loop ended:", err)
				}
				return
			}

			var probe struct {
				Type string `json:"type"`
			}
			if err := json.Unmarshal(message, &probe); err != nil {
				Warn("failed to decode manager packet:", err)
				continue
			}

			switch probe.Type {
			case "ping":
				var packet managerPingPacket
				if err := json.Unmarshal(message, &packet); err != nil {
					continue
				}
				_ = m.writeJSON(managerPingPacket{Type: "pong", Time: packet.Time})
			case "pong":
				lastPingSent.Store(0)
			case "create":
				if err := m.respondToCreate(message); err != nil {
					Warn("failed to respond to create event:", err)
				}
			case "check_response":
				var response managerCheckResponse
				if err := json.Unmarshal(message, &response); err != nil {
					Warn("failed to decode check response:", err)
					continue
				}
				m.completePending(response)
			case "invalidate":
				var event managerInvalidateEvent
				if err := json.Unmarshal(message, &event); err != nil {
					Warn("failed to decode invalidate event:", err)
					continue
				}
				m.authCache.Invalidate(event.Key)
			case "delete":
				var event managerDeleteEvent
				if err := json.Unmarshal(message, &event); err != nil {
					Warn("failed to decode delete event:", err)
					continue
				}
				m.authCache.InvalidatePlate(strconv.Itoa(event.ID))
				if err := m.meta.DeletePlate(strconv.Itoa(event.ID)); err != nil {
					Warn("failed to delete local plate metadata:", err)
				}
			default:
				Warn("Ignoring unsupported manager event:", probe.Type)
			}
		}
	}()

	pingTicker := time.NewTicker(m.cfg.PingInterval)
	defer pingTicker.Stop()
	watchdogTicker := time.NewTicker(time.Second)
	defer watchdogTicker.Stop()

	for {
		select {
		case <-ctx.Done():
			m.closeConnection("shutdown")
			<-readDone
			return false
		case <-readDone:
			return true
		case <-pingTicker.C:
			if lastPingSent.Load() != 0 {
				Warn("Previous ping still awaiting pong, forcing reconnect")
				m.closeConnection("missed pong")
				<-readDone
				return true
			}
			if err := m.writeJSON(managerPingPacket{Type: "ping", Time: time.Now().UnixMilli()}); err != nil {
				Warn("Failed to send ping:", err)
				m.closeConnection("ping write failure")
				<-readDone
				return true
			}
			lastPingSent.Store(time.Now().UnixNano())
		case <-watchdogTicker.C:
			last := lastPingSent.Load()
			if last == 0 {
				continue
			}
			if time.Since(time.Unix(0, last)) > m.cfg.PongMissTimeout {
				Warn("Missed pong for too long, forcing reconnect")
				m.closeConnection("missed pong")
				<-readDone
				return true
			}
		}
	}
}

func (m *ManagerClient) writeJSON(payload any) error {
	conn := m.currentConn()
	if conn == nil {
		return errors.New("manager connection is not ready")
	}
	message, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	m.writeMu.Lock()
	defer m.writeMu.Unlock()
	if err := conn.SetWriteDeadline(time.Now().Add(m.cfg.WriteTimeout)); err != nil {
		return err
	}
	err = conn.WriteMessage(websocket.TextMessage, message)
	_ = conn.SetWriteDeadline(time.Time{})
	return err
}

func (m *ManagerClient) currentConn() *websocket.Conn {
	m.connMu.RLock()
	defer m.connMu.RUnlock()
	return m.conn
}

func (m *ManagerClient) closeConnection(reason string) {
	if !m.closed.CompareAndSwap(0, 1) {
		return
	}
	Warn("Closing manager socket:", reason)
	m.connMu.Lock()
	conn := m.conn
	m.conn = nil
	m.connMu.Unlock()
	if conn != nil {
		_ = conn.Close()
	}
}

func (m *ManagerClient) failPending(err error) {
	m.pendingMu.Lock()
	defer m.pendingMu.Unlock()
	for requestID, ch := range m.pending {
		ch.ch <- managerResult{err: err}
		close(ch.ch)
		delete(m.pending, requestID)
	}
}

func (m *ManagerClient) completePending(response managerCheckResponse) {
	m.pendingMu.Lock()
	pending, ok := m.pending[response.RequestID]
	if ok {
		delete(m.pending, response.RequestID)
	}
	m.pendingMu.Unlock()
	if !ok {
		Warn("Received orphaned check response:", response.RequestID)
		return
	}
	if pending.plateID != response.PlateID || pending.key != response.Key {
		pending.ch <- managerResult{err: fmt.Errorf("manager response mismatch for request %s", response.RequestID)}
		close(pending.ch)
		return
	}
	decision := AuthDecision{
		PlateID: response.PlateID,
		Key:     response.Key,
		Service: ServiceType,
		Valid:   response.Valid,
	}
	pending.ch <- managerResult{decision: decision}
	close(pending.ch)
}

func (m *ManagerClient) nextRequestID() string {
	value := m.requestID.Add(1)
	return fmt.Sprintf("%s-%d-%d", m.cfg.ServiceID, time.Now().UnixNano(), value)
}

func (m *ManagerClient) respondToCreate(message []byte) error {
	var payload map[string]json.RawMessage
	if err := json.Unmarshal(message, &payload); err != nil {
		return err
	}
	payload["type"] = json.RawMessage(`"created"`)
	encoded, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	return m.writeRaw(encoded)
}

func (m *ManagerClient) writeRaw(message []byte) error {
	conn := m.currentConn()
	if conn == nil {
		return errors.New("manager connection is not ready")
	}
	m.writeMu.Lock()
	defer m.writeMu.Unlock()
	if err := conn.SetWriteDeadline(time.Now().Add(m.cfg.WriteTimeout)); err != nil {
		return err
	}
	err := conn.WriteMessage(websocket.TextMessage, message)
	_ = conn.SetWriteDeadline(time.Time{})
	return err
}
