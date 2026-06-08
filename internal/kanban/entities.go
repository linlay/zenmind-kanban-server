package kanban

import "time"

type UserAccount struct {
	ID          string     `json:"id"`
	Email       string     `json:"email"`
	DisplayName string     `json:"displayName"`
	AvatarURL   *string    `json:"avatarUrl,omitempty"`
	Status      string     `json:"status"`
	CreatedAt   time.Time  `json:"createdAt"`
	UpdatedAt   time.Time  `json:"updatedAt"`
	DeletedAt   *time.Time `json:"deletedAt,omitempty"`
}

type UserInput struct {
	Email       string  `json:"email"`
	DisplayName string  `json:"displayName"`
	AvatarURL   *string `json:"avatarUrl"`
	Status      *string `json:"status"`
}

type UserUpdateInput struct {
	Email       *string `json:"email"`
	DisplayName *string `json:"displayName"`
	AvatarURL   *string `json:"avatarUrl"`
	Status      *string `json:"status"`
}

type TeamInput struct {
	Slug        *string `json:"slug"`
	Name        string  `json:"name"`
	Description *string `json:"description"`
}

type TeamUpdateInput struct {
	Slug        *string `json:"slug"`
	Name        *string `json:"name"`
	Description *string `json:"description"`
}

type TeamMember struct {
	TeamID    string     `json:"teamId"`
	UserID    string     `json:"userId"`
	Role      string     `json:"role"`
	InvitedBy *string    `json:"invitedBy,omitempty"`
	JoinedAt  time.Time  `json:"joinedAt"`
	LeftAt    *time.Time `json:"leftAt,omitempty"`
}

type TeamMemberInput struct {
	TeamID    string  `json:"teamId"`
	UserID    string  `json:"userId"`
	Role      string  `json:"role"`
	InvitedBy *string `json:"invitedBy"`
}

type TeamMemberUpdateInput struct {
	Role   *string `json:"role"`
	LeftAt *string `json:"leftAt"`
}

type Agent struct {
	ID           string     `json:"id"`
	AgentKey     string     `json:"agentKey"`
	Name         string     `json:"name"`
	Description  string     `json:"description"`
	Role         string     `json:"role"`
	Model        *string    `json:"model,omitempty"`
	ModelVersion *string    `json:"modelVersion,omitempty"`
	Enabled      bool       `json:"enabled"`
	CreatedBy    *string    `json:"createdBy,omitempty"`
	UpdatedBy    *string    `json:"updatedBy,omitempty"`
	CreatedAt    time.Time  `json:"createdAt"`
	UpdatedAt    time.Time  `json:"updatedAt"`
	DeletedAt    *time.Time `json:"deletedAt,omitempty"`
}

type AgentInput struct {
	AgentKey     string  `json:"agentKey"`
	Name         string  `json:"name"`
	Description  *string `json:"description"`
	Role         *string `json:"role"`
	Model        *string `json:"model"`
	ModelVersion *string `json:"modelVersion"`
	Enabled      *bool   `json:"enabled"`
}

type AgentUpdateInput struct {
	AgentKey     *string `json:"agentKey"`
	Name         *string `json:"name"`
	Description  *string `json:"description"`
	Role         *string `json:"role"`
	Model        *string `json:"model"`
	ModelVersion *string `json:"modelVersion"`
	Enabled      *bool   `json:"enabled"`
}

type ProjectPermissionInput struct {
	ProjectID         string `json:"projectId"`
	PrincipalType     string `json:"principalType"`
	PrincipalID       string `json:"principalId"`
	Role              string `json:"role"`
	InheritToChildren *bool  `json:"inheritToChildren"`
}

type ProjectPermissionUpdateInput struct {
	Role              *string `json:"role"`
	InheritToChildren *bool   `json:"inheritToChildren"`
}

type IssueLabel struct {
	ID        string     `json:"id"`
	ProjectID string     `json:"projectId"`
	Key       string     `json:"key"`
	Name      string     `json:"name"`
	Color     string     `json:"color"`
	CreatedBy *string    `json:"createdBy,omitempty"`
	UpdatedBy *string    `json:"updatedBy,omitempty"`
	CreatedAt time.Time  `json:"createdAt"`
	UpdatedAt time.Time  `json:"updatedAt"`
	DeletedAt *time.Time `json:"deletedAt,omitempty"`
}

type IssueLabelInput struct {
	ProjectID *string `json:"projectId"`
	Key       *string `json:"key"`
	Name      string  `json:"name"`
	Color     *string `json:"color"`
}

type IssueLabelUpdateInput struct {
	Key   *string `json:"key"`
	Name  *string `json:"name"`
	Color *string `json:"color"`
}

type IssueLabelLink struct {
	IssueID string `json:"issueId"`
	LabelID string `json:"labelId"`
}

type IssueLabelsSetInput struct {
	IssueID  string   `json:"issueId"`
	LabelIDs []string `json:"labelIds"`
}

type IssueDependency struct {
	ID          string     `json:"id"`
	FromIssueID string     `json:"fromIssueId"`
	ToIssueID   string     `json:"toIssueId"`
	Type        string     `json:"type"`
	CreatedBy   *string    `json:"createdBy,omitempty"`
	CreatedAt   time.Time  `json:"createdAt"`
	DeletedAt   *time.Time `json:"deletedAt,omitempty"`
}

