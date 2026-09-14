package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/NikhilKumarMandal/olx-api/internal/httpx"
	"github.com/NikhilKumarMandal/olx-api/internal/middleware"
	"github.com/jackc/pgx/v5/pgconn"
	"golang.org/x/crypto/bcrypt"
)

type user struct{
	ID string `json:"id"`
	Name string `json:"name"`
	Email string `json:"email"`
	Password string `json:"password"`
	CreatedAt time.Time `json:"created_at"`
}

type AuthHandler struct {
	db     *sql.DB
	logger *slog.Logger
}

func NewAuthHandler(db *sql.DB, logger *slog.Logger) *AuthHandler {
	return &AuthHandler{
		db:     db,
		logger: logger,
	}
}

func(ah AuthHandler) Singup(w http.ResponseWriter, r *http.Request) {


	ctx := r.Context()
	requestId := middleware.RequestIdFromContext(ctx)
	log := ah.logger.With("request_id",requestId)

	var req SignupRequest
	if err :=json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Error("failed to decode","err",err)
		httpx.Error(w,http.StatusBadRequest,"invalid body","Code malformed json")
		return
	}

	if err := req.Validate(); err != nil {
		var verr *ValidationError
		errors.As(err,&verr)
		httpx.ValidationError(w,http.StatusUnprocessableEntity,err.Error(),"Validation failed",verr.Field)
		return
	}


	hash,err := bcrypt.GenerateFromPassword([]byte(req.Password),10)
	if err != nil {
		log.Error("hashing failed","err",err)
		httpx.Error(w,http.StatusInternalServerError,"something went wrong",httpx.CodeInternalError)
		return
	}

	row :=ah.db.QueryRowContext(ctx,
		`INSERT INTO users (name,email,password) VALUES ($1, $2, $3) RETURNING id, created_at`,req.Name,req.Email,hash)

	var u user
	err = row.Scan(&u.ID,&u.CreatedAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505"{
			httpx.Error(w,http.StatusConflict,"email already taken",httpx.CodeConflict)
			return
		}

		ah.logger.Error("scanning failed","err",err)
		httpx.Error(w,http.StatusInternalServerError,"something went wrong",httpx.CodeInternalError)
		return
	}

	out := SignupResponse{
		ID: u.ID,
		CreatedAt: u.CreatedAt,
	}

	ah.logger.Info("new user register","user_id",out.ID)

	w.Header().Set("Content-Type","application/json")
	w.WriteHeader(http.StatusCreated)

	_ = json.NewEncoder(w).Encode(out)
}