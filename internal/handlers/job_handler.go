package handlers

import (
	"time"

	"github.com/gin-gonic/gin"

	"be/internal/models/response"
	"be/internal/repositories"
	"be/internal/services"
)

type JobHandler interface {
	StartSettlement(c *gin.Context)
	Get(c *gin.Context)
	Cancel(c *gin.Context)
}

type jobHandler struct {
	jobs repositories.JobRepository
	svc  services.JobService
}

func NewJobHandler(j repositories.JobRepository, s services.JobService) JobHandler {
	return &jobHandler{jobs: j, svc: s}
}

type settlementReq struct {
	From string `json:"from"`
	To   string `json:"to"`
}

func (h *jobHandler) StartSettlement(c *gin.Context) {
	var req settlementReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	from, err1 := time.Parse("2006-01-02", req.From)
	to, err2 := time.Parse("2006-01-02", req.To)
	if err1 != nil || err2 != nil {
		response.BadRequest(c, "invalid date format")
		return
	}
	jobID := "job_" + time.Now().Format("20060102150405")
	if err := h.svc.StartSettlement(c.Request.Context(), jobID, from, to); err != nil {
		response.Internal(c, err.Error())
		return
	}
	response.Created(c, gin.H{"job_id": jobID, "status": string(repositories.JobStatusQueued)})
}

func (h *jobHandler) Get(c *gin.Context) {
	id := c.Param("id")
	jr, err := h.jobs.Get(c.Request.Context(), id)
	if err != nil {
		response.NotFound(c, "not found")
		return
	}
	progress := int64(0)
	if jr.Total > 0 {
		progress = (jr.Processed * 100) / jr.Total
	}
	resp := gin.H{
		"job_id":    jr.ID,
		"status":    jr.Status,
		"processed": jr.Processed,
		"total":     jr.Total,
		"progress":  progress,
	}
	if jr.Status == repositories.JobStatusCompleted && jr.ResultPath.Valid {
		// Expose a simple local download path via static /downloads
		resp["download_url"] = "/downloads/" + id + ".csv"
	}
	response.OK(c, resp)
}

func (h *jobHandler) Cancel(c *gin.Context) {
	id := c.Param("id")
	if err := h.jobs.RequestCancel(c.Request.Context(), id); err != nil {
		response.NotFound(c, "not found")
		return
	}
	response.OK(c, gin.H{"job_id": id, "status": "CANCEL_REQUESTED"})
}