type IssueDependencyInput struct {
	FromIssueID string  `json:"fromIssueId"`
	ToIssueID   string  `json:"toIssueId"`
	Type        *string `json:"type"`
}

type Review struct {
	ID          string     `json:"id"`
	IssueID     string     `json:"issueId"`
	AgentRunID  *string    `json:"agentRunId,omitempty"`
	ReviewType  string     `json:"reviewType"`
	ReviewerID  *string    `json:"reviewerId,omitempty"`
	Status      string     `json:"status"`
	RequestedBy *string    `json:"requestedBy,omitempty"`
	RequestedAt time.Time  `json:"requestedAt"`
	SubmittedAt *time.Time `json:"submittedAt,omitempty"`
	Summary     string     `json:"summary"`
	CreatedAt   time.Time  `json:"createdAt"`
	UpdatedAt   time.Time  `json:"updatedAt"`
	DeletedAt   *time.Time `json:"deletedAt,omitempty"`
}

type ReviewInput struct {
	IssueID    string  `json:"issueId"`
	AgentRunID *string `json:"agentRunId"`
	ReviewType *string `json:"reviewType"`
	ReviewerID *string `json:"reviewerId"`
	Status     *string `json:"status"`
	Summary    *string `json:"summary"`
}

type ReviewUpdateInput struct {
	ReviewerID  *string `json:"reviewerId"`
	Status      *string `json:"status"`
	SubmittedAt *string `json:"submittedAt"`
	Summary     *string `json:"summary"`
}

type ReviewComment struct {
	ID          string         `json:"id"`
	ReviewID    string         `json:"reviewId"`
	IssueID     string         `json:"issueId"`
	AuthorID    *string        `json:"authorId,omitempty"`
	AuthorAgent *string        `json:"authorAgent,omitempty"`
	Body        string         `json:"body"`
	Anchor      map[string]any `json:"anchor"`
	CreatedAt   time.Time      `json:"createdAt"`
	UpdatedAt   time.Time      `json:"updatedAt"`
	DeletedAt   *time.Time     `json:"deletedAt,omitempty"`
}

type ReviewCommentInput struct {
	ReviewID    string         `json:"reviewId"`
	IssueID     string         `json:"issueId"`
	AuthorID    *string        `json:"authorId"`
	AuthorAgent *string        `json:"authorAgent"`
	Body        string         `json:"body"`
	Anchor      map[string]any `json:"anchor"`
}

type ReviewCommentUpdateInput struct {
	Body   *string        `json:"body"`
	Anchor map[string]any `json:"anchor"`
}

type AgentRun struct {
	ID           string     `json:"id"`
	IssueID      string     `json:"issueId"`
	AgentID      *string    `json:"agentId,omitempty"`
	WorkerAgent  *string    `json:"workerAgent,omitempty"`
	DelegatedBy  *string    `json:"delegatedBy,omitempty"`
	ChatID       *string    `json:"chatId,omitempty"`
	RunID        *string    `json:"runId,omitempty"`
	SessionID    *string    `json:"sessionId,omitempty"`
	Status       string     `json:"status"`
	Confidence   *float64   `json:"confidence,omitempty"`
	InputTokens  *int64     `json:"inputTokens,omitempty"`
	OutputTokens *int64     `json:"outputTokens,omitempty"`
	CostMicros   *int64     `json:"costMicros,omitempty"`
	StartedAt    *time.Time `json:"startedAt,omitempty"`
	FinishedAt   *time.Time `json:"finishedAt,omitempty"`
	ErrorMessage *string    `json:"errorMessage,omitempty"`
	CreatedAt    time.Time  `json:"createdAt"`
	UpdatedAt    time.Time  `json:"updatedAt"`
}

type AgentToolCall struct {
	ID           string         `json:"id"`
	AgentRunID   string         `json:"agentRunId"`
	ToolName     string         `json:"toolName"`
	ToolType     string         `json:"toolType"`
	Status       string         `json:"status"`
	Input        map[string]any `json:"input"`
	Output       map[string]any `json:"output"`
	StartedAt    *time.Time     `json:"startedAt,omitempty"`
	FinishedAt   *time.Time     `json:"finishedAt,omitempty"`
	ErrorMessage *string        `json:"errorMessage,omitempty"`
}

type EventLogItem struct {
	ID         int64          `json:"id"`
	BoardID    string         `json:"boardId"`
	ProjectID  *string        `json:"projectId,omitempty"`
	IssueID    *string        `json:"issueId,omitempty"`
	Revision   int64          `json:"revision"`
	EventType  string         `json:"eventType"`
	ActorID    *string        `json:"actorId,omitempty"`
	ActorAgent *string        `json:"actorAgent,omitempty"`
	Payload    map[string]any `json:"payload"`
	CreatedAt  time.Time      `json:"createdAt"`
}

type TransitionInput struct {
	ID           string   `json:"id"`
	TransitionID *string  `json:"transitionId"`
	ActionKey    *string  `json:"actionKey"`
	StageID      *string  `json:"stageId"`
	StatusID     *string  `json:"statusId"`
	Position     *float64 `json:"position"`
}

type MutationResult struct {
	OK        bool   `json:"ok"`
	Message   string `json:"message"`
	BoardID   string `json:"boardId"`
	ProjectID string `json:"projectId,omitempty"`
	Revision  int64  `json:"revision"`
}
