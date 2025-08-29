package models

import (
	"time"

	"gorm.io/gorm"
)

// Order representa o modelo de dados para a tabela 'orders'
type Order struct {
	gorm.Model
	ID            uint      `gorm:"primaryKey" json:"id"`
	Customer_name string    `json:"customer_name"`
	Product_name  string    `json:"product_name"`
	Quantity      int       `json:"quantity"`
	CreatedAt     time.Time `json:"createdAt"`
}
