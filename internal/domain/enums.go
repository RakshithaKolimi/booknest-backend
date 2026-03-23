package domain

type UserRole string // @name UserRole

const (
	UserRoleUser  UserRole = "USER"
	UserRoleAdmin UserRole = "ADMIN"
)

type PaymentStatus string // @name PaymentStatus

const (
	PaymentPending         PaymentStatus = "PENDING"
	PaymentPaid            PaymentStatus = "PAID"
	PaymentRefundInitiated PaymentStatus = "REFUND_INITIATED"
	PaymentRefunded        PaymentStatus = "REFUNDED"
	PaymentFailed          PaymentStatus = "FAILED"
)

type PaymentMethod string

const (
	PaymentCOD        PaymentMethod = "COD"
	PaymentCreditCard PaymentMethod = "CREDIT_CARD"
	PaymentDebitCard  PaymentMethod = "DEBIT_CARD"
	PaymentNetBanking PaymentMethod = "NET_BANKING"
	PaymentUPI        PaymentMethod = "UPI"
)

type OrderStatus string

const (
	OrderPending   OrderStatus = "PENDING"
	OrderFailed    OrderStatus = "FAILED"
	OrderCancelled OrderStatus = "CANCELLED"
	OrderCompleted OrderStatus = "COMPLETED"
)
