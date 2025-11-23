package handlers

import (
	"net/http"
	"pr_service/internal/models"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func TeamAdd(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req models.Team
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, models.ErrorResponse{
				Error: models.APIError{
					Code:    models.ErrNotFound,
					Message: "invalid json",
				},
			})
			return
		}

		var existing models.Team
		err := db.Where("team_name = ?", req.TeamName).First(&existing).Error
		if err == nil {
			c.JSON(http.StatusBadRequest, models.ErrorResponse{
				Error: models.APIError{
					Code:    models.ErrTeamExists,
					Message: "team already exists",
				},
			})
			return
		}

		team := models.Team{
			TeamName: req.TeamName,
		}
		db.Create(&team)

		for _, m := range req.Members {
			db.Where(models.User{UserID: m.UserID}).
				Assign(models.User{
					Username: m.Username,
					TeamName: req.TeamName,
					IsActive: m.IsActive,
				}).
				FirstOrCreate(&models.User{})
		}

		var out models.Team
		db.Preload("Members").Where("team_name = ?", req.TeamName).First(&out)

		c.JSON(http.StatusCreated, gin.H{"team": out})
	}
}

func TeamGet(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		teamName := c.Query("team_name")
		if teamName == "" {
			c.JSON(http.StatusBadRequest, models.ErrorResponse{
				Error: models.APIError{
					Code:    models.ErrNotFound,
					Message: "missing team_name",
				},
			})
			return
		}

		var team models.Team
		err := db.Preload("Members").Where("team_name = ?", teamName).First(&team).Error
		if err != nil {
			c.JSON(http.StatusNotFound, models.ErrorResponse{
				Error: models.APIError{
					Code:    models.ErrNotFound,
					Message: "team not found",
				},
			})
			return
		}

		c.JSON(http.StatusOK, team)
	}
}

func SetIsActive(db *gorm.DB) gin.HandlerFunc {
	type Req struct {
		UserID   string `json:"user_id"`
		IsActive bool   `json:"is_active"`
	}

	return func(c *gin.Context) {
		var req Req
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, models.ErrorResponse{
				Error: models.APIError{Code: models.ErrNotFound, Message: "invalid json"},
			})
			return
		}

		var user models.User
		err := db.Where("user_id = ?", req.UserID).First(&user).Error
		if err != nil {
			c.JSON(http.StatusNotFound, models.ErrorResponse{
				Error: models.APIError{Code: models.ErrNotFound, Message: "user not found"},
			})
			return
		}

		user.IsActive = req.IsActive
		db.Save(&user)

		c.JSON(http.StatusOK, gin.H{"user": user})
	}
}

func GetAssignedPRs(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.Query("user_id")
		if userID == "" {
			c.JSON(http.StatusBadRequest, models.ErrorResponse{
				Error: models.APIError{Code: models.ErrNotFound, Message: "missing user_id"},
			})
			return
		}

		var user models.User
		if db.Where("user_id = ?", userID).First(&user).Error != nil {
			c.JSON(http.StatusNotFound, models.ErrorResponse{
				Error: models.APIError{Code: models.ErrNotFound, Message: "user not found"},
			})
			return
		}

		var prs []models.PullRequest
		db.Joins("JOIN pr_reviewers ON pr_reviewers.pull_request_id = pull_requests.pull_request_id").
			Where("pr_reviewers.user_id = ?", userID).
			Find(&prs)

		out := make([]models.PullRequestShort, 0, len(prs))
		for _, p := range prs {
			out = append(out, models.PullRequestShort{
				PullRequestID:   p.PullRequestID,
				PullRequestName: p.PullRequestName,
				AuthorID:        p.AuthorID,
				Status:          p.Status,
			})
		}

		c.JSON(http.StatusOK, gin.H{
			"user_id":       userID,
			"pull_requests": out,
		})
	}
}

func CreatePR(db *gorm.DB) gin.HandlerFunc {
	type Req struct {
		PullRequestID   string `json:"pull_request_id"`
		PullRequestName string `json:"pull_request_name"`
		AuthorID        string `json:"author_id"`
	}

	return func(c *gin.Context) {
		var req Req
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, models.ErrorResponse{
				Error: models.APIError{Code: models.ErrNotFound, Message: "invalid json"},
			})
			return
		}

		var exists models.PullRequest
		if db.Where("pull_request_id = ?", req.PullRequestID).First(&exists).Error == nil {
			c.JSON(http.StatusConflict, models.ErrorResponse{
				Error: models.APIError{Code: models.ErrPRExists, Message: "PR already exists"},
			})
			return
		}

		var author models.User
		if db.Where("user_id = ?", req.AuthorID).First(&author).Error != nil {
			c.JSON(http.StatusNotFound, models.ErrorResponse{
				Error: models.APIError{Code: models.ErrNotFound, Message: "author not found"},
			})
			return
		}

		var team models.Team
		if db.Preload("Members").Where("team_name = ?", author.TeamName).First(&team).Error != nil {
			c.JSON(http.StatusNotFound, models.ErrorResponse{
				Error: models.APIError{
					Code:    models.ErrNotFound,
					Message: "team not found",
				},
			})
			return
		}

		candidates := make([]models.User, 0)
		for _, m := range team.Members {
			if m.IsActive && m.UserID != author.UserID {
				candidates = append(candidates, m)
			}
		}

		limit := 2
		if len(candidates) < limit {
			limit = len(candidates)
		}

		selected := candidates[:limit]

		now := time.Now().UTC()
		pr := models.PullRequest{
			PullRequestID:   req.PullRequestID,
			PullRequestName: req.PullRequestName,
			AuthorID:        req.AuthorID,
			Status:          "OPEN",
			CreatedAt:       &now,
		}
		db.Create(&pr)

		for _, u := range selected {
			db.Exec("INSERT INTO pr_reviewers (pull_request_id, user_id) VALUES (?, ?)",
				req.PullRequestID, u.UserID)
		}

		out := pr
		out.AssignedReviewers = make([]string, len(selected))
		for i, u := range selected {
			out.AssignedReviewers[i] = u.UserID
		}

		c.JSON(http.StatusCreated, gin.H{"pr": out})
	}
}

