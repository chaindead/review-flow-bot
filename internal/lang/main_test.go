package lang

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/samber/do/v2"
	"github.com/stretchr/testify/require"

	"github.com/chaindead/review-flow-bot/internal/db/models"
)

func TestLocalizer_Get(t *testing.T) {
	injector := do.New()
	do.Provide(injector, New)

	loc := do.MustInvoke[*Localizer](injector)

	//common
	var mr_title_link = models.MergeRequest{
		Title: "EXAMPLE_TITLE: TASK-1 test",
		Link:  "https://gitlab.com/gitlab-org/gitlab-test/-/merge_requests/1",
	}

	tests := []struct {
		name string
		args Args
		data string
	}{
		{
			name: "admin.list.teams",
			args: Args{
				"Teams":      []models.Team{},
				"Unassigned": []models.User{},
			},
			data: `{"Unassigned":[{"AuthoredMRs":null,"CreatedAt":"2025-10-08T20:06:25Z","GitlabID":111,"GitlabUsername":"gitlab_user-name","ID":111,"ReviewingMRs":null,"TeamMembership":null,"TelegramUsername":"telegram_user_name","UpdatedAt":"2025-10-08T20:06:25Z"}],"Teams":[{"CreatedAt":"2025-10-08T20:06:25Z","Members":[{"Role":"reviewer","Team":null,"TeamID":"backend","User":{"AuthoredMRs":null,"CreatedAt":"2025-10-08T20:06:25Z","GitlabID":6,"GitlabUsername":"reviewer2","ID":7076621751,"ReviewingMRs":null,"TeamMembership":null,"TelegramUsername":"chaindead","UpdatedAt":"2025-10-08T20:06:25Z"},"UserID":7076621751},{"Role":"reviewer","Team":null,"TeamID":"backend","User":{"AuthoredMRs":null,"CreatedAt":"2025-10-08T20:06:25Z","GitlabID":5,"GitlabUsername":"reviewer1","ID":268850924,"ReviewingMRs":null,"TeamMembership":null,"TelegramUsername":"doejon","UpdatedAt":"2025-10-08T20:06:25Z"},"UserID":268850924},{"Role":"member","Team":null,"TeamID":"backend","User":{"AuthoredMRs":null,"CreatedAt":"2025-10-08T20:06:25Z","GitlabID":4,"GitlabUsername":"developer2","ID":5875418647,"ReviewingMRs":null,"TeamMembership":null,"TelegramUsername":"letztes","UpdatedAt":"2025-10-08T20:06:25Z"},"UserID":5875418647}],"Name":"backend"},{"CreatedAt":"2025-10-08T20:06:25Z","Members":null,"Name":"frontend"}]}`,
		},
		{
			name: "admin.list.teams:no unassigned",
			args: Args{
				"Teams":      []models.Team{},
				"Unassigned": []models.User{},
			},
			data: `{"Teams":[{"CreatedAt":"2025-10-08T20:06:25Z","Members":[{"Role":"reviewer","Team":null,"TeamID":"backend","User":{"AuthoredMRs":null,"CreatedAt":"2025-10-08T20:06:25Z","GitlabID":6,"GitlabUsername":"reviewer2","ID":7076621751,"ReviewingMRs":null,"TeamMembership":null,"TelegramUsername":"chaindead","UpdatedAt":"2025-10-08T20:06:25Z"},"UserID":7076621751},{"Role":"reviewer","Team":null,"TeamID":"backend","User":{"AuthoredMRs":null,"CreatedAt":"2025-10-08T20:06:25Z","GitlabID":5,"GitlabUsername":"reviewer1","ID":268850924,"ReviewingMRs":null,"TeamMembership":null,"TelegramUsername":"doejon","UpdatedAt":"2025-10-08T20:06:25Z"},"UserID":268850924},{"Role":"member","Team":null,"TeamID":"backend","User":{"AuthoredMRs":null,"CreatedAt":"2025-10-08T20:06:25Z","GitlabID":4,"GitlabUsername":"developer2","ID":5875418647,"ReviewingMRs":null,"TeamMembership":null,"TelegramUsername":"letztes","UpdatedAt":"2025-10-08T20:06:25Z"},"UserID":5875418647}],"Name":"backend"},{"CreatedAt":"2025-10-08T20:06:25Z","Members":null,"Name":"frontend"}]}`,
		},
		{
			name: "admin.assign.usage",
			args: NoArgs,
		},
		{
			name: "admin.assign.bad_role",
			args: NoArgs,
		},
		{
			name: "admin.assign.not_logged",
			args: Args{
				"TelegramUsername": "example_username",
			},
		},
		{
			name: "admin.assign.success",
			args: Args{
				"TelegramUsername": "my_username",
				"Team":             "my_team",
				"Role":             "my_role",
			},
		},
		{
			name: "start.welcome",
			args: NoArgs,
		},
		{
			name: "help.commands:admin",
			args: Args{
				"IsAdmin": true,
			},
		},
		{
			name: "help.commands:user",
			args: Args{
				"IsAdmin": false,
			},
		},
		{
			name: "login.usage",
			args: NoArgs,
		},
		{
			name: "login.fail_gitlab",
			args: NoArgs,
		},
		{
			name: "login.success",
			args: Args{
				"TelegramUsername": "my_tg",
				"GitlabUsername":   "my_gitlab",
			},
		},
		{
			name: "my.success: empty",
			args: Args{
				"User": &models.User{
					GitlabUsername: "my_gitlab",
				},
			},
		},
		{
			name: "my.success: no review",
			args: Args{
				"User": &models.User{
					GitlabUsername: "my_gitlab",
					TeamMembership: &models.TeamMember{TeamID: "my_team", Role: models.RoleReviewer},
				},
			},
		},
		{
			name: "my.success: mr author",
			args: Args{
				"User": &models.User{
					GitlabUsername: "my_gitlab",
					TeamMembership: &models.TeamMember{TeamID: "my_team"},
					AuthoredMRs: []*models.MergeRequest{
						{
							Title: "EXAMPLE_TITLE: TASK-228 test",
							Link:  "https://gitlab.com/gitlab-org/gitlab-test/-/merge_requests/513",
							Reviewers: []*models.MRReviewer{
								{
									Status: models.MRStatusPending,
									Reviewer: &models.User{
										TelegramUsername: "tg_reviewer",
										GitlabUsername:   "gl_reviewer",
									},
								},
								{
									Status: models.MRStatusApproved,
									Reviewer: &models.User{
										TelegramUsername: "tg_reviewer2",
										GitlabUsername:   "gl_reviewer2",
									},
								},
								{
									Status: models.MRStatusRejected,
									Reviewer: &models.User{
										TelegramUsername: "tg_reviewer3",
										GitlabUsername:   "gl_reviewer3",
									},
								},
							},
						},
						{
							Title: "EXAMPLE_TITLE: no reviwer",
							Link:  "https://gitlab.com/gitlab-org/gitlab-test/-/merge_requests/1",
						},
					},
				},
			},
		},
		{
			name: "my.success: review",
			args: Args{
				"User": &models.User{
					GitlabUsername: "my_gitlab",
					TeamMembership: &models.TeamMember{
						TeamID: "my_team",
						Role:   models.RoleReviewer,
					},
					ReviewingMRs: []*models.MRReviewer{
						{
							Status: models.MRStatusApproved,
							MergeRequest: &models.MergeRequest{
								Title: "EXAMPLE_TITLE: TASK-1 test",
								Link:  "https://gitlab.com/gitlab-org/gitlab-test/-/merge_requests/1",
								Author: &models.User{
									TelegramUsername: "tg_author",
									GitlabUsername:   "gl_author",
								},
							},
						},
						{
							Status: models.MRStatusPending,
							MergeRequest: &models.MergeRequest{
								Title: "EXAMPLE_TITLE: TASK-2 test",
								Link:  "https://gitlab.com/gitlab-org/gitlab-test/-/merge_requests/1",
								Author: &models.User{
									TelegramUsername: "tg2_author",
									GitlabUsername:   "gl2_author",
								},
							},
						},
					},
				},
			},
		},
		{
			name: "review.notify",
			args: Args{
				"MR": &models.MergeRequest{
					Title: "EXAMPLE_TITLE: TASK-1 test",
					Link:  "https://localhost:90/gitlab-org/gitlab-test/-/merge_requests/1",
					Author: &models.User{
						TelegramUsername: "tg_author",
						GitlabUsername:   "gl_author",
					},
					Reviewers: []*models.MRReviewer{
						{
							Reviewer: &models.User{
								TelegramUsername: "tg_reviewer",
								GitlabUsername:   "gl_reviewer",
							},
						},
						{
							Reviewer: &models.User{
								TelegramUsername: "tg_reviewer2",
								GitlabUsername:   "gl_reviewer2",
							},
						},
					},
				},
			},
		},
		{
			name: "review.success",
			args: Args{
				"MR": &models.MergeRequest{
					Title: "EXAMPLE_TITLE: TASK-1 test",
					Link:  "https://localhost:90/gitlab-org/gitlab-test/-/merge_requests/1",
					Reviewers: []*models.MRReviewer{
						{
							Reviewer: &models.User{
								TelegramUsername: "tg_reviewer",
								GitlabUsername:   "gl_reviewer",
							},
						},
						{
							Reviewer: &models.User{
								TelegramUsername: "tg_reviewer2",
								GitlabUsername:   "gl_reviewer2",
							},
						},
					},
				},
			},
		},
		{
			name: "review.bad_force",
			args: Args{
				"Nicks": []string{"test", "test2"},
			},
		},
		{
			name: "watch.notify.merged",
			args: Args{
				"MR": mr_title_link,
			},
		},
		{
			name: "watch.notify.closed",
			args: Args{
				"MR": mr_title_link,
			},
		},
		{
			name: "watch.notify.reviewer: nil status",
			args: Args{
				"MR":           mr_title_link,
				"StatusChange": nil,
				"Comments":     []string{"test"},
			},
		},
		{
			name: "watch.notify.reviewer: only status",
			args: Args{
				"MR": mr_title_link,
				"StatusChange": &models.MRReviewer{
					Status: models.MRStatusPending,
				},
			},
		},
		{
			name: "watch.notify.reviewer: status not pending",
			args: Args{
				"MR": mr_title_link,
				"StatusChange": &models.MRReviewer{
					Status: models.MRStatusApproved,
				},
			},
		},
		{
			name: "watch.notify.reviewer: status+comments",
			args: Args{
				"MR": mr_title_link,
				"StatusChange": &models.MRReviewer{
					Status: models.MRStatusPending,
				},
				"Comments": []string{"test", "test2", "test3"},
			},
		},
		{
			name: "watch.notify.author",
			args: Args{
				"MR": mr_title_link,
				"StatusChanges": []*models.MRReviewer{
					{
						Status: models.MRStatusApproved,
						Reviewer: &models.User{
							TelegramUsername: "tg_reviewer",
							GitlabUsername:   "gl_reviewer",
						},
					},
					{
						Status: models.MRStatusPending,
						Reviewer: &models.User{
							TelegramUsername: "tg_reviewer2",
							GitlabUsername:   "gl_reviewer2",
						},
					},
				},
				"Comments": []string{"test", "test2"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id := strings.Split(tt.name, ":")[0]

			if len(tt.data) != 0 {
				err := json.Unmarshal([]byte(tt.data), &tt.args)
				require.NoError(t, err)
			}

			text := loc.Get(id, tt.args)
			fmt.Println(text)
		})
	}
}
