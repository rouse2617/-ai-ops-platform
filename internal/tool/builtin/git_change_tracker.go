package builtin

import (
	"context"
	"fmt"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"ai-ops/internal/model"
)

// GitChangeTracker tracks git changes across repositories
type GitChangeTracker struct {
	repositories map[string]*git.Repository
	repoBasePath string
}

// NewGitChangeTracker creates a new git change tracker
func NewGitChangeTracker(repoBasePath string) *GitChangeTracker {
	return &GitChangeTracker{
		repositories: make(map[string]*git.Repository),
		repoBasePath: repoBasePath,
	}
}

// AddRepository adds a repository to track
func (g *GitChangeTracker) AddRepository(name, path string) error {
	repo, err := git.PlainOpen(path)
	if err != nil {
		return fmt.Errorf("failed to open repository %s: %w", name, err)
	}
	g.repositories[name] = repo
	return nil
}

// GetCommitsInTimeRange retrieves commits within a time range
func (g *GitChangeTracker) GetCommitsInTimeRange(ctx context.Context, repoName string, start, end time.Time) ([]model.GitCommit, error) {
	repo, ok := g.repositories[repoName]
	if !ok {
		return nil, fmt.Errorf("repository %s not found", repoName)
	}

	ref, err := repo.Head()
	if err != nil {
		return nil, fmt.Errorf("failed to get HEAD: %w", err)
	}

	commitIter, err := repo.Log(&git.LogOptions{From: ref.Hash()})
	if err != nil {
		return nil, fmt.Errorf("failed to get log: %w", err)
	}

	var commits []model.GitCommit
	err = commitIter.ForEach(func(c *object.Commit) error {
		if ctx.Err() != nil {
			return ctx.Err()
		}

		if c.Author.When.Before(start) {
			return nil
		}
		if c.Author.When.After(end) {
			return nil
		}

		changedFiles, err := g.getChangedFiles(c)
		if err != nil {
			return err
		}

		commits = append(commits, model.GitCommit{
			Hash:         c.Hash.String(),
			Author:       c.Author.Name,
			Email:        c.Author.Email,
			Message:      c.Message,
			Timestamp:    c.Author.When,
			ChangedFiles: changedFiles,
			Repository:   repoName,
			Branch:       ref.Name().Short(),
		})

		return nil
	})

	return commits, err
}

// getChangedFiles extracts changed files from a commit
func (g *GitChangeTracker) getChangedFiles(commit *object.Commit) ([]string, error) {
	var files []string

	if commit.NumParents() == 0 {
		tree, err := commit.Tree()
		if err != nil {
			return nil, err
		}
		tree.Files().ForEach(func(f *object.File) error {
			files = append(files, f.Name)
			return nil
		})
		return files, nil
	}

	parent, err := commit.Parent(0)
	if err != nil {
		return nil, err
	}

	patch, err := parent.Patch(commit)
	if err != nil {
		return nil, err
	}

	for _, filePatch := range patch.FilePatches() {
		from, to := filePatch.Files()
		if from != nil {
			files = append(files, from.Path())
		}
		if to != nil && (from == nil || from.Path() != to.Path()) {
			files = append(files, to.Path())
		}
	}

	return files, nil
}

// GetDiff retrieves the diff for a specific commit
func (g *GitChangeTracker) GetDiff(repoName, commitHash string) (string, error) {
	repo, ok := g.repositories[repoName]
	if !ok {
		return "", fmt.Errorf("repository %s not found", repoName)
	}

	commit, err := repo.CommitObject(plumbing.NewHash(commitHash))
	if err != nil {
		return "", fmt.Errorf("failed to get commit: %w", err)
	}

	if commit.NumParents() == 0 {
		return "Initial commit - no diff available", nil
	}

	parent, err := commit.Parent(0)
	if err != nil {
		return "", err
	}

	patch, err := parent.Patch(commit)
	if err != nil {
		return "", err
	}

	return patch.String(), nil
}

// Tool implementation for Agent system
type GitChangeTrackerTool struct {
	tracker *GitChangeTracker
}

func NewGitChangeTrackerTool(repoBasePath string) *GitChangeTrackerTool {
	return &GitChangeTrackerTool{
		tracker: NewGitChangeTracker(repoBasePath),
	}
}

func (t *GitChangeTrackerTool) Name() string {
	return "git_change_tracker"
}

func (t *GitChangeTrackerTool) Description() string {
	return "Track git changes and correlate with alerts. Searches for commits in a time range and analyzes changed files."
}

func (t *GitChangeTrackerTool) InputSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"repository": map[string]interface{}{
				"type":        "string",
				"description": "Repository name to search",
			},
			"time_range_minutes": map[string]interface{}{
				"type":        "number",
				"description": "Time range in minutes before alert (default: 30)",
			},
			"alert_time": map[string]interface{}{
				"type":        "string",
				"description": "Alert timestamp in RFC3339 format",
			},
		},
		"required": []string{"repository", "alert_time"},
	}
}

func (t *GitChangeTrackerTool) Execute(ctx context.Context, input map[string]interface{}) (interface{}, error) {
	repo, _ := input["repository"].(string)
	alertTimeStr, _ := input["alert_time"].(string)
	timeRange, _ := input["time_range_minutes"].(float64)

	if timeRange == 0 {
		timeRange = 30
	}

	alertTime, err := time.Parse(time.RFC3339, alertTimeStr)
	if err != nil {
		return nil, fmt.Errorf("invalid alert_time format: %w", err)
	}

	start := alertTime.Add(-time.Duration(timeRange) * time.Minute)
	commits, err := t.tracker.GetCommitsInTimeRange(ctx, repo, start, alertTime)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"commits": commits,
		"count":   len(commits),
		"time_window": map[string]interface{}{
			"start": start,
			"end":   alertTime,
		},
	}, nil
}
