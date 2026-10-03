package api

import (
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/gin-gonic/gin"
	"github.com/lib/pq"
	"github.com/lib/pq/pqerror"

	m "github.com/reveliant/mabinouze/models"
	"github.com/reveliant/mabinouze/utils"
)

type RoundHandler struct {
  	Handler
}

// Common handler to get and return round
// 'resolveRound' defines how to get round from database
func (h RoundHandler) read(c *gin.Context, resolveRound func(c *gin.Context) *m.RoundModel) {
	// Read round
	model := resolveRound(c)
	if model == nil { return }
	
	// Handled HEAD request
	if c.Request.Method == http.MethodHead {
		c.Status(http.StatusNoContent)
		return
	}

	// Verify credentials if access token required
	if !h.verifyRoundAccessToken(c, model.Round) { return }

	// Send summary
	summary, err := model.Summary()
	if err != nil {
		c.Error(err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Internal server error while exporting round summary"})
		return
	}

	c.JSON(http.StatusOK, summary)
}

// Common handler to get and return round
// 'resolveRound' defines how to get round from database
func (h RoundHandler) read_details(c *gin.Context, resolveRound func(c *gin.Context) *m.RoundModel) {
	// Read round
	model := resolveRound(c)
	if model == nil { return }

	// Verify organizer password
	if !h.verifyRoundOrganizerPassword(c, model.Round) { return }

	details, err := model.Details()
	if err != nil {
		c.Error(err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Internal server error while exporting round details"})
		return
	}

	c.JSON(http.StatusOK, details)
}

// Get by UUID
func (h RoundHandler) Get(c *gin.Context) {
	h.read(c, h.getRoundByID)
}

// Get details by UUID
func (h RoundHandler) GetDetails(c *gin.Context) {
	h.read_details(c, h.getRoundByID)
}

// Search by name
func (h RoundHandler) Search(c *gin.Context) {
	h.read(c, h.getRoundByName)
}

// Search details by name
func (h RoundHandler) SearchDetails(c *gin.Context) {
	h.read_details(c, h.getRoundByName)
}


// Post from payload
func (h RoundHandler) Post(c *gin.Context) {
	// Bind payload to round model
	payload := new(m.Round)
	if err := c.ShouldBind(&payload); err != nil {
		c.Error(err)
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Cannot parse round"})
		return
	}
	
	// Prepare new drink object from payload
	model := m.NewRoundModel(h.DB, c.Request.Context())
	model.Round.ID = uuid.New()
	model.Round.Name = payload.Name
	model.Round.Description = payload.Description
	model.Round.Time = payload.Time.Round(time.Minute)
	model.Round.Expires = model.Round.Time.Add(time.Hour * 6)
	model.Round.Organizer = utils.Crypt(payload.Organizer)
	if payload.AccessToken != nil {
		access_token := utils.Crypt(*payload.AccessToken)
		model.Round.AccessToken = &access_token
	}

	// Create in database
	if err := model.Create(); err != nil {
		if pqErr := pq.As(err, pqerror.UniqueViolation); pqErr != nil {
			c.AbortWithStatusJSON(http.StatusConflict, gin.H{"error": "Round's name already defined", "name": model.Round.Name})
			return
		}
		c.Error(err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Internal server error while inserting round", "data": model.Round.Clean()})
		return
	}

	c.JSON(http.StatusCreated, model.Round.Clean())
}

func (h RoundHandler) Put(c *gin.Context) {
	// Read round
	model := h.getRoundByID(c)
	if model == nil { return }

	// Verify organizer password
	if !h.verifyRoundOrganizerPassword(c, model.Round) { return }

	// Bind payload to round model
	payload := new(m.Round)
	if err := c.ShouldBind(&payload); err != nil {
		c.Error(err)
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Cannot parse round"})
		return
	}

	// Bind updated properties
	model.Round.Description = payload.Description
	model.Round.Time = payload.Time.Round(time.Minute)

	if !model.Round.Locked {
		// Check if new round times is after expiration
		if model.Round.Time.After(model.Round.Expires) {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Requested time is after round expiration"})
			return
		}
	} else {
		model.Round.Expires = model.Round.Time.Add(time.Hour * 6)
	}
	
	// Update in database
	if err := model.Update(); err != nil {
		c.Error(err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Internal server error while updating round"})
		return
	}

	c.JSON(http.StatusOK, model.Round.Clean())
}

func (h RoundHandler) Delete(c *gin.Context) {
	// Read round
	model := h.getRoundByID(c)
	if model == nil { return }

	// Verify organizer password
	if !h.verifyRoundOrganizerPassword(c, model.Round) { return }

	if err := model.Delete(); err != nil {
		c.Error(err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Internal server error while deleting round"})
		return
	}

	c.Status(http.StatusNoContent)
}