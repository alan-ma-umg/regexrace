package middlewares

import (
	"net/http"

	"github.com/spf13/viper"
	"github.com/thylong/regexrace/models"

	"context"
)

// MongoHandler insert Mgo.session in context and serve the request.
func MongoHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		dbSession := viper.Get("MONGO_SESSION").(models.Session).Copy()
		r = r.WithContext(
			context.WithValue(r.Context(), "db", dbSession))

		next.ServeHTTP(w, r)
		dbSession.Close()
	})
}
