package save

import (
	resp "denet/internal/lib/api/response"
	"denet/internal/lib/logger/sl"
	"denet/internal/lib/random"
	"denet/internal/storage"
	"errors"
	"log/slog"
	"net/http"

	"github.com/go-chi/render"
	"github.com/go-playground/validator"
)

type Request struct {
	Username    string `json:"username" validate:"required"`
	Password    string `json:"password" validate:"required,min=5"`
	Points      int64  `json:"points"`
	Referral_id int64  `json:"referral_id"`
}

type Response struct {
	resp.Response
	Username string `json:"username,omitempty"`
}

type USERSaver interface {
	SaveUser(username, password string, points, referral_id int64) (int64, error)
}

func New(log *slog.Logger, userSaver USERSaver) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.user.save.New"

		log = log.With(
			slog.String("op", op),
		)

		var req Request
		err := render.DecodeJSON(r.Body, &req)
		if err != nil {
			log.Error("failed to decode request body", sl.Err(err))
			render.JSON(w, r, resp.Error("failed to decode request: "+err.Error()))
			return
		}
		log.Info("request body decoded", slog.Any("request", req))

		if err := validator.New().Struct(req); err != nil {
			validateErr := err.(validator.ValidationErrors)

			log.Error("invalid request", sl.Err(err))

			render.JSON(w, r, resp.ValidationError(validateErr))
			return
		}

		username := req.Username
		password := req.Password
		points := req.Points
		referral_id := req.Referral_id
		if points == 0 {
			points = random.NewRandomInt()
		}
		if referral_id == 0 {
			referral_id = random.NewRandomInt()
		}

		id, err := userSaver.SaveUser(username, password, points, referral_id)
		if errors.Is(err, storage.ErrUserExists) {
			log.Info("user already exists", slog.String("user", req.Username))
			render.JSON(w, r, resp.Error("user already exists"))
			return
		}

		if err != nil {
			log.Error("failed to add user", sl.Err(err))
			render.JSON(w, r, resp.Error("failed to add user"))
			return
		}

		log.Info("user added", slog.Int64("id", id))
		render.JSON(w, r, Response{
			Response: resp.OK(),
			Username: username,
		})
	}
}
