package v1

import (
	"github.com/gin-gonic/gin"

	"booknest/internal/domain"
)

const Version = "v1"

// Router owns all v1 route registrations.
type Router struct {
	user      domain.UserController
	book      domain.BookController
	author    domain.AuthorController
	category  domain.CategoryController
	publisher domain.PublisherController
	cart      domain.CartController
	order     domain.OrderController
}

func NewRouter(
	user domain.UserController,
	book domain.BookController,
	author domain.AuthorController,
	category domain.CategoryController,
	publisher domain.PublisherController,
	cart domain.CartController,
	order domain.OrderController,
) *Router {
	return &Router{
		user:      user,
		book:      book,
		author:    author,
		category:  category,
		publisher: publisher,
		cart:      cart,
		order:     order,
	}
}

func (r *Router) Version() string {
	return Version
}

func (r *Router) RegisterRoutes(group *gin.RouterGroup) {
	r.user.RegisterRoutes(group)
	r.book.RegisterRoutes(group)
	r.author.RegisterRoutes(group)
	r.category.RegisterRoutes(group)
	r.publisher.RegisterRoutes(group)
	r.cart.RegisterRoutes(group)
	r.order.RegisterRoutes(group)
}
