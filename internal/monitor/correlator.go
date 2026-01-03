package monitor

import (
	"context"
	"path/filepath"
	"strings"
	"time"

	"ai-ops/internal/model"
	"ai-ops/internal/tool/builtin"
)

// CodeChangeCorrelator correlates alerts with code changes
type CodeChangeCorrelator struct {
	gitTracker      *builtin.GitChangeTracker
	serviceMappings map[string]*model.ServiceCodeMapping
}

// NewCodeChangeCorrelator creates a new correlator
func NewCodeChangeCorrelator(gitTracker *builtin.GitChangeTracker) *CodeChangeCorrelator {
	return &CodeChangeCorrelator{
		gitTracker:      gitTracker,
		serviceMappings: make(map[string]*model.ServiceCodeMapping),
	}
}

// RegisterServiceMapping registers a service to code mapping
func (c *CodeChangeCorrelator) RegisterServiceMapping(mapping *model.ServiceCodeMapping) {
	c.serviceMappings[mapping.ServiceName] = mapping
}

// AnalyzeAlert analyzes an alert and finds correlated code changes
func (c *CodeChangeCorrelator) AnalyzeAlert(ctx context.Context, alert *Alert) (*model.CodeChangeCorrelation, error) {
	mapping, ok := c.serviceMappings[alert.Service]
	if !ok {
		return nil, nil // No mapping configured
	}

	timeWindow := model.TimeWindow{
		AlertTime:   alert.Timestamp,
		SearchStart: alert.Timestamp.Add(-30 * time.Minute),
		SearchEnd:   alert.Timestamp,
	}

	var allCommits []model.GitCommit
	for _, repo := range mapping.Repositories {
		commits, err := c.gitTracker.GetCommitsInTimeRange(ctx, repo, timeWindow.SearchStart, timeWindow.SearchEnd)
		if err != nil {
			continue // Skip failed repos
		}
		allCommits = append(allCommits, commits...)
	}

	if len(allCommits) == 0 {
		return nil, nil
	}

	suspiciousFiles := c.rankSuspiciousFiles(allCommits, mapping, alert)
	score := c.calculateCorrelationScore(suspiciousFiles, allCommits, alert)

	return &model.CodeChangeCorrelation{
		AlertID:          alert.ID,
		AlertName:        alert.Name,
		Service:          alert.Service,
		Commits:          allCommits,
		CorrelationScore: score,
		SuspiciousFiles:  suspiciousFiles,
		TimeWindow:       timeWindow,
	}, nil
}

// rankSuspiciousFiles ranks files by suspicion level
func (c *CodeChangeCorrelator) rankSuspiciousFiles(commits []model.GitCommit, mapping *model.ServiceCodeMapping, alert *Alert) []model.SuspiciousFile {
	fileScores := make(map[string]*model.SuspiciousFile)

	for _, commit := range commits {
		timeDelta := alert.Timestamp.Sub(commit.Timestamp).Minutes()
		timeScore := 1.0 - (timeDelta / 30.0) // Closer to alert = higher score

		for _, file := range commit.ChangedFiles {
			score := timeScore
			reason := "Recent change"

			// Path matching
			if c.matchesCodePaths(file, mapping.CodePaths) {
				score += 0.3
				reason += ", matches service path"
			}

			// Keyword matching in commit message
			if c.containsKeywords(commit.Message, mapping.Keywords) {
				score += 0.2
				reason += ", relevant commit message"
			}

			// Alert-specific keywords
			if c.containsKeywords(file, c.extractKeywordsFromAlert(alert)) {
				score += 0.3
				reason += ", matches alert context"
			}

			// File type scoring
			if strings.HasSuffix(file, ".py") || strings.HasSuffix(file, ".go") || strings.HasSuffix(file, ".java") {
				score += 0.1
			}

			if existing, ok := fileScores[file]; !ok || score > existing.Score {
				fileScores[file] = &model.SuspiciousFile{
					FilePath:   file,
					CommitHash: commit.Hash,
					ChangeType: "modified",
					Score:      score,
					Reason:     reason,
					DiffURL:    c.generateDiffURL(commit.Repository, commit.Hash, file),
				}
			}
		}
	}

	// Convert to sorted slice
	var result []model.SuspiciousFile
	for _, sf := range fileScores {
		result = append(result, *sf)
	}

	// Sort by score descending
	for i := 0; i < len(result)-1; i++ {
		for j := i + 1; j < len(result); j++ {
			if result[j].Score > result[i].Score {
				result[i], result[j] = result[j], result[i]
			}
		}
	}

	// Return top 5
	if len(result) > 5 {
		result = result[:5]
	}

	return result
}

// calculateCorrelationScore calculates overall correlation score
func (c *CodeChangeCorrelator) calculateCorrelationScore(files []model.SuspiciousFile, commits []model.GitCommit, alert *Alert) float64 {
	if len(files) == 0 {
		return 0.0
	}

	// Base score from top suspicious file
	score := files[0].Score

	// Boost for multiple commits
	if len(commits) > 1 {
		score += 0.1
	}

	// Boost for high severity alerts
	if alert.Severity == "critical" {
		score += 0.1
	}

	// Normalize to 0-1
	if score > 1.0 {
		score = 1.0
	}

	return score
}

// matchesCodePaths checks if file matches any code path pattern
func (c *CodeChangeCorrelator) matchesCodePaths(file string, paths []string) bool {
	for _, path := range paths {
		if strings.Contains(file, path) {
			return true
		}
		if matched, _ := filepath.Match(path, file); matched {
			return true
		}
	}
	return false
}

// containsKeywords checks if text contains any keywords
func (c *CodeChangeCorrelator) containsKeywords(text string, keywords []string) bool {
	lower := strings.ToLower(text)
	for _, kw := range keywords {
		if strings.Contains(lower, strings.ToLower(kw)) {
			return true
		}
	}
	return false
}

// extractKeywordsFromAlert extracts relevant keywords from alert
func (c *CodeChangeCorrelator) extractKeywordsFromAlert(alert *Alert) []string {
	keywords := []string{alert.Service}

	// Extract from alert name
	parts := strings.Split(alert.Name, "_")
	keywords = append(keywords, parts...)

	// Extract from labels
	if db, ok := alert.Labels["database"]; ok {
		keywords = append(keywords, db)
	}
	if endpoint, ok := alert.Labels["endpoint"]; ok {
		keywords = append(keywords, endpoint)
	}

	return keywords
}

// generateDiffURL generates a URL to view the diff
func (c *CodeChangeCorrelator) generateDiffURL(repo, commitHash, file string) string {
	// This would be customized based on your Git hosting (GitHub, GitLab, etc.)
	return ""
}
