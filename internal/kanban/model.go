package kanban

import "time"

const DefaultBoardID = "default"
const DefaultProjectID = "default"
const DefaultWorkflowID = "workflow-standard-requirement"
const DefaultWorkflowKey = "standard_requirement"

type Status string

const (
	StatusBacklog    Status = "backlog"
	StatusTodo       Status = "todo"
	StatusInProgress Status = "in_progress"
	StatusInReview   Status = "in_review"
	StatusCompleted  Status = "completed"
)

type Priority string

const (
	PriorityHigh   Priority = "high"
	PriorityMedium Priority = "medium"
	PriorityLow    Priority = "low"
)

type Severity string

const (
	SeverityCritical Severity = "critical"
	SeverityHigh     Severity = "high"
	SeverityMedium   Severity = "medium"
	SeverityLow      Severity = "low"
)

type RunState string

const (
	RunStateRunning   RunState = "running"
	RunStateCompleted RunState = "completed"
	RunStateFailed    RunState = "failed"
	RunStateCancelled RunState = "cancelled"
)

type Attachment map[string]any

type Project struct {
	ID                string     `json:"id"`
	ParentID          *string    `json:"parentId"`
	Slug              string     `json:"slug"`
	Key               string     `json:"key,omitempty"`
	Name              string     `json:"name"`
	Description       string     `json:"description,omitempty"`
	Path              string     `json:"path"`
	Depth             int        `json:"depth"`
	Position          float64    `json:"position"`
	Visibility        string     `json:"visibility,omitempty"`
	DefaultWorkflowID string     `json:"defaultWorkflowId,omitempty"`
	ArchivedAt        *time.Time `json:"archivedAt,omitempty"`
	CreatedAt         time.Time  `json:"createdAt"`
	UpdatedAt         time.Time  `json:"updatedAt"`
	DeletedAt         *time.Time `json:"deletedAt,omitempty"`
	CreatedBy         *string    `json:"createdBy,omitempty"`
	UpdatedBy         *string    `json:"updatedBy,omitempty"`
}

type ProjectIssueStat struct {
	ProjectID            string `json:"projectId"`
	IssueCount           int    `json:"issueCount"`
	InProgressIssueCount int    `json:"inProgressIssueCount"`
}

type ProjectInput struct {
	ParentID          *string  `json:"parentId"`
	Name              string   `json:"name"`
	Slug              *string  `json:"slug"`
	Description       *string  `json:"description"`
	Visibility        *string  `json:"visibility"`
	DefaultWorkflowID *string  `json:"defaultWorkflowId"`
	Position          *float64 `json:"position"`
}

type ProjectUpdateInput struct {
	Name              *string `json:"name"`
	Slug              *string `json:"slug"`
	Description       *string `json:"description"`
	Visibility        *string `json:"visibility"`
	DefaultWorkflowID *string `json:"defaultWorkflowId"`
}

type ProjectMoveInput struct {
	ID       string   `json:"id"`
	ParentID *string  `json:"parentId"`
	Position *float64 `json:"position"`
}

