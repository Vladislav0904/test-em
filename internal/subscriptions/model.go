package subscriptions

import (
	"github.com/google/uuid"
	"time"
)

type Subscription struct {
	ID          uuid.UUID  `db:"id" json:"id"`
	ServiceName string     `db:"service_name" json:"service_name" `
	Price       int        `db:"price" json:"price"`
	UserID      uuid.UUID  `db:"user_id" json:"user_id"`
	StartDate   time.Time  `db:"start_date" json:"start_date"`
	EndDate     *time.Time `db:"end_date" json:"end_date,omitempty"`
}
