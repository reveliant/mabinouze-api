package api

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/gin-gonic/gin"
	"github.com/lib/pq"
	"github.com/lib/pq/pqerror"

	m "github.com/reveliant/mabinouze/models"
	"github.com/reveliant/mabinouze/utils"
)

type OrderHandler struct {
  	Handler
}


// Common handler to get and return order
// 'resolveOrder' defines how to build order from request headers or payload
func (h OrderHandler) read(c *gin.Context, resolveOrder func() *m.OrderModel) {
	// Get order
	model := resolveOrder()
	if model == nil { return }
	
	// Handle HEAD request
	if c.Request.Method == http.MethodHead {
		c.Status(http.StatusNoContent)
		return
	}
	
	// Check tippler credentials
	if !h.verifyOrderCredentials(c, model.Order) { return }

	// Populate drinks
	model.ReadDrinks()

	c.JSON(http.StatusOK, model.Order.Clean())
}

// Get from order UUID
func (h OrderHandler) Get(c *gin.Context) {
	h.read(c, func() *m.OrderModel {
		return h.getOrderByID(c)
	})
}

// Get from round UUID
func (h OrderHandler) GetFromRound(c *gin.Context) {
	h.read(c, func() *m.OrderModel {
		creds := c.MustGet(AuthCredentials).(Credentials)
		round_id := c.MustGet("uuid").(uuid.UUID)

		// Prepare order model
		model := m.NewOrderModel(h.DB, c.Request.Context())
		model.Order.RoundID = &round_id
		model.Order.Tippler = creds.User

		// Get order
		if err := model.Read(); err != nil {
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "No such order", "round_id": *model.Order.RoundID, "name": model.Order.Tippler})
			return nil
		}
		return model
	})
}

// Get from round searched by name
func (h OrderHandler) GetFromRoundSearch(c *gin.Context) {
	h.read(c, func() *m.OrderModel {
		creds := c.MustGet(AuthCredentials).(Credentials)

		// Get round
		round_model := h.getRoundByName(c)
		if round_model == nil {
			return nil
		}

		// Prepare order model
		model := m.NewOrderModel(h.DB, c.Request.Context())
		model.Order.RoundID = &round_model.Round.ID
		model.Order.Tippler = creds.User
		
		// Get order
		if err := model.Search(); err != nil {
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "No such order", "round_id": *model.Order.RoundID, "name": model.Order.Tippler})
			return nil
		}
		return model
	})
}

// Common handler to create and return order
// 'resolveRound' defines how to get parent round from request headers or body
// 'buildOrder' defines how to build order from parent round, request header or body 
func (h OrderHandler) create(c *gin.Context, resolveRound func(c *gin.Context) *m.RoundModel) {
	// Bind payload to order model
	payload := new(m.Order)
	if err := c.ShouldBind(&payload); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if payload.Tippler == "" && payload.InTippler == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Missing tippler name"})
		return
	}

	// Read parent round
	round_model := resolveRound(c)
	if round_model == nil { return }

	// Verify credentials if access token required
	if !h.verifyRoundAccessToken(c, round_model.Round) { return }

	// Prepare new order object from payload 
	model := m.NewOrderModel(h.DB, c.Request.Context())
	model.Order.ID = uuid.New()
	model.Order.RoundID = &round_model.Round.ID
	model.Order.Tippler = payload.Tippler
	if payload.Tippler == "" {
		model.Order.Tippler = payload.InTippler
	}
	model.Order.Password = utils.Crypt(payload.Password)

	// Create order in database
	if err := model.Create(); err != nil {
		if pqErr := pq.As(err, pqerror.ForeignKeyViolation); pqErr != nil {
			c.AbortWithStatusJSON(http.StatusConflict, gin.H{"error": "Specified round does not exist", "order": model.Order.RoundID})
			return
		}
		if pqErr := pq.As(err, pqerror.UniqueViolation); pqErr != nil {
			c.AbortWithStatusJSON(http.StatusConflict, gin.H{"error": "Order's name already defined on round", "order": model.Order.RoundID, "name": model.Order.Tippler})
			return
		}
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error(), "data": model.Order.Clean()})
		return
	}

	c.JSON(http.StatusCreated, model.Order.Clean())
}

// Post from whole request body
func (h OrderHandler) Post(c *gin.Context) {
	payload := new(m.Order)
	if err := c.ShouldBind(&payload); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	h.create(c, func(c *gin.Context) *m.RoundModel {
		model := m.NewRoundModel(h.DB, c.Request.Context())
		model.Round.ID = *payload.RoundID

		if err := model.Read(); err != nil {
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "No such round", "id": model.Round.ID})
			return nil
		}
		return model
	})
}

// Post from round UUID, request headers and body
func (h OrderHandler) PostFromRound(c *gin.Context) {
	h.create(c, h.getRoundByID)
}

// Post from round search by name, request headers and body
func (h OrderHandler) PostFromRoundSearch(c *gin.Context) {
	h.create(c, h.getRoundByName)
}

// Delete order
func (h OrderHandler) Delete(c *gin.Context) {
	// Get order
	model := h.getOrderByID(c)
	if model == nil { return }
	
	// Check tippler credentials
	if !h.verifyOrderCredentials(c, model.Order) { return }

	// Delete from database
	if err := model.Delete(); err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}