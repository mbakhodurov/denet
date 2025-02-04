package referrer

import (
	"denet/internal/lib/api/response"
	"denet/internal/lib/logger/sl"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/go-playground/validator"
)

type Request struct {
	ReferalId int64 `json:"referalId" validate:"required"`
}

type Response struct {
	response.Response
	Message string `json:"message,omitempty"`
}

type ReferalTask interface {
	SetReferral(userID int64, referralID int64) error
}

func NewReferalTask(log *slog.Logger, referalTask ReferalTask) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.user.ReferalTask.New"

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

		referealid := req.ReferalId
		referalidInt64 := int64(referealid)

		err = referalTask.SetReferral(id, referalidInt64)
		if err != nil {
			log.Error("failed to enter referalid: ", sl.Err(err))
			render.JSON(w, r, response.Error("internal error: "+err.Error()))
			return
		}

		log.Info("referalid entered, added point : 5")

		render.JSON(w, r, Response{
			Response: response.OK(),
			Message:  "Successfully entered referalid, added point : 5",
		})
	}
}
