package api

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/gin-gonic/gin"
	"github.com/lib/pq"
	"github.com/lib/pq/pqerror"

	m "github.com/reveliant/mabinouze/models"
)

type DrinkHandler struct {
  	Handler
}

// Read drink
func (h DrinkHandler) Get(c *gin.Context) {
	// Get drink
	model := h.getDrinkByID(c)
	if model == nil { return }

	// Verify tippler authorization on parent order
	if !h.verifyDrinkAuthorization(c, model.Drink) { return }

	// Handle HEAD request
	if c.Request.Method == http.MethodHead {
		c.Status(http.StatusNoContent)
		return
	}
	
	c.JSON(http.StatusOK, model.Drink.Clean())
}

// Create drink
func (h DrinkHandler) Post(c *gin.Context) {
	// Bind payload to drink model
	payload := new(m.Drink)
	if err := c.ShouldBind(&payload); err != nil {
		c.Error(err)
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Cannot parse drink"})
		return
	}
	if payload.OrderID == nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Missing order ID"})
		return
	}
	
	// Prepare new drink object from payload
	id := uuid.New()
	model := m.NewDrinkModel(h.DB, c.Request.Context())
	model.Drink.ID = &id
	model.Drink.OrderID = payload.OrderID
	model.Drink.Name = payload.Name
	model.Drink.Quantity = payload.Quantity

	// Verify tippler authorization on parent order
	if !h.verifyDrinkAuthorization(c, model.Drink) { return }

	// Create in database
	if err := model.Create(); err != nil {
		if pqErr := pq.As(err, pqerror.ForeignKeyViolation); pqErr != nil {
			c.AbortWithStatusJSON(http.StatusConflict, gin.H{"error": "Specified order does not exist", "order": model.Drink.OrderID})
			return
		}
		if pqErr := pq.As(err, pqerror.UniqueViolation); pqErr != nil {
			c.AbortWithStatusJSON(http.StatusConflict, gin.H{"error": "Drink's name already defined on order", "order": model.Drink.OrderID, "name": model.Drink.Name})
			return
		}
		c.Error(err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Internal server error while inserting drink", "data": model.Drink.Clean()})
		return
	}

	c.JSON(http.StatusCreated, model.Drink.Clean())
}

// Update drink quantity
func (h DrinkHandler) Put(c *gin.Context) {
	// Get drink
	model := h.getDrinkByID(c)
	if model == nil { return }

	// Verify tippler authorization on parent order
	if !h.verifyDrinkAuthorization(c, model.Drink) { return }

	// Bind payload to drink model
	payload := new(m.Drink)
	if err := c.ShouldBind(&payload); err != nil {
		c.Error(err)
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error":  "Cannot parse drink"})
		return
	}

	// Update or delete in database
	var err error
	if payload.Quantity > 0 {
		model.Drink.Quantity = payload.Quantity
		err = model.Update()
	} else {
		err = model.Delete()
	}
	if err != nil {
		c.Error(err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Internal server error while updating drink"})
		return
	}
	
	c.JSON(http.StatusOK, model.Drink.Clean())
}

// Delete drink
func (h DrinkHandler) Delete(c *gin.Context) {
	// Get drink
	model := h.getDrinkByID(c)
	if model == nil { return }

	// Verify tippler authorization on parent order
	if !h.verifyDrinkAuthorization(c, model.Drink) { return }

	// Delete from database
	if err := model.Delete(); err != nil {
		c.Error(err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Internal server error while deleting drink"})
		return
	}

	c.Status(http.StatusNoContent)
}