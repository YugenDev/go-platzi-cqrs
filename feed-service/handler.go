package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/YugenDev/go-platzi-CQRS/events"
	"github.com/YugenDev/go-platzi-CQRS/models"
	"github.com/YugenDev/go-platzi-CQRS/repository"
	"github.com/segmentio/ksuid"
)

type CreateFeedRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
}

func CreateFeedHandler(w http.ResponseWriter, r *http.Request) {
	var req CreateFeedRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	createdAt := time.Now().UTC()
	id, err := ksuid.NewRandom()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	feed := models.Feed{
		ID:          id.String(),
		Title:       req.Title,
		Description: req.Description,
		CreatedAt:   createdAt,
	}

	if err := repository.InsertFeed(r.Context(), &feed); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := events.PublishCreatedFeed(r.Context(), &feed); err != nil {
		log.Fatal(err)
	}

	w.WriteHeader(http.StatusCreated)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(feed)

}