type Issue struct {
	BoardID            string       `json:"boardId"`
	ProjectID          string       `json:"projectId"`
	ProjectPath        string       `json:"projectPath,omitempty"`
	ProjectName        string       `json:"projectName,omitempty"`
	WorkflowID         string       `json:"workflowId"`
	StageID            string       `json:"stageId,omitempty"`
	StageKey           string       `json:"stageKey,omitempty"`
	StageName          string       `json:"stageName,omitempty"`
	StatusID           string       `json:"statusId,omitempty"`
	StatusKey          string       `json:"statusKey,omitempty"`
	StatusName         string       `json:"statusName,omitempty"`
	ColumnKey          string       `json:"columnKey,omitempty"`
	ID                 string       `json:"id"`
	Title              string       `json:"title"`
	Description        string       `json:"description"`
	Status             Status       `json:"status"`
	Priority           Priority     `json:"priority"`
	Severity           Severity     `json:"severity"`
	AssigneeAgentKey   *string      `json:"assigneeAgentKey"`
	AssigneeID         *string      `json:"assigneeId"`
	WorkerType         *string      `json:"workerType"`
	WorkerID           *string      `json:"workerId"`
	WorkerAgent        *string      `json:"workerAgent"`
	ReviewerID         *string      `json:"reviewerId"`
	ReviewRequired     bool         `json:"reviewRequired"`
	ActiveReviewID     *string      `json:"activeReviewId"`
	ActiveRunID        *string      `json:"activeRunId"`
	Position           float64      `json:"position"`
	ChatID             *string      `json:"chatId"`
	RunID              *string      `json:"runId"`
	RunState           *RunState    `json:"runState"`
	AutomationID       *string      `json:"automationId"`
	AutomationEnabled  bool         `json:"automationEnabled"`
	AutomationCron     *string      `json:"automationCron"`
	AutomationMessage  *string      `json:"automationMessage"`
	AutomationTimezone *string      `json:"automationTimezone"`
	AttachmentChatID   *string      `json:"attachmentChatId"`
	Attachments        []Attachment `json:"attachments"`
	CreatedAt          time.Time    `json:"createdAt"`
	UpdatedAt          time.Time    `json:"updatedAt"`
	Revision           int64        `json:"revision"`
	DeletedAt          *time.Time   `json:"deletedAt,omitempty"`
	CreatedBy          *string      `json:"createdBy"`
	UpdatedBy          *string      `json:"updatedBy"`
	CreatedByAgent     *string      `json:"createdByAgent"`
	UpdatedByAgent     *string      `json:"updatedByAgent"`
}

type IssueInput struct {
	Title              string       `json:"title"`
	ProjectID          *string      `json:"projectId"`
	WorkflowID         *string      `json:"workflowId"`
	StageID            *string      `json:"stageId"`
	StatusID           *string      `json:"statusId"`
	Description        *string      `json:"description"`
	Status             *string      `json:"status"`
	Priority           *string      `json:"priority"`
	Severity           *string      `json:"severity"`
	AssigneeAgentKey   *string      `json:"assigneeAgentKey"`
	AssigneeID         *string      `json:"assigneeId"`
	WorkerType         *string      `json:"workerType"`
	WorkerID           *string      `json:"workerId"`
	WorkerAgent        *string      `json:"workerAgent"`
	ReviewerID         *string      `json:"reviewerId"`
	ReviewRequired     *bool        `json:"reviewRequired"`
	RunState           *string      `json:"runState"`
	AutomationID       *string      `json:"automationId"`
	AutomationEnabled  *bool        `json:"automationEnabled"`
	AutomationCron     *string      `json:"automationCron"`
	AutomationMessage  *string      `json:"automationMessage"`
	AutomationTimezone *string      `json:"automationTimezone"`
	AttachmentChatID   *string      `json:"attachmentChatId"`
	Attachments        []Attachment `json:"attachments"`
}

type IssueUpdateInput struct {
	Title              *string      `json:"title"`
	ProjectID          *string      `json:"projectId"`
	WorkflowID         *string      `json:"workflowId"`
	StageID            *string      `json:"stageId"`
	StatusID           *string      `json:"statusId"`
	Description        *string      `json:"description"`
	Status             *string      `json:"status"`
	Priority           *string      `json:"priority"`
	Severity           *string      `json:"severity"`
	AssigneeAgentKey   *string      `json:"assigneeAgentKey"`
	AssigneeID         *string      `json:"assigneeId"`
	WorkerType         *string      `json:"workerType"`
	WorkerID           *string      `json:"workerId"`
	WorkerAgent        *string      `json:"workerAgent"`
	ReviewerID         *string      `json:"reviewerId"`
	ReviewRequired     *bool        `json:"reviewRequired"`
	ChatID             *string      `json:"chatId"`
	RunID              *string      `json:"runId"`
	RunState           *string      `json:"runState"`
	AutomationID       *string      `json:"automationId"`
	AutomationEnabled  *bool        `json:"automationEnabled"`
	AutomationCron     *string      `json:"automationCron"`
	AutomationMessage  *string      `json:"automationMessage"`
	AutomationTimezone *string      `json:"automationTimezone"`
	AttachmentChatID   *string      `json:"attachmentChatId"`
	Attachments        []Attachment `json:"attachments"`
	BaseIssueRevision  *int64       `json:"baseIssueRevision"`
}

