package routes

import (
	"time"

	"github.com/Jkrish1011/shorten-url/helpers"
	"github.com/gofiber/fiber"
)

// A format the frontend can expect the APIs request and response can be.
type request struct {
	URL         string        `json:"url"`
	CustomShort string        `json:"short"`
	Expiry      time.Duration `json:"expiry"`
}
type response struct {
	URL             string        `json:"url"`
	CustomShort     string        `json:"short"`
	Expiry          time.Duration `json:"expiry"`
	XRateRemaining  int           `json:"rate_limit"`
	XRateLimitReset time.Duration `json:"rate_limit_reset"`
}

func ShortenURL(c *fiber.Ctx) error {
	body := new(request)
	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "cannot parse JSON"})
	}

	// Implement rate limit
	// Check the IP stored in the DB of all the users. Every 30 mins, 10 request per user. Post 30 mins, the counter will reset.
	if !govalidator.IsURL(body.URL) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid url"})
	}
	// Check if the input sent by the user is an url is not a localhost dev url.
	// Check for domain error
	if !helpers.RemoveDomainError(body.URL) {
		c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{"error": "service unavailable"})
	}
	// Enforce https for all urls given.
	body.URL := helpers.EnforceHTTP(body.URL)
}
