package realtime

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/gorilla/websocket"

	"zenmind-kanban-server/internal/kanban"
)

type Session struct {
	id           string
	role         string
	board        string
	projectID    string
	deviceID     string
	currentUser  DesktopCurrentUser
	scope        string
	capabilities []string
	agents       []kanban.DesktopAgentOption
	lastSeenAt   time.Time
	conn         *websocket.Conn
	hub          *Hub
	send         chan OutEnvelope
	closed       chan struct{}
}

type DesktopCurrentUser struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Email  string `json:"email"`
	Source string `json:"source"`
}

func (s *Session) actorID() string {
	if s.role == "desktop" {
		if userID := strings.TrimSpace(s.currentUser.ID); userID != "" {
			return userID
		}
	}
	return s.role
}

func (s *Session) readPump() {
	defer func() {
		s.hub.unregister(s)
		_ = s.conn.Close()
	}()
	s.conn.SetReadLimit(maxMessageSize)
	_ = s.conn.SetReadDeadline(time.Now().Add(pongWait))
	s.conn.SetPongHandler(func(string) error {
		_ = s.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})
	for {
		var env Envelope
		if err := s.conn.ReadJSON(&env); err != nil {
			return
		}
		if env.V == 0 {
			env.V = ProtocolVersion
		}
		if env.BoardID == "" {
			env.BoardID = s.board
		}
		if env.ProjectID == "" {
			env.ProjectID = s.projectID
		}
		if s.role == "desktop" {
			s.lastSeenAt = time.Now().UTC()
		}
		s.hub.handle(s, env)
	}
}

func (s *Session) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		_ = s.conn.Close()
	}()
	for {
		select {
		case message, ok := <-s.send:
			_ = s.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				_ = s.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			if err := s.conn.WriteJSON(message); err != nil {
				return
			}
		case <-ticker.C:
			_ = s.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := s.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func (s *Session) sendSnapshot() {
	env := Envelope{ID: randomID(), Op: "kanban.snapshot", BoardID: kanban.DefaultBoardID, ProjectID: s.projectID}
	s.sendSnapshotFor(env)
}

func (s *Session) enqueue(env OutEnvelope) (ok bool) {
	defer func() {
		if recover() != nil {
			ok = false
			s.hub.unregister(s)
		}
	}()
	select {
	case s.send <- env:
		return true
	default:
		s.hub.unregister(s)
		return false
	}
}

func (s *Session) sendSnapshotFor(env Envelope) {
	projectID := env.ProjectID
	var payload struct {
		ProjectID string `json:"projectId"`
	}
	if len(env.Payload) > 0 {
		_ = json.Unmarshal(env.Payload, &payload)
	}
	if payload.ProjectID != "" {
		projectID = payload.ProjectID
	}
	snapshot, err := s.hub.snapshotForSession(s, env.BoardID, projectID)
	if err != nil {
		s.respondError(env, "server_error", err.Error())
		return
	}
	s.projectID = snapshot.ProjectID
	outType := "event"
	if protocolOp(env.Op) == "snapshot.get" {
		outType = "rpc.res"
	}
	s.enqueue(OutEnvelope{
		V:         ProtocolVersion,
		Type:      outType,
		ID:        env.ID,
		Op:        "kanban.snapshot",
		Role:      "server",
		BoardID:   snapshot.BoardID,
		ProjectID: snapshot.ProjectID,
		Revision:  snapshot.Revision,
		OK:        boolPtr(true),
		Payload:   snapshot,
	})
}

func (s *Session) respond(env Envelope, payload any) {
	s.enqueue(OutEnvelope{
		V:         ProtocolVersion,
		Type:      "rpc.res",
		ID:        env.ID,
		Op:        env.Op,
		Role:      "server",
		BoardID:   env.BoardID,
		ProjectID: env.ProjectID,
		OK:        boolPtr(true),
		Payload:   payload,
	})
}

func (s *Session) respondError(env Envelope, code string, message string) {
	if message == "" {
		message = "操作失败。"
	}
	payload, _ := json.Marshal(map[string]any{
		"ok":      false,
		"message": message,
	})
	s.enqueue(OutEnvelope{
		V:         ProtocolVersion,
		Type:      "rpc.res",
		ID:        env.ID,
		Op:        env.Op,
		Role:      "server",
		BoardID:   env.BoardID,
		ProjectID: env.ProjectID,
		OK:        boolPtr(false),
		Error:     &ErrorPayload{Code: code, Message: message},
		Payload:   json.RawMessage(payload),
	})
}
