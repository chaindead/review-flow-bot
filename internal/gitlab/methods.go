package gitlab

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
	client "gitlab.com/gitlab-org/api/client-go"
)

type GitLabUser struct {
	ID       int
	Username string
	Name     string
}

type MRInfo struct {
	IID       int
	ProjectID int
	Title     string
	SHA       string
	Status    string
}

func (g *Gitlab) GetCurrentUser(ctx context.Context, token string) (*GitLabUser, error) {
	// Create temporary client with user's token
	tempClient, err := client.NewClient(token, client.WithBaseURL(g.cfg.URL))
	if err != nil {
		return nil, fmt.Errorf("create temp client: %w", err)
	}

	user, _, err := tempClient.Users.CurrentUser(client.WithContext(ctx))
	if err != nil {
		return nil, fmt.Errorf("get current user: %w", err)
	}

	return &GitLabUser{
		ID:       user.ID,
		Username: user.Username,
		Name:     user.Name,
	}, nil
}

func (g *Gitlab) ParseMRURL(mrURL string) (projectID int, mrIID int, err error) {
	// Parse URL like: https://gitlab.com/group/project/-/merge_requests/123
	parsedURL, err := url.Parse(mrURL)
	if err != nil {
		return 0, 0, fmt.Errorf("parse url: %w", err)
	}

	// Split path: /group/project/-/merge_requests/123
	parts := strings.Split(strings.Trim(parsedURL.Path, "/"), "/")

	// Find merge_requests index
	mrIndex := -1
	for i, part := range parts {
		if part == "merge_requests" {
			mrIndex = i
			break
		}
	}

	if mrIndex == -1 || mrIndex+1 >= len(parts) {
		return 0, 0, fmt.Errorf("invalid merge request URL format")
	}

	// Get MR IID
	mrIID, err = strconv.Atoi(parts[mrIndex+1])
	if err != nil {
		return 0, 0, fmt.Errorf("parse MR IID: %w", err)
	}

	// Find dash index to get project path
	dashIndex := -1
	for i, part := range parts {
		if part == "-" {
			dashIndex = i
			break
		}
	}

	if dashIndex == -1 || dashIndex == 0 {
		return 0, 0, fmt.Errorf("invalid project path in URL")
	}

	// Get project path (everything before /-/)
	projectPath := strings.Join(parts[:dashIndex], "/")

	// Get project by path to get project ID
	project, _, err := g.client.Projects.GetProject(projectPath, nil, client.WithContext(context.Background()))
	if err != nil {
		return 0, 0, fmt.Errorf("get project by path %s: %w", projectPath, err)
	}

	return project.ID, mrIID, nil
}

func (g *Gitlab) GetMergeRequest(ctx context.Context, projectID, mrIID int) (*client.MergeRequest, error) {
	mr, _, err := g.client.MergeRequests.GetMergeRequest(
		projectID,
		mrIID,
		nil,
		client.WithContext(ctx),
	)
	if err != nil {
		return nil, fmt.Errorf("get merge request: %w", err)
	}

	return mr, nil
}

func (g *Gitlab) GetDiscussion(ctx context.Context, projectID, mrIID int) ([]*client.Discussion, time.Time, error) {
	const (
		maxPages = 10
		perPage  = 100
	)
	var dss []*client.Discussion
	for i := range maxPages {
		ds, _, err := g.client.Discussions.ListMergeRequestDiscussions(projectID, mrIID,
			&client.ListMergeRequestDiscussionsOptions{
				PerPage: perPage,
				Page:    i + 1,
			},
			client.WithContext(ctx),
		)
		if err != nil {
			return nil, time.Time{}, errors.Wrap(err, "list discussions")
		}

		dss = append(dss, ds...)
		if len(ds) < perPage {
			break
		}
	}

	if len(dss) >= maxPages*perPage {
		return dss, time.Time{}, errors.New("too many discussions")
	}

	lastUpdated := time.Time{}
	for _, d := range dss {
		for _, n := range d.Notes {
			if n.UpdatedAt != nil && n.UpdatedAt.After(lastUpdated) {
				lastUpdated = *n.UpdatedAt
			}
		}
	}

	return dss, lastUpdated, nil

}
