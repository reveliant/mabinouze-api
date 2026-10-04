package models

import (
	"fmt"
	"context"
	"database/sql"
	"time"
	"github.com/google/uuid"
	"github.com/reveliant/mabinouze/utils"
)

//
// Round object
//

type Round struct {
	ID			uuid.UUID	`json:"id"`
	Name		string		`json:"name" binding:"max=255"`
	Description	string	 	`json:"description" binding:"required,max=255"`
	Time		time.Time	`json:"time" binding:"required"`
	Expires		time.Time	`json:"expires"`
	Organizer	string		`json:"password,omitempty"`
	AccessToken	*string		`json:"access_token,omitempty"`
	Locked		bool		`json:"-" default:false`
}

type RoundSummary struct {
	Round
	Tipplers	uint16		`json:"tipplers" default:1`
	Drinks		[]Drink		`json:"drinks"`
}

type RoundDetails struct {
	Round
	Tipplers	map[string]Order	`json:"tipplers"`
}

func NewRound(name, description string, when time.Time, password string) *Round {
	r := new(Round)
	r.ID = uuid.New()
	r.Name = name
	r.Description = description
	r.Time = when.Round(time.Minute)
	r.Expires = r.Time.Add(time.Hour * 6)
	r.Organizer = utils.Crypt(password)
	r.AccessToken = nil
	r.Locked = false
	return r
}

func (r Round) String() string {
	return fmt.Sprintf("%s (%s)", r.Name, r.Time)
}

// Clean before export
func (r Round) Clean() Round {
	r.Organizer = ""
	r.AccessToken = nil
	return r
}

//
// Round database model
//

type RoundModel struct {
	Connector
	Round
}

func NewRoundModel(db *sql.DB, ctx context.Context) *RoundModel {
	c := new(RoundModel) 
	c.DB = db
	c.Context = ctx
	return c
}

//
// Round CRUD methods
//

// Create round in database
func (m RoundModel) Create() error {
	_, err := m.Exec("INSERT INTO rounds(round_id, name, description, time, expires, password, access_token) VALUES ($1, $2, $3, $4, $5, $6, $7)", m.Round.ID, m.Round.Name, m.Round.Description, m.Round.Time, m.Round.Expires, m.Round.Organizer, m.Round.AccessToken)
	return err
}

// Read a round from database
func (m *RoundModel) Read() error {
	row := m.QueryRow("SELECT name, description, time, expires, password, access_token, locked FROM rounds WHERE round_id = $1", m.Round.ID)
	return row.Scan(&m.Round.Name, &m.Round.Description, &m.Round.Time, &m.Round.Expires, &m.Round.Organizer, &m.Round.AccessToken, &m.Round.Locked)
}

// Update round in database
func (m *RoundModel) Update() error {
	if !m.Round.Locked {
		m.Round.Expires = m.Round.Time.Add(time.Hour * 6)
	}
	_, err := m.Exec("UPDATE rounds SET name = $1, description = $2, time = $3, expires = $4, password = $5, access_token = $6 WHERE round_id = $7", m.Round.Name, m.Round.Description, m.Round.Time, m.Round.Expires, m.Round.Organizer, m.Round.AccessToken, m.Round.ID)
	return err
}

// Delete round in database
func (m RoundModel) Delete() error {
	tx, err := m.Begin()
	if err != nil {
		return err
	}

	_, err = tx.Exec("DELETE FROM drinks WHERE order_id IN (SELECT order_id FROM orders WHERE round_id = $1)", m.Round.ID)
	if err != nil {
		_ = tx.Rollback()
		return err
	}
	_, err = tx.Exec("DELETE FROM orders WHERE round_id = $1", m.Round.ID)
	if err != nil {
		_ = tx.Rollback()
		return err
	}
	_, err = tx.Exec("DELETE FROM rounds WHERE round_id = $1", m.Round.ID)
	if err != nil {
		_ = tx.Rollback()
		return err
	}
	return tx.Commit()
}

//
// Searches
//

