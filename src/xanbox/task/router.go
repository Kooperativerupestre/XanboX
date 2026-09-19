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

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
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

		dbID, dockerID, err, deleteErr := service.Create(
			r.Context(),
			task,
		)

		if err != nil {
			if deleteErr != nil {
				http.Error(
					w,
					"failed to create task and failed to clean up execution",
					http.StatusInternalServerError,
				)
				return
			}

			http.Error(
				w,
				"failed to create task",
				http.StatusInternalServerError,
			)
			return
		}

		response := struct {
			ID       uuid.UUID `json:"id"`
			DockerID string    `json:"docker_id"`
		}{
			ID:       dbID,
			DockerID: dockerID,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)

		if err := json.NewEncoder(w).Encode(response); err != nil {
			return
		}
	}
}

func getHandler(service TaskService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := uuid.Parse(chi.URLParam(r, "id"))
		if err != nil {
			http.Error(w, "invalid task ID", http.StatusBadRequest)
			return
		}

		task, err := service.Get(r.Context(), id)
		if err != nil {
			http.Error(w, "failed to get task", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")

		if err := json.NewEncoder(w).Encode(task); err != nil {
			return
		}
	}
}

func syncHandler(service TaskService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := uuid.Parse(chi.URLParam(r, "id"))
		if err != nil {
			http.Error(w, "invalid task ID", http.StatusBadRequest)
			return
		}

		var req SyncTaskRequest

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid JSON", http.StatusBadRequest)
			return
		}

		if err := service.Sync(r.Context(), id, req.DockerID); err != nil {
			http.Error(w, "failed to sync task", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}

func NewTaskRouter(service TaskService) chi.Router {
	r := chi.NewRouter()

	r.Post("/", createHandler(service))
	r.Get("/{id}", getHandler(service))
	r.Post("/{id}/sync", syncHandler(service))

	return r
}