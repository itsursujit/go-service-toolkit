package app

import (
	"github.com/gofiber/fiber"
	"github.com/jinzhu/gorm"
	"net/http"
	"os"
	"time"
	"toolkit/app/cache"
	"toolkit/app/observance"
	"toolkit/app/toolkit"
)

// User holds all basic user information.
type User struct {
	ID   uint64 `json:"id"`
	Name string `json:"name" validate:"required"`
	Age  uint64 `json:"age" validate:"gte=18"`
}

func Serve() {
	// Load environment variables.
	toolkit.MustLoadEnvs("")

	// Set up observance (logging).
	obsConfig := toolkit.ObsConfig{
		AppName:  os.Getenv("APP_NAME"),
		LogLevel: os.Getenv("LOG_LEVEL"),
		// SentryURL:            os.Getenv("SENTRY_URL"),
		Version: os.Getenv("APP_VERSION"),
		// MetricsURL:           os.Getenv("METRICS_URL"),
		MetricsFlushInterval: 1 * time.Second,
		LoggedHeaders: map[string]string{
			"FastBill-RequestId": "requestId",
		},
	}
	obs := toolkit.MustNewObs(obsConfig)
	defer obs.PanicRecover()

	// Set up DB connection and run migrations.
	dbConfig := toolkit.DBConfig{
		Dialect:  os.Getenv("DB_DIALECT"),
		Host:     os.Getenv("DATABASE_HOST"),
		Port:     os.Getenv("DATABASE_PORT"),
		User:     os.Getenv("DATABASE_USER"),
		Password: os.Getenv("DATABASE_PASSWORD"),
		Name:     os.Getenv("DATABASE_NAME"),
	}
	db := toolkit.MustSetupDB(dbConfig, obs.Logger)
	defer func() {
		if err := db.Close(); err != nil {
			obs.Logger.WithError(err).Error("failed to close DB connection")
		}
	}()

	toolkit.MustEnsureDBMigrations("migrations", dbConfig)

	// Set up REDIS newCache.
	newCache := toolkit.MustNewCache(os.Getenv("REDIS_HOST"), os.Getenv("REDIS_PORT"), "testPrefix")
	defer func() {
		if err := newCache.Close(); err != nil {
			obs.Logger.WithError(err).Error("failed to close REDIS connection")
		}
	}()

	// Set up the server.
	addr := ":" + os.Getenv("PORT")
	startFiberRoutes(addr, obs, db, newCache)

}

func startFiberRoutes(addr string, obs *observance.Obs, db *gorm.DB, cache cache.Cache) {
	e, _ := toolkit.MustNewFiberServer(obs)

	// Set up a routes and handlers.
	e.Post("/users", func(c *fiber.Ctx) {
		obs.Logger.Info("incoming request to create new user")

		newUser := &User{}
		err := c.Locals("newUser", newUser)
		if err != nil {
			_ = c.JSON(map[string]string{"msg": err.(string)})
		}

		if err = db.Save(newUser).Error; err != nil {
			_ = c.JSON(map[string]string{"msg": err.(string)})
		}

		// Nonsense cache usage example
		err = cache.SetJSON("latestNewUser", newUser, 0)
		if err != nil {
			_ = c.JSON(map[string]string{"msg": err.(string)})
		}

		c.SendStatus(http.StatusCreated)
	})
	e.Get("/users", func(c *fiber.Ctx) {
		obs.Logger.Info("incoming request to to list al; users")
		users := &[]User{}
		db.Find(&users)
		_ = c.JSON(users)
		c.SendStatus(http.StatusOK)
	})
	e.Get("/", func(c *fiber.Ctx) {
		_ = c.Render("index", fiber.Map{})
	})
	_ = e.Listen(addr)
}
