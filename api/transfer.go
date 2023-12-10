package api

import (
	"fmt"
	"net/http"

	db "github.com/ernitingarg/golang-postgres-sqlc-bank-backend/db/sqlc"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/shopspring/decimal"
)

type createTransferRequest struct {
	FromAccountId int64           `json:"from_account_id" binding:"required,min=1"`
	ToAccountId   int64           `json:"to_account_id" binding:"required,min=1"`
	Amount        decimal.Decimal `json:"amount" binding:"required"`
	Currency      db.Currency     `json:"currency" binding:"required"`
}

func (server *Server) createTransferHandler(ctx *gin.Context) {
	var req createTransferRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	if !server.validateCurrency(ctx, req.FromAccountId, req.Currency) {
		return
	}

	if !server.validateCurrency(ctx, req.ToAccountId, req.Currency) {
		return
	}

	arg := db.TransferTxParams{
		FromAccountId: req.FromAccountId,
		ToAccountId:   req.ToAccountId,
		Amount:        req.Amount,
	}

	result, err := server.store.TransferTx(ctx, arg)
	if err != nil {
		if err == pgx.ErrNoRows {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return
		}

		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, result)
}

func (server *Server) validateCurrency(
	ctx *gin.Context,
	accountId int64,
	expectedCurrency db.Currency) bool {

	account, err := server.store.GetAccount(ctx, accountId)
	if err != nil {
		if err == pgx.ErrNoRows {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return false
		}

		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return false
	}

	if account.Currency != expectedCurrency {
		err := fmt.Errorf(
			"account [%d] currency mismatch, actual: %s, given: %s",
			accountId,
			account.Currency,
			expectedCurrency)

		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return false
	}

	return true
}