type MoveInput struct {
	ID                string  `json:"id"`
	Status            string  `json:"status"`
	Position          float64 `json:"position"`
	BaseIssueRevision *int64  `json:"baseIssueRevision"`
}

type AssignAndRunInput struct {
	ID                     string  `json:"id"`
	AgentKey               *string `json:"agentKey"`
	BaseIssueRevision      *int64  `json:"baseIssueRevision"`
	IdempotencyKey         string  `json:"idempotencyKey"`
	TargetDesktopSessionID string  `json:"targetDesktopSessionId"`
}

type AssistantEvent struct {
	Type   string  `json:"type"`
	Status *string `json:"status"`
	ChatID *string `json:"chatId"`
	RunID  *string `json:"runId"`
}

type StartRunResult struct {
	OK       bool    `json:"ok"`
	Message  string  `json:"message"`
	ChatID   *string `json:"chatId"`
	RunID    *string `json:"runId"`
	AgentKey *string `json:"agentKey"`
}

type DesktopStatus struct {
	Online            bool                   `json:"online"`
	SessionID         string                 `json:"sessionId,omitempty"`
	Capabilities      []string               `json:"capabilities"`
	SelectedProjectID string                 `json:"selectedProjectId,omitempty"`
	Sessions          []DesktopSessionStatus `json:"sessions,omitempty"`
}

type DesktopSessionStatus struct {
	SessionID         string    `json:"sessionId"`
	DeviceID          string    `json:"deviceId,omitempty"`
	CurrentUserID     string    `json:"currentUserId,omitempty"`
	CurrentUserName   string    `json:"currentUserName,omitempty"`
	SelectedProjectID string    `json:"selectedProjectId,omitempty"`
	Capabilities      []string  `json:"capabilities"`
	LastSeenAt        time.Time `json:"lastSeenAt"`
}

type DesktopAgentOption struct {
	AgentKey    string         `json:"agentKey"`
	DisplayName string         `json:"displayName"`
	Role        string         `json:"role,omitempty"`
	Icon        map[string]any `json:"icon,omitempty"`
}

type DesktopOnlineDevice struct {
	DeviceID          string                 `json:"deviceId"`
	CurrentUserID     string                 `json:"currentUserId,omitempty"`
	CurrentUserName   string                 `json:"currentUserName,omitempty"`
	SelectedProjectID string                 `json:"selectedProjectId,omitempty"`
	Capabilities      []string               `json:"capabilities"`
	LastSeenAt        time.Time              `json:"lastSeenAt"`
	Sessions          []DesktopSessionStatus `json:"sessions"`
	Agents            []DesktopAgentOption   `json:"agents"`
	AgentError        string                 `json:"agentError,omitempty"`
}

type DesktopOnlineListResult struct {
	OK           bool                  `json:"ok"`
	Online       bool                  `json:"online"`
	DeviceCount  int                   `json:"deviceCount"`
	SessionCount int                   `json:"sessionCount"`
	AgentCount   int                   `json:"agentCount"`
	Devices      []DesktopOnlineDevice `json:"devices"`
}

