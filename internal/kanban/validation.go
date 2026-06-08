package kanban

import (
	"math"
	"strings"
	"unicode"
)

const nonDragCompletedTransitionMessage = "只有用户确认完成后才能拖拽到「已完成」。"

var statusAliases = map[string]Status{
	"complete":    StatusCompleted,
	"completed":   StatusCompleted,
	"done":        StatusCompleted,
	"finish":      StatusCompleted,
	"finished":    StatusCompleted,
	"resolved":    StatusCompleted,
	"inprocess":   StatusInProgress,
	"in_process":  StatusInProgress,
	"in-process":  StatusInProgress,
	"inprogress":  StatusInProgress,
	"inreview":    StatusInReview,
	"in_review":   StatusInReview,
	"in-review":   StatusInReview,
	"review":      StatusInReview,
	"reviewing":   StatusInReview,
	"processing":  StatusInProgress,
	"running":     StatusInProgress,
	"backlog":     StatusBacklog,
	"todo":        StatusTodo,
	"in_progress": StatusInProgress,
}

func NormalizeStatus(value string) (Status, bool) {
	raw := strings.ToLower(strings.TrimSpace(value))
	status, ok := statusAliases[raw]
	return status, ok
}

func NormalizePriority(value string) (Priority, bool) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case string(PriorityHigh):
		return PriorityHigh, true
	case string(PriorityMedium):
		return PriorityMedium, true
	case string(PriorityLow):
		return PriorityLow, true
	default:
		return "", false
	}
}

func NormalizeSeverity(value string) (Severity, bool) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case string(SeverityCritical):
		return SeverityCritical, true
	case string(SeverityHigh):
		return SeverityHigh, true
	case string(SeverityMedium):
		return SeverityMedium, true
	case string(SeverityLow):
		return SeverityLow, true
	default:
		return "", false
	}
}

func NormalizeRunState(value string) (*RunState, bool) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "":
		return nil, true
	case string(RunStateRunning):
		state := RunStateRunning
		return &state, true
	case string(RunStateCompleted):
		state := RunStateCompleted
		return &state, true
	case string(RunStateFailed):
		state := RunStateFailed
		return &state, true
	case string(RunStateCancelled):
		state := RunStateCancelled
		return &state, true
	default:
		return nil, false
	}
}

func NullableTrimmed(value *string) *string {
	if value == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*value)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}

func Trimmed(value string) string {
	return strings.TrimSpace(value)
}

func NormalizeDescription(value *string) string {
	if value == nil {
		return ""
	}
	return strings.TrimSpace(*value)
}

func ValidPosition(value float64) bool {
	return !math.IsNaN(value) && !math.IsInf(value, 0)
}

func NormalizeProjectName(value string) string {
	return strings.TrimSpace(value)
}

func NormalizeProjectSlug(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	var builder strings.Builder
	lastWasDash := false
	for _, char := range value {
		switch {
		case unicode.IsLetter(char) || unicode.IsDigit(char) || char == '.' || char == '_':
			builder.WriteRune(char)
			lastWasDash = false
		case char == '-' || unicode.IsSpace(char):
			if !lastWasDash && builder.Len() > 0 {
				builder.WriteRune('-')
				lastWasDash = true
			}
		}
	}
	return strings.Trim(builder.String(), "-")
}

func ProjectSlugFromName(name string) string {
	slug := NormalizeProjectSlug(name)
	if slug != "" {
		return slug
	}
	return "project"
}
