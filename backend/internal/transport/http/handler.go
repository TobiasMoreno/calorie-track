package httptransport

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"calorietrack/backend/internal/application"
	"calorietrack/backend/internal/domain"
)

type Handler struct {
	service       *application.Service
	allowedOrigin string
}

func NewHandler(service *application.Service, allowedOrigin string) *Handler {
	return &Handler{service: service, allowedOrigin: allowedOrigin}
}

func (h *Handler) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/health", h.health)
	mux.HandleFunc("GET /api/settings", h.getSettings)
	mux.HandleFunc("PUT /api/settings", h.updateSettings)
	mux.HandleFunc("GET /api/dashboard", h.dashboard)
	mux.HandleFunc("GET /api/foods", h.listFoods)
	mux.HandleFunc("PATCH /api/foods/{id}", h.updateFood)
	mux.HandleFunc("POST /api/meal-entries", h.createMealEntry)
	mux.HandleFunc("DELETE /api/meal-entries/{id}", h.deleteMealEntry)
	mux.HandleFunc("PATCH /api/tasks/{id}", h.updateTask)
	mux.HandleFunc("POST /api/tasks", h.createTask)
	return h.recover(h.cors(mux))
}

func (h *Handler) health(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *Handler) getSettings(w http.ResponseWriter, r *http.Request) {
	settings, err := h.service.GetSettings(r.Context())
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, settings)
}

func (h *Handler) updateSettings(w http.ResponseWriter, r *http.Request) {
	var input domain.UserSettings
	if err := decodeJSON(r, &input); err != nil {
		writeError(w, err)
		return
	}
	settings, err := h.service.UpdateSettings(r.Context(), input)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, settings)
}

func (h *Handler) dashboard(w http.ResponseWriter, r *http.Request) {
	date := r.URL.Query().Get("date")
	result, err := h.service.Dashboard(r.Context(), date)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func (h *Handler) listFoods(w http.ResponseWriter, r *http.Request) {
	foods, err := h.service.ListFoods(r.Context(), r.URL.Query().Get("q"))
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, foods)
}

func (h *Handler) updateFood(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r.PathValue("id"))
	if err != nil {
		writeError(w, err)
		return
	}
	var input domain.FoodReferenceUpdate
	if err := decodeJSON(r, &input); err != nil {
		writeError(w, err)
		return
	}
	food, err := h.service.UpdateFood(r.Context(), id, input)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, food)
}

func (h *Handler) createMealEntry(w http.ResponseWriter, r *http.Request) {
	var input domain.CreateMealEntryInput
	if err := decodeJSON(r, &input); err != nil {
		writeError(w, err)
		return
	}
	entry, err := h.service.CreateMealEntry(r.Context(), input)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, entry)
}

func (h *Handler) deleteMealEntry(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r.PathValue("id"))
	if err != nil {
		writeError(w, err)
		return
	}
	if err := h.service.DeleteMealEntry(r.Context(), id); err != nil {
		writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) updateTask(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r.PathValue("id"))
	if err != nil {
		writeError(w, err)
		return
	}
	var input struct {
		Status domain.TaskStatus `json:"status"`
	}
	if err := decodeJSON(r, &input); err != nil {
		writeError(w, err)
		return
	}
	if err := h.service.UpdateTaskStatus(r.Context(), id, input.Status); err != nil {
		writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) createTask(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Date string `json:"date"`
		Name string `json:"name"`
	}
	if err := decodeJSON(r, &input); err != nil {
		writeError(w, err)
		return
	}
	task, err := h.service.CreateTask(r.Context(), input.Date, input.Name)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, task)
}

func parseID(value string) (int64, error) {
	id, err := strconv.ParseInt(value, 10, 64)
	if err != nil || id <= 0 {
		return 0, fmt.Errorf("%w: id inválido", domain.ErrValidation)
	}
	return id, nil
}

func decodeJSON(r *http.Request, target any) error {
	decoder := json.NewDecoder(io.LimitReader(r.Body, 1<<20))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return fmt.Errorf("%w: JSON inválido", domain.ErrValidation)
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return fmt.Errorf("%w: el cuerpo debe contener un único objeto JSON", domain.ErrValidation)
	}
	return nil
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func writeError(w http.ResponseWriter, err error) {
	status := http.StatusInternalServerError
	code := "internal_error"
	message := "Ocurrió un error inesperado"
	switch {
	case errors.Is(err, domain.ErrValidation):
		status, code, message = http.StatusBadRequest, "validation_error", cleanMessage(err)
	case errors.Is(err, domain.ErrNotFound):
		status, code, message = http.StatusNotFound, "not_found", "No se encontró el registro"
	case errors.Is(err, domain.ErrConflict):
		status, code, message = http.StatusConflict, "conflict", cleanMessage(err)
	}
	writeJSON(w, status, map[string]any{"error": map[string]string{"code": code, "message": message}})
}

func cleanMessage(err error) string {
	parts := strings.SplitN(err.Error(), ": ", 2)
	if len(parts) == 2 {
		return parts[1]
	}
	return err.Error()
}

func (h *Handler) cors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin == h.allowedOrigin {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Vary", "Origin")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
			w.Header().Set("Access-Control-Allow-Methods", "GET,POST,PUT,PATCH,DELETE,OPTIONS")
		}
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (h *Handler) recover(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if recover() != nil {
				writeError(w, errors.New("panic"))
			}
		}()
		next.ServeHTTP(w, r)
	})
}
