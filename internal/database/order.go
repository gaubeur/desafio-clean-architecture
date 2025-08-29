package database

import (
	"database/sql"
	"time"
)

type Order struct {
	Db            *sql.DB
	ID            int64     `json:"id"`
	Customer_name string    `json:"customer_name"`
	Product_name  string    `json:"product_name"`
	Quantity      int64     `json:"quantity"`
	CreatedAt     time.Time `json:"createdAt"`
}

func NewOrder(db *sql.DB) *Order {
	return &Order{Db: db}
}

func (o *Order) Create(customer_name string, product_name string, quantity int64) (Order, error) {
	_, err := o.Db.Exec("INSERT INTO orders (customer_name, product_name, quantity) VALUES (?, ?, ?)",
		customer_name, product_name, quantity)
	if err != nil {
		return Order{}, err
	}

	return Order{Customer_name: customer_name, Product_name: product_name, Quantity: quantity}, nil
}

func (o *Order) ListOrders() ([]Order, error) {
	rows, err := o.Db.Query("SELECT id, customer_name, product_name, quantity, createdAt FROM orders ORDER BY id DESC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var createdAt sql.NullTime

	var orders []Order
	for rows.Next() {
		var ord Order
		if err := rows.Scan(&ord.ID, &ord.Customer_name, &ord.Product_name, &ord.Quantity, &createdAt); err != nil {
			return nil, err
		}

		if createdAt.Valid {
			ord.CreatedAt = createdAt.Time
		} else {
			ord.CreatedAt = time.Time{}
		}
		orders = append(orders, ord)
	}

	return orders, nil
}
