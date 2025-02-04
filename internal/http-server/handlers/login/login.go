package login

import (
	middlewares "denet/internal/http-server/middleware"
	"denet/internal/lib/api/response"
	"denet/internal/lib/logger/sl"
	"denet/internal/lib/models"
	"denet/internal/storage"
	"errors"
	"log/slog"
	"net/http"

	"github.com/go-chi/render"
	"github.com/go-playground/validator"
)

type USERLogin interface {
	LoginUser(username string, password string) (*models.User, error)
}

type Request struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
}

type Response struct {
	response.Response
	Token string `json:"token,omitempty"`
}

func NewLogin(log *slog.Logger, uSERLogin USERLogin) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.url.uSERInfo.New"

		log := log.With(
			slog.String("op", op),
		)
		log.Info("Request received", slog.String("user", r.URL.String()))

		var req Request
		err := render.DecodeJSON(r.Body, &req)
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

		username := req.Username
		passwords := req.Password

		resUSER, err := uSERLogin.LoginUser(username, passwords)
		if errors.Is(err, storage.ErrUserNotFound) {
			log.Info("user not found", "id", username)
			render.JSON(w, r, response.Error("user not found"))
			return
		}
		if err != nil {
			log.Error("failed to get user", sl.Err(err))

			render.JSON(w, r, response.Error("internal error: "+err.Error()))

			return
		}

		token, err := middlewares.GenerateJWT(resUSER.Username)
		if err != nil {
			render.JSON(w, r, response.Error("Failed to generate token"))
			return
		}

		log.Info("got user", slog.String("user", resUSER.Username))

		render.JSON(w, r, Response{
			Response: response.OK(),
			Token:    token,
		})
	}
}