type ListResult struct {
	OK                  bool                 `json:"ok"`
	Message             string               `json:"message"`
	BoardID             string               `json:"boardId"`
	ProjectID           string               `json:"projectId"`
	Revision            int64                `json:"revision"`
	Complete            bool                 `json:"complete"`
	Scope               string               `json:"scope"`
	Projects            []Project            `json:"projects"`
	Issues              []Issue              `json:"issues"`
	ProjectIssueStats   []ProjectIssueStat   `json:"projectIssueStats"`
	Users               []UserAccount        `json:"users"`
	Workflows           []Workflow           `json:"workflows"`
	WorkflowStageDefs   []WorkflowStageDef   `json:"workflowStageDefs"`
	WorkflowStatusDefs  []WorkflowStatusDef  `json:"workflowStatusDefs"`
	WorkflowStages      []WorkflowStage      `json:"workflowStages"`
	WorkflowStatuses    []WorkflowStatus     `json:"workflowStatuses"`
	WorkflowTransitions []WorkflowTransition `json:"workflowTransitions"`
	Teams               []Team               `json:"teams"`
	TeamMembers         []TeamMember         `json:"teamMembers"`
	ProjectPermissions  []ProjectPermission  `json:"projectPermissions"`
	IssueLabels         []IssueLabel         `json:"issueLabels"`
	IssueLabelLinks     []IssueLabelLink     `json:"issueLabelLinks"`
	IssueDependencies   []IssueDependency    `json:"issueDependencies"`
	Reviews             []Review             `json:"reviews"`
	ReviewComments      []ReviewComment      `json:"reviewComments"`
	Agents              []Agent              `json:"agents"`
	AgentRuns           []AgentRun           `json:"agentRuns"`
	AgentToolCalls      []AgentToolCall      `json:"agentToolCalls"`
	RecentEvents        []EventLogItem       `json:"recentEvents"`
	DesktopStatus       DesktopStatus        `json:"desktopStatus"`
	StoragePath         string               `json:"storagePath,omitempty"`
}

type IssuesResult struct {
	OK        bool    `json:"ok"`
	Message   string  `json:"message"`
	BoardID   string  `json:"boardId"`
	ProjectID string  `json:"projectId"`
	Revision  int64   `json:"revision"`
	Count     int     `json:"count"`
	Issues    []Issue `json:"issues"`
}

type ChangeResult struct {
	OK                bool               `json:"ok"`
	Code              string             `json:"code,omitempty"`
	Message           string             `json:"message"`
	BoardID           string             `json:"boardId"`
	ProjectID         string             `json:"projectId"`
	Revision          int64              `json:"revision"`
	Complete          bool               `json:"complete"`
	Scope             string             `json:"scope"`
	Issue             *Issue             `json:"issue,omitempty"`
	Issues            []Issue            `json:"issues"`
	ProjectIssueStats []ProjectIssueStat `json:"projectIssueStats,omitempty"`
	DeletedIssueID    string             `json:"deletedIssueId,omitempty"`
}

type ProjectChangeResult struct {
	OK        bool      `json:"ok"`
	Message   string    `json:"message"`
	BoardID   string    `json:"boardId"`
	ProjectID string    `json:"projectId"`
	Revision  int64     `json:"revision"`
	Project   *Project  `json:"project,omitempty"`
	Projects  []Project `json:"projects"`
}

type AgentOption struct {
	AgentKey    string         `json:"agentKey"`
	DisplayName string         `json:"displayName"`
	Role        string         `json:"role,omitempty"`
	Icon        map[string]any `json:"icon,omitempty"`
}

type Workflow struct {
	ID             string     `json:"id"`
	Key            string     `json:"key"`
	Name           string     `json:"name"`
	Description    string     `json:"description"`
	IsDefault      bool       `json:"isDefault"`
	TransitionMode string     `json:"transitionMode"`
	CreatedAt      time.Time  `json:"createdAt"`
	UpdatedAt      time.Time  `json:"updatedAt"`
	DeletedAt      *time.Time `json:"deletedAt,omitempty"`
}

type WorkflowInput struct {
	Key            string `json:"key"`
	Name           string `json:"name"`
	Description    string `json:"description,omitempty"`
	TransitionMode string `json:"transitionMode"`
}

type WorkflowUpdateInput struct {
	Name           *string `json:"name,omitempty"`
	Description    *string `json:"description,omitempty"`
	TransitionMode *string `json:"transitionMode,omitempty"`
}