func MergePR(db *gorm.DB) gin.HandlerFunc {
	type Req struct {
		PullRequestID string `json:"pull_request_id"`
	}

	return func(c *gin.Context) {
		var req Req
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, models.ErrorResponse{
				Error: models.APIError{Code: models.ErrNotFound, Message: "invalid json"},
			})
			return
		}

		var pr models.PullRequest
		err := db.Where("pull_request_id = ?", req.PullRequestID).First(&pr).Error
		if err != nil {
			c.JSON(http.StatusNotFound, models.ErrorResponse{
				Error: models.APIError{Code: models.ErrNotFound, Message: "PR not found"},
			})
			return
		}

		if pr.Status == "MERGED" {
			loadReviewers(db, &pr)
			c.JSON(http.StatusOK, gin.H{"pr": pr})
			return
		}

		now := time.Now().UTC()
		pr.Status = "MERGED"
		pr.MergedAt = &now
		db.Save(&pr)

		loadReviewers(db, &pr)

		c.JSON(http.StatusOK, gin.H{"pr": pr})
	}
}

func ReassignReviewer(db *gorm.DB) gin.HandlerFunc {
	type Req struct {
		PullRequestID string `json:"pull_request_id"`
		OldUserID     string `json:"old_user_id"`
	}

	return func(c *gin.Context) {
		var req Req
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, models.ErrorResponse{
				Error: models.APIError{Code: models.ErrNotFound, Message: "invalid json"},
			})
			return
		}

		var pr models.PullRequest
		if db.Where("pull_request_id = ?", req.PullRequestID).First(&pr).Error != nil {
			c.JSON(http.StatusNotFound, models.ErrorResponse{
				Error: models.APIError{Code: models.ErrNotFound, Message: "PR not found"},
			})
			return
		}

		if pr.Status == "MERGED" {
			c.JSON(http.StatusConflict, models.ErrorResponse{
				Error: models.APIError{
					Code:    models.ErrPRMerged,
					Message: "cannot reassign on merged PR",
				},
			})
			return
		}

		loadReviewers(db, &pr)

		assignedMap := map[string]bool{}
		for _, uid := range pr.AssignedReviewers {
			assignedMap[uid] = true
		}

		if !assignedMap[req.OldUserID] {
			c.JSON(http.StatusConflict, models.ErrorResponse{
				Error: models.APIError{
					Code:    models.ErrNotAssigned,
					Message: "reviewer is not assigned",
				},
			})
			return
		}

		var old models.User
		db.Where("user_id = ?", req.OldUserID).First(&old)

		var teamMembers []models.User
		db.Where("team_name = ? AND is_active = ?", old.TeamName, true).Find(&teamMembers)

		candidates := make([]string, 0)
		for _, u := range teamMembers {
			if u.UserID != req.OldUserID && !assignedMap[u.UserID] {
				candidates = append(candidates, u.UserID)
			}
		}

		if len(candidates) == 0 {
			c.JSON(http.StatusConflict, models.ErrorResponse{
				Error: models.APIError{
					Code:    models.ErrNoCandidate,
					Message: "no active replacement candidate in team",
				},
			})
			return
		}

		newID := candidates[0]

		db.Exec("DELETE FROM pr_reviewers WHERE pull_request_id = ? AND user_id = ?",
			req.PullRequestID, req.OldUserID)

		db.Exec("INSERT INTO pr_reviewers (pull_request_id, user_id) VALUES (?, ?)",
			req.PullRequestID, newID)

		loadReviewers(db, &pr)

		c.JSON(http.StatusOK, gin.H{
			"pr":          pr,
			"replaced_by": newID,
		})
	}
}

func loadReviewers(db *gorm.DB, pr *models.PullRequest) {
	var reviewers []models.User
	db.Joins("JOIN pr_reviewers ON pr_reviewers.user_id = users.user_id").
		Where("pr_reviewers.pull_request_id = ?", pr.PullRequestID).
		Find(&reviewers)

	pr.Reviewers = reviewers
	pr.AssignedReviewers = make([]string, len(reviewers))
	for i, u := range reviewers {
		pr.AssignedReviewers[i] = u.UserID
	}
}
