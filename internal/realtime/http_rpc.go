package realtime

import (
	"context"
	"errors"

	"zenmind-kanban-server/internal/kanban"
)

// HandleHTTPRPC runs one WebSocket RPC envelope through the same hub handler
// used by browser clients. It is a fallback for browsers whose WebSocket path
// is blocked by local proxy settings.
func (h *Hub) HandleHTTPRPC(ctx context.Context, env Envelope) (OutEnvelope, error) {
	if env.V == 0 {
		env.V = ProtocolVersion
	}
	if env.Type == "" {
		env.Type = "rpc.req"
	}
	if env.ID == "" {
		env.ID = randomID()
	}
	if env.Role == "" {
		env.Role = "web"
	}
	if env.BoardID == "" {
		env.BoardID = kanban.DefaultBoardID
	}
	if env.ProjectID == "" {
		env.ProjectID = kanban.DefaultProjectID
	}

	session := &Session{
		id:        randomID(),
		role:      "web",
		hub:       h,
		send:      make(chan OutEnvelope, 8),
		board:     env.BoardID,
		projectID: env.ProjectID,
		closed:    make(chan struct{}),
	}

	done := make(chan struct{})
	go func() {
		defer close(done)
		h.handle(session, env)
	}()

	select {
	case out := <-session.send:
		return out, nil
	case <-done:
		select {
		case out := <-session.send:
			return out, nil
		default:
			return OutEnvelope{}, errors.New("RPC 没有返回结果。")
		}
	case <-ctx.Done():
		return OutEnvelope{}, ctx.Err()
	}
}
