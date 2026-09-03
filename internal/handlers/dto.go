package handlers

import (
	"strings"
	"time"
	"fmt"
)

type CreateListingRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Price       int64  `json:"price"`
	City        string `json:"city"`
}

type CreateListingResponse struct {
	ID    string `json:"id"`
	Title string `json:"title"`
	CreatedAt time.Time `json:"created_at"`
}

type ValidationError struct {
	Field string
	Msg string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("%s: %s",e.Field,e.Msg)
}

func (req CreateListingRequest) validate() error {
	if strings.TrimSpace(req.Title) == "" {
		return &ValidationError{Field:"title",Msg: "Must not be empty"}
	}

	//Todo: add validation for other remaining field

	return nil
}