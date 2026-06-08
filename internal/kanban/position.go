package kanban

import "sort"

func SortIssues(issues []Issue) []Issue {
	next := append([]Issue{}, issues...)
	sort.SliceStable(next, func(i, j int) bool {
		left := next[i]
		right := next[j]
		if statusRank(left.Status) != statusRank(right.Status) {
			return statusRank(left.Status) < statusRank(right.Status)
		}
		if left.Position != right.Position {
			return left.Position < right.Position
		}
		if !left.UpdatedAt.Equal(right.UpdatedAt) {
			return left.UpdatedAt.After(right.UpdatedAt)
		}
		return left.ID < right.ID
	})
	return next
}

func NextPosition(issues []Issue, status Status) float64 {
	var max float64
	for _, issue := range issues {
		if issue.Status == status && issue.DeletedAt == nil && issue.Position > max {
			max = issue.Position
		}
	}
	return max + 1
}

func statusRank(status Status) int {
	switch status {
	case StatusBacklog:
		return 0
	case StatusTodo:
		return 1
	case StatusInProgress:
		return 2
	case StatusInReview:
		return 3
	case StatusCompleted:
		return 4
	default:
		return 99
	}
}
