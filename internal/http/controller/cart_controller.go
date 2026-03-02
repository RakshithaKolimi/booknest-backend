package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"booknest/internal/domain"
)

type cartController struct {
	service domain.CartService
}

func NewCartController(service domain.CartService) domain.CartController {
	return &cartController{service: service}
}

func (c *cartController) RegisterRoutes(r gin.IRouter) {
	RegisterCartRoutes(r, getJWTConfig(), c)
}

// GetCart godoc
// @Summary      Get cart
// @Description  Fetches the authenticated user's cart with items and totals
// @Tags         Cart
// @Produce      json
// @Success      200  {object}  domain.CartView
// @Failure      401  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Security     BearerAuth
// @Router       /cart [get]
func (c *cartController) GetCart(ctx *gin.Context) {
	userID, err := getUserID(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	cart, err := c.service.GetCart(ctx, userID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, cart)
}

// AddItem godoc
// @Summary      Add item to cart
// @Description  Adds a book to the authenticated user's cart
// @Tags         Cart
// @Accept       json
// @Produce      json
// @Param        payload  body  domain.CartItemInput  true  "Cart item input"
// @Success      200  {object}  domain.CartView
// @Failure      400  {object}  map[string]string
// @Failure      401  {object}  map[string]string
// @Security     BearerAuth
// @Router       /cart/items [post]
func (c *cartController) AddItem(ctx *gin.Context) {
	userID, err := getUserID(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var input domain.CartItemInput
	if err := ctx.ShouldBindJSON(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := sanitizeInput(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	cart, err := c.service.AddItem(ctx, userID, input)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, cart)
}

// UpdateItem godoc
// @Summary      Update cart item
// @Description  Updates quantity for a cart item in the authenticated user's cart
// @Tags         Cart
// @Accept       json
// @Produce      json
// @Param        payload  body  domain.CartItemInput  true  "Cart item input"
// @Success      200  {object}  domain.CartView
// @Failure      400  {object}  map[string]string
// @Failure      401  {object}  map[string]string
// @Security     BearerAuth
// @Router       /cart/items [put]
func (c *cartController) UpdateItem(ctx *gin.Context) {
	userID, err := getUserID(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var input domain.CartItemInput
	if err := ctx.ShouldBindJSON(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := sanitizeInput(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	cart, err := c.service.UpdateItem(ctx, userID, input)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, cart)
}

// RemoveItem godoc
// @Summary      Remove item from cart
// @Description  Removes a book from the authenticated user's cart
// @Tags         Cart
// @Produce      json
// @Param        book_id  path  string  true  "Book ID"
// @Success      200  {object}  domain.CartView
// @Failure      400  {object}  map[string]string
// @Failure      401  {object}  map[string]string
// @Security     BearerAuth
// @Router       /cart/items/{book_id} [delete]
func (c *cartController) RemoveItem(ctx *gin.Context) {
	userID, err := getUserID(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	bookID, err := uuid.Parse(ctx.Param("book_id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid book id"})
		return
	}

	cart, err := c.service.RemoveItem(ctx, userID, bookID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, cart)
}

// ClearCart godoc
// @Summary      Clear cart
// @Description  Removes all items from the authenticated user's cart
// @Tags         Cart
// @Produce      json
// @Success      200  {object}  map[string]string
// @Failure      401  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Security     BearerAuth
// @Router       /cart/clear [post]
func (c *cartController) ClearCart(ctx *gin.Context) {
	userID, err := getUserID(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	if err := c.service.Clear(ctx, userID); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "cart cleared"})
}
