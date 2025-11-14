package api

import (
	"net/http"
	"push_service/models"

	"github.com/gin-gonic/gin"
)

// HealthHandler godoc
// @Summary      Get service metrics
// @Description  Gets message stats for the service
// @Tags         health
// @Accept       json
// @Produce      json
// @Success      200  {object}  models.ConsumerMetrics  "Service metrics"
// @Router       /health [get] // Assuming this is your health check path
func HealthHandler(c *models.Consumer, ctx *gin.Context) {
	ctx.JSON(http.StatusOK, c.ConsumerMetrics)
}
