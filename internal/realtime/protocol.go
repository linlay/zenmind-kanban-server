package realtime

import "encoding/json"

const ProtocolVersion = 1

type Envelope struct {
	V         int             `json:"v"`
	Type      string          `json:"type"`
	ID        string          `json:"id,omitempty"`
	Op        string          `json:"op,omitempty"`
	Role      string          `json:"role,omitempty"`
	BoardID   string          `json:"boardId,omitempty"`
	ProjectID string          `json:"projectId,omitempty"`
	Revision  int64           `json:"revision,omitempty"`
	Payload   json.RawMessage `json:"payload,omitempty"`
	OK        *bool           `json:"ok,omitempty"`
	Error     *ErrorPayload   `json:"error,omitempty"`
}

type OutEnvelope struct {
	V         int           `json:"v"`
	Type      string        `json:"type"`
	ID        string        `json:"id,omitempty"`
	Op        string        `json:"op,omitempty"`
	Role      string        `json:"role,omitempty"`
	BoardID   string        `json:"boardId,omitempty"`
	ProjectID string        `json:"projectId,omitempty"`
	Revision  int64         `json:"revision,omitempty"`
	Payload   any           `json:"payload,omitempty"`
	OK        *bool         `json:"ok,omitempty"`
	Error     *ErrorPayload `json:"error,omitempty"`
}

type ErrorPayload struct {
	Message string `json:"message"`
	Code    string `json:"code,omitempty"`
}

func boolPtr(value bool) *bool {
	return &value
}
