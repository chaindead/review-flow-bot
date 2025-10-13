package models

import (
	"time"

	"github.com/uptrace/bun"
)

type User struct {
	bun.BaseModel `bun:"table:users,alias:u"`

	ID               int64     `bun:"id,pk"`
	GitlabID         int64     `bun:"gitlab_id,notnull"`
	TelegramUsername string    `bun:"telegram_username,notnull"`
	GitlabUsername   string    `bun:"gitlab_username,notnull"`
	CreatedAt        time.Time `bun:"created_at,notnull,default:current_timestamp"`
	UpdatedAt        time.Time `bun:"updated_at,notnull,default:current_timestamp"`

	// Relationships
	TeamMembership *TeamMember     `bun:"rel:has-one,join:id=user_id"`
	AuthoredMRs    []*MergeRequest `bun:"rel:has-many,join:id=author_id"`
	ReviewingMRs   []*MRReviewer   `bun:"rel:has-many,join:id=reviewer_id"`
}

type Team struct {
	bun.BaseModel `bun:"table:teams,alias:t"`

	Name      string    `bun:"name,pk"`
	CreatedAt time.Time `bun:"created_at,notnull,default:current_timestamp"`

	// Has-many relationships
	Members []*TeamMember `bun:"rel:has-many,join:name=team_id"`
}

const (
	RoleMember   = "member"
	RoleReviewer = "reviewer"
)

type TeamMember struct {
	bun.BaseModel `bun:"table:team_members,alias:tm"`

	UserID int64  `bun:"user_id,pk"`
	TeamID string `bun:"team_id,notnull"`
	Role   string `bun:"role,notnull,default:'member'"`

	// Belongs-to relationships
	User *User `bun:"rel:belongs-to,join:user_id=id"`
	Team *Team `bun:"rel:belongs-to,join:team_id=name"`
}

type MergeRequest struct {
	bun.BaseModel `bun:"table:merge_requests,alias:mr"`

	MRID            int       `bun:"mr_id,pk,"`
	ProjectID       int       `bun:"project_id,pk"`
	AuthorID        int64     `bun:"author_id,notnull"`
	Title           string    `bun:"title,notnull"`
	Link            string    `bun:"link,notnull"`
	GitlabUpdatedAt time.Time `bun:"ext_updated_at,nullzero"`
	UpdatedAt       time.Time `bun:"updated_at,notnull,default:current_timestamp"`

	// Relationships
	Author    *User         `bun:"rel:belongs-to,join:author_id=id,on_delete:cascade"`
	Reviewers []*MRReviewer `bun:"rel:has-many,join:mr_id=mr_id,join:project_id=project_id,on_delete:cascade"`
}

const (
	MRStatusPending  = "pending"
	MRStatusApproved = "approved"
	MRStatusRejected = "feedback"
)

type MRReviewer struct {
	bun.BaseModel `bun:"table:mr_reviewers,alias:mrr"`

	MRID           int       `bun:"mr_id,pk"`
	ProjectID      int       `bun:"project_id,pk"`
	ReviewerID     int64     `bun:"reviewer_id,pk"`
	AssignedAt     time.Time `bun:"assigned_at,notnull,default:current_timestamp"`
	LastReminderAt time.Time `bun:"last_reminder_at,notnull,default:current_timestamp"`
	Status         string    `bun:"status,notnull,default:'pending'"`

	// Belongs-to relationships
	MergeRequest *MergeRequest `bun:"rel:belongs-to,join:mr_id=mr_id,join:project_id=project_id"`
	Reviewer     *User         `bun:"rel:belongs-to,join:reviewer_id=id"`
}
