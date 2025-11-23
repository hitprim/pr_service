package models

import (
	"time"
)

type Team struct {
	ID       uint   `gorm:"primaryKey"`
	TeamName string `gorm:"uniqueIndex;not null" json:"team_name"`
	Members  []User `gorm:"foreignKey:TeamName;references:TeamName" json:"members"`
}

type User struct {
	UserID   string `gorm:"primaryKey;column:user_id" json:"user_id"`
	Username string `gorm:"not null" json:"username"`
	TeamName string `gorm:"index" json:"team_name"`
	IsActive bool   `gorm:"not null;default:true" json:"is_active"`
}

type PullRequest struct {
	PullRequestID     string     `gorm:"primaryKey;column:pull_request_id" json:"pull_request_id"`
	PullRequestName   string     `json:"pull_request_name"`
	AuthorID          string     `gorm:"index" json:"author_id"`
	Status            string     `gorm:"type:varchar(16);default:OPEN" json:"status"`
	AssignedReviewers []string   `gorm:"-" json:"assigned_reviewers"`
	Reviewers         []User     `gorm:"many2many:pr_reviewers;joinForeignKey:PullRequestID;JoinReferences:UserID" json:"-"`
	CreatedAt         *time.Time `gorm:"column:created_at" json:"createdAt"`
	MergedAt          *time.Time `gorm:"column:merged_at" json:"mergedAt"`
}

type PullRequestShort struct {
	PullRequestID   string `json:"pull_request_id"`
	PullRequestName string `json:"pull_request_name"`
	AuthorID        string `json:"author_id"`
	Status          string `json:"status"`
}
