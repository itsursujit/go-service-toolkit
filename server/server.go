package server

import (
	"github.com/gofiber/fiber"
	"toolkit/observance"
)

// New creates an echo server instance with the given logger, CORS middleware if CORSOrigins was supplied
// and optionally a timeout setting that is applied for read and write.
func NewFiber(obs *observance.Obs, CORSOrigins string, timeout ...string) (*fiber.App, error) {
	srv := fiber.New()
	return srv, nil
}
