package gitlab

import (
	"fmt"

	"github.com/samber/do/v2"
	client "gitlab.com/gitlab-org/api/client-go"

	"github.com/chaindead/review-flow-bot/internal/config"
)

type Gitlab struct {
	cfg config.Gitlab `do:"cfg.gitlab"`

	client *client.Client
}

func New(i do.Injector) (*Gitlab, error) {
	git, err := do.InvokeStruct[*Gitlab](i)
	if err != nil {
		return nil, fmt.Errorf("failed to invoke struct: %w", err)
	}

	cli, err := client.NewClient(
		git.cfg.Token,
		client.WithBaseURL(git.cfg.URL),
	)
	if err != nil {
		return nil, fmt.Errorf("create gitlab client: %w", err)
	}

	git.client = cli

	return git, nil
}

func (g *Gitlab) C() *client.Client {
	return g.client
}
