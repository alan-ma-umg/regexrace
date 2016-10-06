package main

import (
	"net/http"
	"time"

	mgo "gopkg.in/mgo.v2"

	"github.com/spf13/viper"
	"github.com/thylong/regexrace/config"
	"github.com/thylong/regexrace/handlers"
	"github.com/thylong/regexrace/middlewares"
	"github.com/thylong/regexrace/models"

	"github.com/justinas/alice"
)

func main() {
	config.LoadConfig()

	// Create and store a Mongo session for every requests.
	session, err := mgo.Dial(viper.GetString("MONGO_URI"))
	if err != nil {
		panic(err)
	}
	session.SetSafe(&mgo.Safe{})
	session.SetSyncTimeout(3 * time.Second)
	session.SetSocketTimeout(3 * time.Second)
	viper.Set("MONGO_SESSION", session)
	defer session.Close()

	// Ensure data looks like expected.
	models.PrepareDB(session)
	models.EnsureQuestionData(session)
	models.EnsureScoreData(session)

	// Middlewares triggered for every requests.
	c := alice.New(
		middlewares.LoggingHandler,
		middlewares.TimeoutHandler,
		middlewares.AccessLogHandler,
		middlewares.MongoHandler,
	)
	if viper.GetString("ENV") != "dev" {
		c.Append(middlewares.PanicRecoveryHandler) // Has to be the latest middleware.
	}

	// Register Handlers.
	http.Handle("/", c.ThenFunc(handlers.HomeHandler))
	http.Handle("/status", c.ThenFunc(handlers.StatusHandler))
	http.Handle("/leaderboard", c.ThenFunc(handlers.LeaderboardHandler))
	http.Handle("/auth", c.ThenFunc(handlers.AuthHandler))
	http.Handle("/score", c.ThenFunc(handlers.ScoreHandler))
	// Serve css and js static files.
	http.Handle("/static/", c.ThenFunc(handlers.StaticHandler))
	http.Handle("/robots.txt", c.ThenFunc(handlers.RobotsHandler))
	// Following Handlers requiring auth.
	c = c.Append(middlewares.WithAuth)
	http.Handle("/answer", c.ThenFunc(handlers.AnswerHandler))
	http.ListenAndServe(":8080", nil)
}
