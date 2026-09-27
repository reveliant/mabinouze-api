package models

import (
	"fmt"
	"context"
	"database/sql"
	"github.com/google/uuid"
)

//
// Drink object
//

type Drink struct {
	ID			*uuid.UUID	`json:"id,omitempty"`
	OrderID		*uuid.UUID	`json:"order_id,omitempty"`
	Name		string		`json:"name"`
	Quantity	uint16		`json:"quantity" default:1`
}

func NewDrink(order_id uuid.UUID, name string) *Drink {
	id := uuid.New()
	d := new(Drink)
	d.ID = &id
	d.OrderID = &order_id
	d.Name = name
	d.Quantity = 1
	return d
}

func (d Drink) String() string {
	return fmt.Sprintf("%s (%d)", d.Name, d.Quantity)
}

// Clean before export
func (d Drink) Clean() Drink {
	return d
}

//
// Drink database model
//

type DrinkModel struct {
	Connector
	Drink
}

func NewDrinkModel(db *sql.DB, ctx context.Context) *DrinkModel {
	c := new(DrinkModel) 
	c.DB = db
	c.Context = ctx
	return c
}

//
// Drink CRUD methods
//

// Create drink in database
func (m DrinkModel) Create() error {
	_, err := m.Exec("INSERT INTO drinks(drink_id, order_id, name, quantity) VALUES ($1, $2, $3, $4)", *m.Drink.ID, *m.Drink.OrderID, m.Drink.Name, m.Drink.Quantity)
	return err
}

// Read drink from database
func (m *DrinkModel) Read() error {
	row := m.QueryRow("SELECT order_id, name, quantity FROM drinks WHERE drink_id = $1", *m.Drink.ID)
	return row.Scan(&m.Drink.OrderID, &m.Drink.Name, &m.Drink.Quantity)
}

// Update drink quantity in database
func (m DrinkModel) Update() error {
	_, err := m.Exec("UPDATE drinks SET quantity = $1 WHERE drink_id = $2", m.Drink.Quantity, *m.Drink.ID)
	return err
}

// Delete drink in database
func (m DrinkModel) Delete() error {
	_, err := m.Exec("DELETE FROM drinks WHERE drink_id = $1", *m.Drink.ID)
	return err
}