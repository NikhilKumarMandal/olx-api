package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/NikhilKumarMandal/olx-api/internal/config"
	"github.com/NikhilKumarMandal/olx-api/internal/httpx"
	"github.com/NikhilKumarMandal/olx-api/internal/middleware"
	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"golang.org/x/crypto/bcrypt"
)

var dummyHash = []byte("$2a$10$bLRdw4JkAFgk5dE50Uj.BOqdB4TI6LH9k6Kr0mdsDfIHo5DLGNxqi")
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
	cfg config.Config
}

func NewAuthHandler(db *sql.DB, logger *slog.Logger,cfg config.Config) *AuthHandler {
	return &AuthHandler{
		db:     db,
		logger: logger,
		cfg: cfg,
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

func(ah AuthHandler) Singin(w http.ResponseWriter, r *http.Request) {

	ctx := r.Context()
	requestId := middleware.RequestIdFromContext(ctx)
	log := ah.logger.With("request_id",requestId)

	var req SigninRequest
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

	var u user
	row := ah.db.QueryRowContext(ctx,
	`SELECT id, email, password FROM users WHERE email = $1`, req.Email)

	err := row.Scan(&u.ID,&u.Email,&u.Password)
	
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// hash 
			_ = bcrypt.CompareHashAndPassword(dummyHash,[]byte(req.Password))

			httpx.Error(w,http.StatusUnauthorized,"email or password don't match",httpx.CodeUnauthenticated)
			return
		}

		log.Error("find user by email failed","err",err)
		httpx.Error(w,http.StatusInternalServerError,"something went wrong",httpx.CodeInternalError)
		return
	}

	if err = bcrypt.CompareHashAndPassword([]byte(u.Password),[]byte(req.Password)); err != nil {
		log.Warn("password mismatch","user_id",u.ID)
		httpx.Error(w,http.StatusUnauthorized,"email or password don't match",httpx.CodeUnauthenticated)
		return
	}

	tokenTTL := 24 * time.Hour 
	now := time.Now()
	claims := jwt.RegisteredClaims{
		Subject: u.ID,
		IssuedAt: jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(now.Add(tokenTTL)),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256,claims)

	signed , err := token.SignedString([]byte(ah.cfg.JwtKey))

	if err != nil {
		log.Error("jwt signed failed","err",err)
		httpx.Error(w,http.StatusInternalServerError,"something went wrong",httpx.CodeInternalError)
		return
	}


	out := SigninResponse{
		Token: signed,
		ExpiresIn: int(tokenTTL.Seconds()),
	}

	ah.logger.Info("new user logged in","user_id",u.ID)

	w.Header().Set("Content-Type","application/json")
	w.WriteHeader(http.StatusOK)

	_ = json.NewEncoder(w).Encode(out)
}