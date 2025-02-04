package task

import (
	"denet/internal/lib/api/response"
	"denet/internal/lib/logger/sl"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/go-playground/validator"
)

type Request struct {
	Points int64 `json:"points" validate:"required"`
}

type Response struct {
	response.Response
	Message string `json:"message,omitempty"`
}

type USERTask interface {
	CompleteTask(userID int64, taskPoints int64) error
}

func NewTask(log *slog.Logger, uSERTask USERTask) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.user.task.New"

		log = log.With(
			slog.String("op", op),
		)

		log.Info("Request received", slog.String("users", r.URL.String()))
		ids := chi.URLParam(r, "id")
		if ids == "" {
			log.Info("id is empty")
			render.JSON(w, r, response.Error("invalid request"))
			return
		}

		id, err := strconv.ParseInt(ids, 10, 64)
		if err != nil {
			log.Error("invalid id format", slog.String("id", ids))
			render.JSON(w, r, response.Error("invalid id format"))
			return
		}

		var req Request
		err = render.DecodeJSON(r.Body, &req)
		if err != nil {
			log.Error("failed to decode request body", sl.Err(err))
			render.JSON(w, r, response.Error("failed to decode request: "+err.Error()))
			return
		}
		log.Info("request body decoded", slog.Any("request", req))

		if err := validator.New().Struct(req); err != nil {
			validateErr := err.(validator.ValidationErrors)

			log.Error("invalid request", sl.Err(err))

			render.JSON(w, r, response.ValidationError(validateErr))
			return
		}

		points := req.Points
		pointsInt64 := int64(points)

		err = uSERTask.CompleteTask(id, pointsInt64)
		if err != nil {
			log.Error("failed to completeTask ", sl.Err(err))
			render.JSON(w, r, response.Error("internal error: "+err.Error()))
			return
		}

		log.Info("completed task, added point : ", slog.String("point:", fmt.Sprintf("%d", points)))

		render.JSON(w, r, Response{
			Response: response.OK(),
			Message:  "Successfully completed task, added point : " + fmt.Sprintf("%d", points),
		})
	}
}
