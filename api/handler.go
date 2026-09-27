package api

import (
	"database/sql"
	"net/http"
	"log"
	
	"github.com/google/uuid"
	"github.com/gin-gonic/gin"

	m "github.com/reveliant/mabinouze/models"
)

type Handler struct {
  	DB *sql.DB
}

//
// Drink
//

// Get drink by UUID
func (h Handler) getDrinkByID(c *gin.Context) *m.DrinkModel {
	id := c.MustGet("uuid").(uuid.UUID)
	model := m.NewDrinkModel(h.DB, c.Request.Context())
	model.Drink.ID = &id

	if err := model.Read(); err != nil {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "No such drink", "id": model.Drink.ID})
		return nil
	}
	return model
}

func (h Handler) verifyDrinkAuthorization(c *gin.Context, drink m.Drink) bool {
	// Get parent order
	if drink.OrderID == nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Missing parent order ID"})
		return false
	}
	model := m.NewOrderModel(h.DB, c.Request.Context())
	model.Order.ID = *drink.OrderID
	
	if err := model.Read(); err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "No such order", "id": model.Order.ID})
		return false
	}

	// Check tippler credentials against order
	return h.verifyOrderCredentials(c, model.Order)
}

//
// Order
//

// Get order by UUID
func (h Handler) getOrderByID(c *gin.Context) *m.OrderModel {
	model := m.NewOrderModel(h.DB, c.Request.Context())
	model.Order.ID = c.MustGet("uuid").(uuid.UUID)

	if err := model.Read(); err != nil {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "No such order", "id": model.Order.ID})
		return nil
	}
	return model
}

// Check tippler credentials
func (h Handler) verifyOrderCredentials(c *gin.Context, order m.Order) bool {
	// Check tippler credentials against order
	if creds, found := c.Get(AuthCredentials); found {
		if err := order.VerifyCredentials(creds.(Credentials).User, creds.(Credentials).Password); err != nil {
			c.Header("WWW-Authenticate", "Bearer realm=\"mabinouze\", error=\"invalid_token\", error_description=\"Invalid credentials\"")
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return false
		}
		return true
	}

	// No tippler credentials, get parent round
	model := m.NewRoundModel(h.DB, c.Request.Context())
	model.Round.ID = *order.RoundID
	log.Println(model.Round.ID, *order.RoundID, order.ID)
	
	if err := model.Read(); err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "No such round", "id": model.Round.ID})
		return false
	}

	// Try organizer password on parent round
	return h.verifyRoundOrganizerPassword(c, model.Round)
}

//
// Round
//

// Get round by UUID
func (h Handler) getRoundByID(c *gin.Context) *m.RoundModel {
	model := m.NewRoundModel(h.DB, c.Request.Context())
	model.Round.ID = c.MustGet("uuid").(uuid.UUID)

	if err := model.Read(); err != nil {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "No such round", "id": model.Round.ID})
		return nil
	}
	return model
}

// Get round by name
func (h Handler) getRoundByName(c *gin.Context) *m.RoundModel {
	model := m.NewRoundModel(h.DB, c.Request.Context())
	model.Round.Name = c.Param("id")

	if err := model.Search(); err != nil {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "No such round", "name": model.Round.Name})
		return nil
	}
	return model
}

// Verify access token or organizer password
func (h Handler) verifyRoundAccessToken(c *gin.Context, round m.Round) bool {
	// Access granted if round does not require access token
	if !round.HasAccessToken() {
		return true
	}

	// Access denied if bearer token is missing
	val, found := c.Get(AuthBearer)
	if !found {
		c.Header("WWW-Authenticate", "Bearer realm=\"mabinouze\"")
		c.AbortWithStatus(http.StatusUnauthorized)
		return false
	}
	token := val.(string)

	// Access granted if bearer token matches access token
	if round.VerifyAccessToken(token) == nil {
		return true
	}
	
	// Access granted if bearer token matches organizer password
	if round.VerifyOrganizerPassword(token) == nil {
		return true
	}

	// Access denied
	c.Header("WWW-Authenticate", "Bearer realm=\"mabinouze\", error=\"invalid_token\", error_description=\"Invalid credentials\"")
	c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid access password"})
	return false
}

// Verify organizer password
func (h Handler) verifyRoundOrganizerPassword(c *gin.Context, round m.Round) bool {
	token := c.MustGet(AuthBearer).(string)

	// Access granted if bearer token matches organizer password
	if round.VerifyOrganizerPassword(token) == nil {
		return true
	}

	// Access denied
	c.Header("WWW-Authenticate", "Bearer realm=\"mabinouze\", error=\"invalid_token\", error_description=\"Invalid credentials\"")
	c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid organizer password"})
	return false
}