type WorkflowStageDef struct {
	ID          string     `json:"id"`
	Key         string     `json:"key"`
	Name        string     `json:"name"`
	Description string     `json:"description"`
	IsSystem    bool       `json:"isSystem"`
	CreatedAt   time.Time  `json:"createdAt"`
	UpdatedAt   time.Time  `json:"updatedAt"`
	DeletedAt   *time.Time `json:"deletedAt,omitempty"`
}

type WorkflowStatusDef struct {
	ID          string     `json:"id"`
	Key         string     `json:"key"`
	Name        string     `json:"name"`
	ColumnKey   string     `json:"columnKey"`
	Description string     `json:"description"`
	IsSystem    bool       `json:"isSystem"`
	CreatedAt   time.Time  `json:"createdAt"`
	UpdatedAt   time.Time  `json:"updatedAt"`
	DeletedAt   *time.Time `json:"deletedAt,omitempty"`
}

type WorkflowStage struct {
	ID         string `json:"id"`
	WorkflowID string `json:"workflowId"`
	StageDefID string `json:"stageDefId"`
	Key        string `json:"key"`
	Name       string `json:"name"`
	Position   int    `json:"position"`
	IsStart    bool   `json:"isStart"`
	IsEnd      bool   `json:"isEnd"`
}

type WorkflowStatus struct {
	ID             string `json:"id"`
	WorkflowID     string `json:"workflowId"`
	StageID        string `json:"stageId,omitempty"`
	StatusDefID    string `json:"statusDefId"`
	Key            string `json:"key"`
	Name           string `json:"name"`
	ColumnKey      string `json:"columnKey"`
	Position       int    `json:"position"`
	IsStart        bool   `json:"isStart"`
	IsTerminal     bool   `json:"isTerminal"`
	IsActive       bool   `json:"isActive"`
	ReviewRequired bool   `json:"reviewRequired"`
}

type WorkflowTransition struct {
	ID             string    `json:"id"`
	WorkflowID     string    `json:"workflowId"`
	FromStageID    string    `json:"fromStageId"`
	FromStatusID   string    `json:"fromStatusId"`
	ToStageID      string    `json:"toStageId"`
	ToStatusID     string    `json:"toStatusId"`
	ActionKey      string    `json:"actionKey"`
	Name           string    `json:"name"`
	ActorType      string    `json:"actorType"`
	RequiresReview bool      `json:"requiresReview"`
	IsActive       bool      `json:"isActive"`
	Position       int       `json:"position"`
	CreatedAt      time.Time `json:"createdAt"`
}

type Team struct {
	ID          string     `json:"id"`
	Slug        string     `json:"slug"`
	Name        string     `json:"name"`
	Description string     `json:"description"`
	CreatedBy   *string    `json:"createdBy"`
	UpdatedBy   *string    `json:"updatedBy"`
	CreatedAt   time.Time  `json:"createdAt"`
	UpdatedAt   time.Time  `json:"updatedAt"`
	DeletedAt   *time.Time `json:"deletedAt,omitempty"`
}

type ProjectPermission struct {
	ID                string     `json:"id"`
	ProjectID         string     `json:"projectId"`
	PrincipalType     string     `json:"principalType"`
	PrincipalID       string     `json:"principalId"`
	Role              string     `json:"role"`
	InheritToChildren bool       `json:"inheritToChildren"`
	CreatedBy         *string    `json:"createdBy"`
	CreatedAt         time.Time  `json:"createdAt"`
	DeletedAt         *time.Time `json:"deletedAt,omitempty"`
}

type WorkflowCatalog struct {
	Workflows           []Workflow           `json:"workflows"`
	WorkflowStageDefs   []WorkflowStageDef   `json:"workflowStageDefs"`
	WorkflowStatusDefs  []WorkflowStatusDef  `json:"workflowStatusDefs"`
	WorkflowStages      []WorkflowStage      `json:"workflowStages"`
	WorkflowStatuses    []WorkflowStatus     `json:"workflowStatuses"`
	WorkflowTransitions []WorkflowTransition `json:"workflowTransitions"`
	Teams               []Team               `json:"teams"`
	ProjectPermissions  []ProjectPermission  `json:"projectPermissions"`
}
