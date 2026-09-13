package task

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

func createHandler(service TaskService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req CreateTaskRequest

		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			http.Error(w, "invalid JSON", http.StatusBadRequest)
			return
		}

		task := &Task{
			MadeAt:                 time.Now(),
			Image:                  req.Image,
			EnvironmentPrepareCode: req.EnvironmentPrepareCode,
			ExecutionCode:          req.ExecutionCode,
			Source:                 req.Source,
			Maker:                  req.Maker,
			Status:                 TaskPending,
		}

		err = service.Create(r.Context(), task)

		if err != nil {
			http.Error(w, "failed to create task", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(task)

	}
}

func deleteHandler(service TaskService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := uuid.Parse(chi.URLParam(r, "id"))

		if err != nil {
			http.Error(w, "invalid task ID", http.StatusBadRequest)
		}

		err = service.Delete(r.Context(), id)
		if err != nil {
			http.Error(w, "failed to delete task", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

func getHandler(service TaskService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := uuid.Parse(chi.URLParam(r, "id"))
		if err != nil {
			http.Error(w, "invalid task ID", http.StatusBadRequest)
			return
		}
		user, err := service.Get(r.Context(), id)

		if err != nil {
			http.Error(w, "failed to get task", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(user)
	}
}

func NewTaskRouter(service TaskService) chi.Router {
	r := chi.NewRouter()

	r.Post("/", createHandler(service))
	r.Get("/{id}", getHandler(service))
	r.Delete("/{id}", deleteHandler(service))
	return r
}
