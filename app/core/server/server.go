package server

import (
	"context"
	"github.com/gofiber/compression"
	"github.com/gofiber/fiber"
	"github.com/gofiber/helmet"
	"github.com/gofiber/recover"
	"github.com/gofiber/requestid"
	"github.com/gofiber/template/mustache"
	"github.com/pkg/errors"
	"os"
	"os/signal"
	"syscall"
	"time"
	"toolkit/app/core/observance"
)

const defaultTimeout = 30 * time.Second

func NewFiber(obs *observance.Obs, timeout ...string) (*fiber.App, error) {
	timeoutDuration := defaultTimeout
	if len(timeout) > 0 {
		parsedTimeout, err := time.ParseDuration(timeout[0])
		if err != nil {
			return nil, errors.Wrap(err, "timeout could not be parsed")
		}
		timeoutDuration = parsedTimeout
	}
	srv := fiber.New()
	/*srv.Use(func(c *fiber.Ctx) {
		c.Fasthttp.
		if c.Error() != nil {
			c.Status(500).SendFile("./errors/500.html")
		} else {
			c.Status(404).SendFile("./errors/404.html")
		}
	})*/
	srv.Settings.DisableStartupMessage = true
	// srv.Settings.Prefork = true
	srv.Settings.IdleTimeout = timeoutDuration
	srv.Settings.ReadTimeout = timeoutDuration
	srv.Settings.WriteTimeout = timeoutDuration
	srv.Settings.ServerHeader = "Verify-Rest"
	cfg := recover.Config{
		Handler: func(c *fiber.Ctx, err error) {
			c.SendString(err.Error())
			c.SendStatus(500)
		},
	}

	srv.Use(recover.New(cfg))
	srv.Use(compression.New())
	srv.Use(requestid.New())
	srv.Use(helmet.New())

	srv.Static("/assets", "./static", fiber.Static{
		Compress:  true,
		ByteRange: true,
	})
	srv.Settings.Templates = mustache.New("./resources/templates", ".mustache")
	// Set up graceful shutdown.
	connsClosed := make(chan struct{})
	sc := make(chan os.Signal)
	go func() {
		s := <-sc
		obs.Logger.WithField("signal", s).Warn("shutting down gracefully")

		_, cancel := context.WithTimeout(context.Background(), 9*time.Second)
		defer cancel()

		err := srv.Shutdown()
		if err != nil {
			obs.Logger.Error(err)
		}
		close(connsClosed)
	}()
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM)

	return srv, nil
}
