package api


import (
	"bank/utils"
	repository "bank/db/sqlc"
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgconn"
)

type CreateUserRequest struct {
	Username       string `json:"username" binding:"required,alphanum"`
	Password       string `json:"password" binding:"required,min=6,max=100"`
	FullName       string `json:"full_name" binding:"required"`
	Email          string `json:"email" binding:"required,email"`
}

func (server *Server) createUser(ctx *gin.Context) {
	var req CreateUserRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	arg := repository.CreateUserParams{
		Username:       req.Username,
		HashedPassword: hashedPassword,
		FullName:       req.FullName,
		Email:          req.Email,
	}

	user, err := server.store.CreateUser(ctx, arg)
	if err != nil {
		if pgErr, ok := err.(*pgconn.PgError); ok {
			switch pgErr.SQLState() {
			case "23505": // unique_violation
				ctx.JSON(http.StatusConflict, errorResponse(err))
				return
			}
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, user)
}

type GetUserRequest struct {
	Username string `uri:"username" binding:"required,alphanum"`
}

func (server *Server) getUser(ctx *gin.Context) {
	var req GetUserRequest
	if err := ctx.ShouldBindUri(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	user, err := server.store.GetUser(ctx, req.Username)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, user)
}