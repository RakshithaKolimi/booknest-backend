package routes

// ====================
// User Routes
// ====================
const (
	HealthRoute = "/health"

	BooksRoute       = "/books"
	BookRoute        = "/book"
	BookIDRoute      = "/book/:id"
	BookReviewsRoute = "/books/:id/reviews"

	UsersRoute           = "/users"
	UserRoute            = "/user/:id"
	UserPreferencesRoute = "/user/:id/preferences"

	ForgotPassword       = "/forgot-password"
	LoginRoute           = "/login"
	RefreshRoute         = "/refresh"
	RegisterRoute        = "/register"
	RegisterAdminRoute   = "/register-admin"
	VerifyEmailRoute     = "/verify-email"
	VerifyMobileRoute    = "/verify-mobile"
	ResendEmailRoute     = "/resend-email-verification"
	ResendMobileOTPRoute = "/resend-mobile-otp"
	ResetPasswordRoute   = "/reset-password"
	ResetPasswordByToken = "/reset-password/confirm"

	CartRoute      = "/cart"
	CartItemsRoute = "/cart/items"
	CartItemRoute  = "/cart/items/:book_id"
	CartClearRoute = "/cart/clear"

	OrdersRoute           = "/orders"
	OrderCheckoutRoute    = "/orders/checkout"
	OrderConfirmRoute     = "/orders/confirm"
	OrderCancelRoute      = "/orders/cancel"
	AdminOrdersRoute      = "/admin/orders"
	AdminOrderStatusRoute = "/admin/orders/status"

	AuthorsRoute    = "/authors"
	AuthorByIDRoute = "/authors/:id"

	CategoriesRoute   = "/categories"
	CategoryByIDRoute = "/categories/:id"
)

// ====================
// Publisher Routes
// ====================
const (
	PublisherRoute       = "/publishers"
	PublisherByIDRoute   = "/publishers/:id"
	PublisherStatusRoute = "/publishers/:id/status"
)
