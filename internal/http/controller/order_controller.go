package controller

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"booknest/internal/domain"
)

type orderController struct {
	service domain.OrderService
}

func NewOrderController(service domain.OrderService) domain.OrderController {
	return &orderController{service: service}
}

func (c *orderController) RegisterRoutes(r gin.IRouter) {
	RegisterOrderRoutes(r, getJWTConfig(), c)
}

// Checkout godoc
// @Summary      Checkout order
// @Description  Creates an order from the authenticated user's cart
// @Tags         Orders
// @Accept       json
// @Produce      json
// @Param        payload  body  domain.CheckoutInput  true  "Checkout input"
// @Success      201  {object}  domain.OrderView
// @Failure      400  {object}  map[string]string
// @Failure      401  {object}  map[string]string
// @Security     BearerAuth
// @Router       /orders/checkout [post]
func (c *orderController) Checkout(ctx *gin.Context) {
	userID, err := getUserID(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var input domain.CheckoutInput
	if err := ctx.ShouldBindJSON(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	order, err := c.service.Checkout(ctx, userID, input)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusCreated, order)
}

// ConfirmPayment godoc
// @Summary      Confirm payment
// @Description  Confirms payment for an existing order
// @Tags         Orders
// @Accept       json
// @Produce      json
// @Param        payload  body  domain.PaymentConfirmInput  true  "Payment confirmation input"
// @Success      200  {object}  domain.OrderView
// @Failure      400  {object}  map[string]string
// @Failure      401  {object}  map[string]string
// @Security     BearerAuth
// @Router       /orders/confirm [post]
func (c *orderController) ConfirmPayment(ctx *gin.Context) {
	userID, err := getUserID(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var input domain.PaymentConfirmInput
	if err := ctx.ShouldBindJSON(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	order, err := c.service.ConfirmPayment(ctx, userID, input)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, order)
}

// ListMyOrders godoc
// @Summary      List my orders
// @Description  Lists orders for the authenticated user
// @Tags         Orders
// @Produce      json
// @Param        limit   query  int  false  "Result limit"
// @Param        offset  query  int  false  "Result offset"
// @Success      200  {array}  domain.OrderView
// @Failure      401  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Security     BearerAuth
// @Router       /orders [get]
func (c *orderController) ListMyOrders(ctx *gin.Context) {
	userID, err := getUserID(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	limit := 10
	offset := 0
	if v := ctx.Query("limit"); v != "" {
		limit, _ = strconv.Atoi(v)
	}
	if v := ctx.Query("offset"); v != "" {
		offset, _ = strconv.Atoi(v)
	}

	orders, err := c.service.ListUserOrders(ctx, userID, limit, offset)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, orders)
}

// ListAllOrders godoc
// @Summary      List all orders
// @Description  Lists all orders (admin only)
// @Tags         Orders
// @Produce      json
// @Param        limit   query  int  false  "Result limit"
// @Param        offset  query  int  false  "Result offset"
// @Success      200  {array}  domain.OrderView
// @Failure      500  {object}  map[string]string
// @Security     BearerAuth
// @Router       /admin/orders [get]
func (c *orderController) ListAllOrders(ctx *gin.Context) {
	limit := 10
	offset := 0
	if v := ctx.Query("limit"); v != "" {
		limit, _ = strconv.Atoi(v)
	}
	if v := ctx.Query("offset"); v != "" {
		offset, _ = strconv.Atoi(v)
	}

	orders, err := c.service.ListAllOrders(ctx, limit, offset)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, orders)
}
