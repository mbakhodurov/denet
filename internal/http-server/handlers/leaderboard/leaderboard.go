package leaderboard

import (
	"denet/internal/lib/api/response"
	"denet/internal/lib/logger/sl"
	"denet/internal/lib/models"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/render"
)

type Response struct {
	Response response.Response `json:"response"`
	Users    []UserData        `json:"users"`
}

type UserData struct {
	Username    string    `json:"username,omitempty"`
	Points      int64     `json:"points,omitempty"`
	Referral_id int64     `json:"referral_id"`
	Created_at  time.Time `json:"created_at"`
}

type Leaderboard interface {
	GetLeaderboard() ([]models.User, error)
}

func NewLeaderboard(log *slog.Logger, leaderboard Leaderboard) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.url.Leaderboard.New"

		log := log.With(
			slog.String("op", op),
		)

		resLeaderboard, err := leaderboard.GetLeaderboard()
		if err != nil {
			log.Error("failed to get Leaderboard", sl.Err(err))

			render.JSON(w, r, response.Error("internal error"))

			return
		}

		for _, user := range resLeaderboard {
			log.Info("got user for leaderboard", slog.String("username", user.Username), slog.Int64("points", user.Points))
		}

		// log.Info("got Leaderboard", slog.String("Leaderboard", resLeaderboard[0].Username))
		var users []UserData
		for _, user := range resLeaderboard {
			users = append(users, UserData{
				Username:    user.Username,
				Points:      user.Points,
				Referral_id: user.Referral_id,
				Created_at:  user.Created_at,
			})
		}

		render.JSON(w, r, Response{
			Response: response.OK(),
			Users:    users,
		})
	}
}
