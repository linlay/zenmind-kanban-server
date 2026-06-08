package desktop

const (
	CapabilityAssistant   = "assistant"
	CapabilityAutomation  = "automation"
	CapabilityAttachments = "attachments"
)

type HelloPayload struct {
	Capabilities      []string `json:"capabilities"`
	SelectedProjectID string   `json:"selectedProjectId"`
}
