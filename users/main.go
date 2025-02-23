package main

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/labstack/echo-contrib/prometheus"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
	"github.com/labstack/gommon/random"
	"github.com/vod/users/config"
	"github.com/vod/users/handlers"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	CorrelationID = "x-Request-ID"
)

var (
	c      *mongo.Client
	db     *mongo.Database
	usrcol *mongo.Collection
	cfg    config.Properties
)

// addCorrelationID is a custom middleware function.
// This method will generate a 20 digit requestID and
// added to header of both request and response for traceability.
// Instead of X-Request-ID a custom id X-Correlation-ID is generated.
func addCorrelationID(next echo.HandlerFunc) echo.HandlerFunc {
	return func(ctx echo.Context) error {
		//Generate the Correlation ID
		id := ctx.Request().Header.Get(CorrelationID)
		var cID string
		if id == "" {
			cID = random.String(20)
		} else {
			cID = id
		}
		ctx.Request().Header.Set(CorrelationID, cID)
		ctx.Response().Header().Set(CorrelationID, cID)
		return next(ctx)
	}
}

func init() {
	if err := cleanenv.ReadEnv(&cfg); err != nil {
		log.Fatalf("Configurations cannot be read: %v", err)
	}
	serverAPIOptions := options.ServerAPI(options.ServerAPIVersion1)
	// Build the URI from configuration
	uri := fmt.Sprintf(cfg.DBURL, cfg.DBUser, cfg.DBPass)
	// If the URI does not start with "mongodb://" or "mongodb+srv://", prepend "mongodb://"
	if !strings.HasPrefix(uri, "mongodb://") && !strings.HasPrefix(uri, "mongodb+srv://") {
		uri = "mongodb://" + uri
	}

	clientOptions := options.Client().
		ApplyURI(uri).
		SetServerAPIOptions(serverAPIOptions)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	var err error
	c, err = mongo.Connect(ctx, clientOptions)
	if err != nil {
		log.Fatalf("unable to connect to database: %v", err)
	}
	db = c.Database(cfg.DBName)
	usrcol = db.Collection(cfg.UserCollection)

}

func main() {
	e := echo.New()
	e.Use(middleware.Recover())
	e.Logger.SetLevel(log.INFO)
	e.Pre(middleware.RemoveTrailingSlash())
	e.Pre(addCorrelationID)
	e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{Format: `{"time":"${time_rfc3339_nano}","remote_ip":"${remote_ip}",` +
		`"request_ID":"${header:x-Request-ID}"+"host":"${host}","method":"${method}","uri":"${uri}","user_agent":"${user_agent}",` +
		`"status":${status},"error":"${error}","latency_human":"${latency_human}"` +
		`,"bytes_in":${bytes_in},"bytes_out":${bytes_out}}` + "\n"}))

	uh := &handlers.UsersHandler{Col: usrcol}
	p := prometheus.NewPrometheus("echo", nil)
	p.Use(e)
	e.GET("/", uh.Healthz)

	e.POST("/users", uh.CreateUser)
	e.POST("/auth", uh.AuthnUser)
	e.Logger.Infof("listening for requests on %s:%s", cfg.Host, cfg.Port)
	e.Logger.Fatal(e.Start(fmt.Sprintf("%s:%s", cfg.Host, cfg.Port)))
}
