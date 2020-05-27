package main

import (
	"github.com/gofiber/fiber"
	"github.com/jinzhu/gorm"
	"html/template"
	"io"
	"net/http"
	"os"
	"time"
	"toolkit/cache"
	"toolkit/observance"
	"toolkit/toolkit"

	"github.com/labstack/echo/v4"
)

// User holds all basic user information.
type User struct {
	ID   uint64 `json:"id"`
	Name string `json:"name" validate:"required"`
	Age  uint64 `json:"age" validate:"gte=18"`
}
type Users []User

func main() {
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

	// Set up REDIS cache.
	cache := toolkit.MustNewCache(os.Getenv("REDIS_HOST"), os.Getenv("REDIS_PORT"), "testPrefix")
	defer func() {
		if err := cache.Close(); err != nil {
			obs.Logger.WithError(err).Error("failed to close REDIS connection")
		}
	}()

	// Set up the server.
	startEchoRoutes(obs, db, cache)
	// startFiberRoutes(obs, db, cache)

}

// TemplateRenderer is a custom html/template renderer for Echo framework
type TemplateRenderer struct {
	templates *template.Template
}

// Render renders a template document
func (t *TemplateRenderer) Render(w io.Writer, name string, data interface{}, c echo.Context) error {

	// Add global methods if data is a map
	if viewContext, isMap := data.(map[string]interface{}); isMap {
		viewContext["reverse"] = c.Echo().Reverse
	}

	return t.templates.ExecuteTemplate(w, name, data)
}

func startEchoRoutes(obs *observance.Obs, db *gorm.DB, cache cache.Cache) {
	e, connectionsClosed := toolkit.MustNewEchoServer(obs, "*")
	e.Static("/assets", "assets")
	renderer := &TemplateRenderer{
		templates: template.Must(template.ParseGlob("resources/templates/*.html")),
	}
	e.Renderer = renderer
	// Set up a routes and handlers.
	e.POST("/users", func(c echo.Context) error {
		obs.Logger.Info("incoming request to create new user")

		newUser := &User{}
		err := c.Bind(newUser)
		if err != nil {
			obs.Logger.WithError(err).Warn("invalid request")
			return c.JSON(http.StatusBadRequest, map[string]string{"msg": err.Error()})
		}

		if err = db.Save(newUser).Error; err != nil {
			obs.Logger.WithError(err).Error("failed to save user to DB")
			return c.JSON(http.StatusInternalServerError, map[string]string{"msg": err.Error()})
		}

		// Nonsense cache usage example
		err = cache.SetJSON("latestNewUser", newUser, 0)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"msg": err.Error()})
		}

		return c.NoContent(http.StatusCreated)
	})
	// Named route "foobar"
	e.GET("/", func(c echo.Context) error {
		return c.Render(http.StatusOK, "index.html", map[string]interface{}{
			"name": "Dolly!",
		})
	}).Name = "foobar"
	e.GET("/users", func(c echo.Context) error {
		obs.Logger.Info("incoming request to to list al; users")
		users := &[]User{}
		db.Find(&users)
		return c.JSON(http.StatusOK, users)
	})

	// Start the server.
	port := os.Getenv("PORT")
	obs.Logger.WithField("port", port).Info("server running")
	err := e.Start(":" + port)

	if err != nil {
		obs.Logger.Warn(err)
	}

	<-connectionsClosed // Wait for the graceful shutdown to finish.
}

func startFiberRoutes(obs *observance.Obs, db *gorm.DB, cache cache.Cache) {
	e := fiber.New()

	// Set up a routes and handlers.
	e.Post("/users", func(c *fiber.Ctx) {
		obs.Logger.Info("incoming request to create new user")

		newUser := &User{}
		err := c.Locals("newUser", newUser)
		if err != nil {
			c.JSON(map[string]string{"msg": err.(string)})
		}

		if err = db.Save(newUser).Error; err != nil {
			c.JSON(map[string]string{"msg": err.(string)})
		}

		// Nonsense cache usage example
		err = cache.SetJSON("latestNewUser", newUser, 0)
		if err != nil {
			c.JSON(map[string]string{"msg": err.(string)})
		}

		c.SendStatus(http.StatusCreated)
	})
	e.Get("/users", func(c *fiber.Ctx) {
		obs.Logger.Info("incoming request to to list al; users")
		users := &[]User{}
		db.Find(&users)
		c.JSON(users)
		c.SendStatus(http.StatusOK)
	})

	// Start the server.
	port := os.Getenv("PORT")
	// obs.Logger.WithField("port", port).Info("server running")
	e.Listen(":" + port)
}
