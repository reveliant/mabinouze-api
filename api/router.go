package api

import (
	"database/sql"
	"github.com/gin-gonic/gin"
)

func DatabaseMiddleware(db *sql.DB) gin.HandlerFunc {
  return func(c *gin.Context) {
    c.Set("db", db)
    c.Next()
  }
}

func NewRouter(db *sql.DB, debug bool) *gin.Engine {
	// Disable debug mode
	if (!debug) {
		gin.SetMode(gin.ReleaseMode)
	}
	
	// Set the router as the default one shipped with Gin
	router := gin.Default()
	
	// Plug database to router
  	router.Use(DatabaseMiddleware(db))

	// Setup Security Headers
	router.Use(SecurityHeaders())

	// /v1
	v1 := router.Group("/v1")
	{
		// /v1/drink
		drink_hdlr := DrinkHandler{Handler{DB: db}}
		drinks := v1.Group("/drink")
		drinks.Use(TipplerAuthRequired())

		drinks.OPTIONS("/:uuid", Preflight("OPTIONS, HEAD, GET, PUT, DELETE"))
		drinks.HEAD("/:uuid", ParseUUID(), drink_hdlr.Get)
		drinks.GET("/:uuid", ParseUUID(), drink_hdlr.Get)
		drinks.PUT("/:uuid", ParseUUID(), drink_hdlr.Put)
		drinks.DELETE("/:uuid", ParseUUID(), drink_hdlr.Delete)

		drinks.OPTIONS("", Preflight("OPTIONS, POST"))
		drinks.POST("", drink_hdlr.Post)

		// /v1/order
		order_hdlr := OrderHandler{Handler{DB: db}}
		orders := v1.Group("/order")
		orders.Use(TipplerAuthRequired())

		orders.OPTIONS("/:uuid", Preflight("OPTIONS, HEAD, GET, DELETE"))
		orders.HEAD("/:uuid", ParseUUID(), order_hdlr.Get)
		orders.GET("/:uuid", ParseUUID(), order_hdlr.Get)
		orders.DELETE("/:uuid", ParseUUID(), order_hdlr.Delete)

		orders.OPTIONS("", Preflight("OPTIONS, POST"))
		orders.POST("", order_hdlr.Post)

		// /v1/round
		round_hdlr := RoundHandler{Handler{DB: db}}
		rounds := v1.Group("/round")

		rounds.OPTIONS("/:uuid", Preflight("OPTIONS, HEAD, GET, PUT, DELETE"))
		rounds.HEAD("/:uuid", ParseUUID(), round_hdlr.Get)
		rounds.GET("/:uuid", ParseUUID(), round_hdlr.Get)
		rounds.PUT("/:uuid", ParseUUID(), AdminAuthRequired(), round_hdlr.Put)
		rounds.DELETE("/:uuid", ParseUUID(), AdminAuthRequired(), round_hdlr.Delete)

		rounds.OPTIONS("", Preflight("OPTIONS, POST"))
		rounds.POST("", round_hdlr.Post)

		rounds.OPTIONS("/:uuid/details", Preflight("OPTIONS, GET"))
		rounds.GET("/:uuid/details", ParseUUID(), AdminAuthRequired(), round_hdlr.GetDetails)

		rounds.OPTIONS("/:uuid/order", Preflight("OPTIONS, GET, DELETE"))
		rounds.GET("/:uuid/order", ParseUUID(), TipplerAuthRequired(), order_hdlr.GetFromRound)
		rounds.POST("/:uuid/order", ParseUUID(), TipplerAuthRequired(), order_hdlr.PostFromRound)

		// /v1/search
		search := v1.Group("/search")

		search.OPTIONS("/:id", Preflight("OPTIONS, HEAD, GET"))
		search.HEAD("/:id", round_hdlr.Search)
		search.GET("/:id", round_hdlr.Search)

		search.OPTIONS("/:id/details", Preflight("OPTIONS, GET"))
		search.GET("/:id/details", AdminAuthRequired(), round_hdlr.SearchDetails)

		search.OPTIONS("/:id/order", Preflight("OPTIONS, GET, POST"))
		search.GET("/:id/order", TipplerAuthRequired(), order_hdlr.GetFromRoundSearch)
		search.POST("/:id/order", TipplerAuthRequired(), order_hdlr.PostFromRoundSearch)
	}

	return router
}