package handlers

import (
	"be/internal/models/response"
	"be/internal/repositories"
	"be/internal/services"
	"strconv"

	"github.com/gin-gonic/gin"
)

type OrderHandler interface {
	Create(c *gin.Context)
	Get(c *gin.Context)
}

type orderHandler struct {
	svc services.OrderService
}

func NewOrderHandler(svc services.OrderService) OrderHandler {
	return &orderHandler{svc: svc}
}

type createOrderReq struct {
	ProductID int64  `json:"product_id" binding:"required"`
	Quantity  int    `json:"quantity" binding:"required,min=1"`
	BuyerID   string `json:"buyer_id" binding:"required"`
}

func (h *orderHandler) Create(c *gin.Context) {
	var req createOrderReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	order, err := h.svc.Create(c.Request.Context(), req.ProductID, req.Quantity, req.BuyerID)
	if err != nil {
		if err == repositories.ErrOutOfStock {
			response.Conflict(c, "OUT_OF_STOCK")
			return
		}
		response.Internal(c, err.Error())
		return
	}
	response.Created(c, order)
}

func (h *orderHandler) Get(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid id")
		return
	}
	order, err := h.svc.Get(c.Request.Context(), id)
	if err != nil {
		response.NotFound(c, "not found")
		return
	}
	response.OK(c, order)
}
