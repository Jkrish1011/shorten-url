package routes

import (
	"os"
	"strconv"
	"time"

	"github.com/Jkrish1011/shorten-url/database"
	"github.com/Jkrish1011/shorten-url/helpers"
	"github.com/asaskevich/govalidator"
	"github.com/go-redis/redis/v8"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
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
	r2 := database.CreateClient(1)
	defer r2.Close()

	// here IP of the user is the key
	val, err := r2.Get(database.Ctx, c.IP()).Result()
	if err == redis.Nil {
		// User has not the service in past 30 mins - so set the values and start the counter
		_ = r2.Set(database.Ctx, c.IP(), os.Getenv(os.Getenv("API_QUOTA")), 30*60*time.Second).Err()
	} else {
		// The user has used this service
		val, _ := r2.Get(database.Ctx, c.IP()).Result()
		// valInt will have how many requests the user can make
		valInt, _ := strconv.Atoi(val)
		if valInt <= 0 {
			limit, _ := r2.TTL(database.Ctx, c.IP()).Result()
			return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{"error": "Rate limit exceeded", "rate_limit_reset": limit / time.Nanosecond / time.Minute})
		} else {
			// Quota exhausted
		}
	}

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
	body.URL = helpers.EnforceHTTP(body.URL)

	/*
		Functionality to create a shorten url which the user desires - Custom URL Shortener
		- So check if other's haven't used it yet before creating it
		- If user has not sent a custom url request, the system should create a url for the current request
	*/
	var id string
	if body.CustomShort == "" {
		id = uuid.New().String()[:6]
	} else {
		id = body.CustomShort
	}

	r := database.CreateClient(0)
	defer r.Close()

	val, _ = r.Get(database.Ctx, id).Result()

	if val != "" {
		//Something was found in the database - so the url is already created
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "URL CustomShort is already being used!"})
	}

	if body.Expiry == 0 {
		body.Expiry = 24
	}

	r.Set(database.Ctx, id, body.URL, body.Expiry*3600*time.Second)

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "unable to connect to server"})
	}

	resp := response{
		URL:             body.URL,
		CustomShort:     "",
		Expiry:          body.Expiry,
		XRateRemaining:  10,
		XRateLimitReset: 30,
	}

	r2.Decr(database.Ctx, c.IP())

	val, _ = r2.Get(database.Ctx, c.IP()).Result()
	resp.XRateRemaining, _ = strconv.Atoi(val)
	ttl, _ := r2.TTL(database.Ctx, c.IP()).Result()
	resp.XRateLimitReset = ttl / time.Nanosecond / time.Minute

	resp.CustomShort = os.Getenv("DOMAIN") + "/" + id
	return c.Status(fiber.StatusOK).JSON(resp)

}
