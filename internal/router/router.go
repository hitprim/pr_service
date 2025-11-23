package router

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)
import h "pr_service/internal/handlers"

func SetupRouter(db *gorm.DB) *gin.Engine {
	r := gin.Default()

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	team := r.Group("/team")
	team.POST("/add", h.TeamAdd(db))
	team.GET("/get", h.TeamGet(db))

	users := r.Group("/users")
	users.POST("/setIsActive", h.SetIsActive(db))
	users.GET("/getReview", h.GetAssignedPRs(db))

	pr := r.Group("/pullRequest")
	pr.POST("/create", h.CreatePR(db))
	pr.POST("/merge", h.MergePR(db))
	pr.POST("/reassign", h.ReassignReviewer(db))

	return r
}
