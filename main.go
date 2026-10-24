package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type Property struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	Location  string    `json:"location"`
	Price     float64   `json:"price"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

type apiResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

var properties = []Property{
	{ID: 1, Name: "Villa Maju", Location: "Bandung", Price: 2500000000, Status: "Available", CreatedAt: time.Now()},
	{ID: 2, Name: "Apartemen Harmoni", Location: "Jakarta", Price: 1800000000, Status: "Booked", CreatedAt: time.Now()},
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", healthHandler)
	mux.HandleFunc("/api/properties", propertiesHandler)
	mux.HandleFunc("/api/properties/", propertyDetailHandler)

	server := &http.Server{
		Addr:              ":8080",
		Handler:           withCORS(mux),
		ReadHeaderTimeout: 5 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       15 * time.Second,
	}

	fmt.Println("Server started on http://localhost:8080")
	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatal(err)
	}
}

func withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, apiResponse{Success: false, Message: "Method not allowed"})
		return
	}

	writeJSON(w, http.StatusOK, apiResponse{
		Success: true,
		Message: "Backend is running",
		Data: map[string]string{
			"status": "ok",
			"time":   time.Now().Format(time.RFC3339),
		},
	})
}

func propertiesHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		writeJSON(w, http.StatusOK, apiResponse{Success: true, Data: properties})
	case http.MethodPost:
		var payload Property
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			writeJSON(w, http.StatusBadRequest, apiResponse{Success: false, Message: "Invalid JSON body"})
			return
		}

		if strings.TrimSpace(payload.Name) == "" || strings.TrimSpace(payload.Location) == "" {
			writeJSON(w, http.StatusBadRequest, apiResponse{Success: false, Message: "Name and location are required"})
			return
		}

		payload.ID = nextPropertyID()
		payload.CreatedAt = time.Now()
		properties = append(properties, payload)

		writeJSON(w, http.StatusCreated, apiResponse{Success: true, Message: "Property created", Data: payload})
	default:
		writeJSON(w, http.StatusMethodNotAllowed, apiResponse{Success: false, Message: "Method not allowed"})
	}
}

func propertyDetailHandler(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/properties/")
	if path == "" || path == "/api/properties" {
		writeJSON(w, http.StatusNotFound, apiResponse{Success: false, Message: "Property not found"})
		return
	}

	id, err := strconv.Atoi(path)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, apiResponse{Success: false, Message: "Invalid property ID"})
		return
	}

	for i := range properties {
		if properties[i].ID == id {
			switch r.Method {
			case http.MethodGet:
				writeJSON(w, http.StatusOK, apiResponse{Success: true, Data: properties[i]})
			case http.MethodPut, http.MethodPatch:
				var payload Property
				if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
					writeJSON(w, http.StatusBadRequest, apiResponse{Success: false, Message: "Invalid JSON body"})
					return
				}

				if strings.TrimSpace(payload.Name) == "" || strings.TrimSpace(payload.Location) == "" {
					writeJSON(w, http.StatusBadRequest, apiResponse{Success: false, Message: "Name and location are required"})
					return
				}

				properties[i].Name = payload.Name
				properties[i].Location = payload.Location
				properties[i].Price = payload.Price
				properties[i].Status = payload.Status
				writeJSON(w, http.StatusOK, apiResponse{Success: true, Message: "Property updated", Data: properties[i]})
			case http.MethodDelete:
				properties = append(properties[:i], properties[i+1:]...)
				writeJSON(w, http.StatusOK, apiResponse{Success: true, Message: "Property deleted"})
			default:
				writeJSON(w, http.StatusMethodNotAllowed, apiResponse{Success: false, Message: "Method not allowed"})
			}
			return
		}
	}

	writeJSON(w, http.StatusNotFound, apiResponse{Success: false, Message: "Property not found"})
}

func nextPropertyID() int {
	maxID := 0
	for _, property := range properties {
		if property.ID > maxID {
			maxID = property.ID
		}
	}
	return maxID + 1
}

func writeJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		log.Printf("failed to encode response: %v", err)
	}
}
