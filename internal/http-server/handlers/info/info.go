package info

import (
	"denet/internal/lib/api/response"
	"denet/internal/lib/logger/sl"
	"denet/internal/lib/models"
	"denet/internal/storage"
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
)

type USERInfo interface {
	GetUSER(id int64) (*models.User, error)
}

type Response struct {
	response.Response
	Username    string    `json:"username,omitempty"`
	Points      int64     `json:"points,omitempty"`
	Referral_id int64     `json:"referral_id"`
	Created_at  time.Time `json:"created_at"`
}

func NewUserInfo(log *slog.Logger, uSERInfo USERInfo) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.url.uSERInfo.New"

		log := log.With(
			slog.String("op", op),
		)
		log.Info("Request received", slog.String("user", r.URL.String()))
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

		resUSER, err := uSERInfo.GetUSER(id)
		if errors.Is(err, storage.ErrUserNotFound) {
			log.Info("user not found", "id", ids)
			render.JSON(w, r, response.Error("user not found"))
			return
		}
		if err != nil {
			log.Error("failed to get user", sl.Err(err))

			render.JSON(w, r, response.Error("internal error"))

			return
		}

		log.Info("got user", slog.String("user", resUSER.Username))

		render.JSON(w, r, Response{
			Response:    response.OK(),
			Username:    resUSER.Username,
			Points:      resUSER.Points,
			Referral_id: resUSER.Referral_id,
			Created_at:  resUSER.Created_at,
		})
	}
}
