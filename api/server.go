package api

import (
	db "github.com/ernitingarg/golang-postgres-sqlc-bank-backend/db/sqlc"
	"github.com/gin-gonic/gin"
)

// Server serves HTTP requests
type Server struct {
	store  db.Store
	router *gin.Engine
}

// NewServer creates an instance of Server and setup routing
func NewServer(store db.Store) *Server {
	server := &Server{
		store: store,
	}

	router := gin.Default()

	// add routes to the router
	router.POST("/api/accounts", server.createAccountHandler)
	router.GET("/api/accounts/:id", server.getAccountHandler)
	router.GET("/api/accounts", server.listAccountsHandler)
	router.PUT("/api/accounts", server.updateAccountsHandler)
	router.DELETE("/api/accounts/:id", server.deleteAccountsHandler)

	router.POST("/api/transfers", server.createTransferHandler)

	router.POST("/api/users", server.createUserHandler)

	server.router = router

	return server
}

// Start runs HTTP server on a specific address
func (server *Server) Start(address string) error {
	return server.router.Run(address)
}

func errorResponse(err error) gin.H {
	return gin.H{
		"error": err.Error(),
	}
}
