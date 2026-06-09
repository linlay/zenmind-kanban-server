package kanban

func ComputeProjectIssueStats(projects []Project, issues []Issue) []ProjectIssueStat {
	statsByProjectID := make(map[string]*ProjectIssueStat, len(projects))
	projectsByID := make(map[string]Project, len(projects))
	for _, project := range projects {
		projectID := project.ID
		statsByProjectID[projectID] = &ProjectIssueStat{ProjectID: projectID}
		projectsByID[projectID] = project
	}

	for _, issue := range issues {
		projectID := issue.ProjectID
		if projectID == "" {
			projectID = DefaultProjectID
		}
		columnKey := issue.ColumnKey
		if columnKey == "" {
			columnKey = string(issue.Status)
		}
		seen := map[string]bool{}
		for projectID != "" && !seen[projectID] {
			seen[projectID] = true
			stat := statsByProjectID[projectID]
			if stat == nil {
				break
			}
			stat.IssueCount++
			if columnKey == string(StatusInProgress) {
				stat.InProgressIssueCount++
			}
			project := projectsByID[projectID]
			if project.ParentID == nil {
				break
			}
			projectID = *project.ParentID
		}
	}

	stats := make([]ProjectIssueStat, 0, len(projects))
	for _, project := range projects {
		if stat := statsByProjectID[project.ID]; stat != nil {
			stats = append(stats, *stat)
		}
	}
	return stats
}
