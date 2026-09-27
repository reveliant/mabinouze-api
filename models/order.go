package models

import (
	"fmt"
	"errors"
	"context"
	"database/sql"
	"github.com/google/uuid"
	"github.com/reveliant/mabinouze/utils"
)

//
// Order object
//

type Order struct {
	ID			uuid.UUID	`json:"id"`
	RoundID		*uuid.UUID	`json:"round,omitempty"`
	Tippler		string		`json:"name"`
	InTippler	string		`json:"tippler,omitempty"` // Alt name for tippler field
	Drinks		[]Drink		`json:"drinks,omitempty"`
	Password	string		`json:"password,omitempty"`
}

// String representation
func (o Order) String() string {
	return fmt.Sprintf("%s: %s", o.Tippler, o.Drinks)
}

// Clean before export
func (o Order) Clean() Order {
	o.Password = ""
	return o
}

//
// Order database model
//

type OrderModel struct {
	Connector
	Order
}

func NewOrderModel(db *sql.DB, ctx context.Context) *OrderModel {
	c := new(OrderModel) 
	c.DB = db
	c.Context = ctx
	return c
}

//
// Order CRUD methods
//

// Create order in database
func (m OrderModel) Create() error {
	_, err := m.Exec("INSERT INTO orders(order_id, round_id, name, password) VALUES ($1, $2, $3, $4)", m.Order.ID, *m.Order.RoundID, m.Order.Tippler, utils.Crypt(m.Order.Password))
	return err
}

// Read an order from database
func (m *OrderModel) Read() error {
	row := m.QueryRow("SELECT round_id, name, password FROM orders WHERE order_id = $1", m.Order.ID)
	return row.Scan(&m.Order.RoundID, &m.Order.Tippler, &m.Order.Password)
}

// Update order in database
func (m OrderModel) Update() error {
	_, err := m.Exec("UPDATE orders SET name = $1, password = $2 WHERE order_id = $3", m.Order.Tippler, utils.Crypt(m.Order.Password), m.Order.ID)
	return err
}

// Delete order in database
func (m OrderModel) Delete() error {
	tx, err := m.Begin()
	if err != nil {
		return err
	}

	_, err = tx.Exec("DELETE FROM drinks WHERE order_id = $1", m.Order.ID)
	if err != nil {
		_ = tx.Rollback()
		return err
	}
	_, err = tx.Exec("DELETE FROM orders WHERE order_id = $1", m.Order.ID)
	if err != nil {
		_ = tx.Rollback()
		return err
	}
	return tx.Commit()
}

//
// Searches
//

// Read an order from database
func (m *OrderModel) Search() error {
	row := m.QueryRow("SELECT order_id, password FROM orders WHERE round_id = $1 AND name = $2", *m.Order.RoundID, m.Order.Tippler)
	return row.Scan(&m.Order.ID, &m.Order.Password)
}

// Read all orders for a given round from database
/*
func (m OrderModel) FromRound(round_id uuid.UUID, include_empty=False) []*Orders:
	orders := new([]*Order)
	rows := m.Query("SELECT round_id, order_id, name, password FROM orders WHERE round_id = $1", roundID)
	err := row.Scan(&order.RoundID, , &order.ID, &order.Tippler, &order.Password)
	if err != nil {
		return nil
	}
	return orders
	for res in cur.fetchall():
		order = cls(**res).read_drinks()
		if len(order.drinks) or include_empty:
			orders.append(order)
	return orders
*/

// Populate order with drinks
func (m *OrderModel) ReadDrinks() error {
	// Read all drinks for a given order in database
	rows, err := m.Query("SELECT drink_id, name, quantity FROM drinks WHERE order_id = $1", m.Order.ID)
	if err != nil {
        return err
    }
    defer rows.Close()

	for rows.Next() {
        var drink Drink
        if err := rows.Scan(&drink.ID, &drink.Name, &drink.Quantity); err != nil {
            return err
        }
        m.Order.Drinks = append(m.Order.Drinks, drink)
    }
    return rows.Err()
}

//
// Specific methods
//

// Verify a tippler credentials
func (o *Order) VerifyCredentials(username, password string) error {
	if username != o.Tippler {
		return errors.New("creds: Username mismatch")
	}
	return utils.Verify(password, o.Password)
}