// Search a round from database
func (m *RoundModel) Search() error {
	row := m.QueryRow("SELECT round_id, description, time, expires, password, access_token, locked FROM rounds WHERE name = $1", m.Round.Name)
	return row.Scan(&m.Round.ID, &m.Round.Description, &m.Round.Time, &m.Round.Expires, &m.Round.Organizer, &m.Round.AccessToken, &m.Round.Locked)
}

// Check if round exists in database
func (m RoundModel) Exists() error {
	row := m.QueryRow("SELECT round_id FROM rounds WHERE name = $1", m.Round.Name)
	return row.Err()
}

//
// Specific methods
//

// Populate order with drinks
func (m *RoundModel) Summary() (*RoundSummary, error) {
	// Read all drinks for a given round
	rows, err := m.Query("SELECT d.order_id, d.name, d.quantity FROM drinks AS d JOIN orders USING(order_id) WHERE orders.round_id = $1", m.Round.ID)
	if err != nil {
        return nil, err
    }
    defer rows.Close()

	tipplers := make(map[uuid.UUID]struct{})
	drinks := make(map[string]*Drink)

	for rows.Next() {
		var order_id uuid.UUID
        var drink Drink

        if err := rows.Scan(&order_id, &drink.Name, &drink.Quantity); err != nil {
            return nil, err
        }
		
		tipplers[order_id] = struct{}{}

		if _, found := drinks[drink.Name]; !found {
			drinks[drink.Name] = &drink
		} else {
			drinks[drink.Name].Quantity += drink.Quantity
		}
    }
	

	summary := new(RoundSummary)
	summary.Round = m.Round.Clean()
	summary.Tipplers = uint16(len(tipplers))
	summary.Drinks = make([]Drink, 0)
	for _, drink := range drinks {
		summary.Drinks = append(summary.Drinks, *drink)
    }

    return summary, rows.Err()
}

// Return a round details, including personnal informations
func (m *RoundModel) Details() (*RoundDetails, error) {
	// Read all orders for given round
	rows, err := m.Query("SELECT order_id, name FROM orders WHERE round_id = $1", m.Round.ID)
	if err != nil {
        return nil, err
    }
    defer rows.Close()

	// Build a map of orders referenced by their UUID
	orders := make(map[uuid.UUID]*Order)

	for rows.Next() {
        var order Order

        if err := rows.Scan(&order.ID, &order.Tippler); err != nil {
            return nil, err
        }

		orders[order.ID] = &order
    }
	if err := rows.Err(); err != nil {
		return nil, err
	}

	// Read all drinks for given round to populate orders
	drinks, err := m.Query("SELECT d.drink_id, d.order_id, d.name, d.quantity FROM drinks AS d JOIN orders USING(order_id) WHERE orders.round_id = $1", m.Round.ID)
	if err != nil {
        return nil, err
    }
    defer drinks.Close()

	for drinks.Next() {
        var drink Drink
		var order_id uuid.UUID

        if err := drinks.Scan(&drink.ID, &order_id, &drink.Name, &drink.Quantity); err != nil {
            return nil, err
        }

		// Associate drink to its order, by their UUID
		orders[order_id].Drinks = append(orders[order_id].Drinks, drink)
    }
	
	details := new(RoundDetails)
	details.Round = m.Round.Clean()
	details.Tipplers = make(map[string]Order)
	// Build a map of orders referenced by tipplers name
	for _, order := range orders {
		details.Tipplers[order.Tippler] = *order
    }

	return details, drinks.Err()
}

// Verify an organizer password
func (r *Round) VerifyOrganizerPassword(token string) error {
	return utils.Verify(token, r.Organizer)
}

// Verify an access token
func (r *Round) VerifyAccessToken(token string) error {
	return utils.Verify(token, *r.AccessToken)
}

// Check if round requires an access token
func (r *Round) HasAccessToken() bool {
	return r.AccessToken != nil
}

// Check if round is locked
func (r *Round) IsLocked() bool {
	return r.Locked
}