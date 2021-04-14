package controller

import (
	"context"
	"errors"
	"github.com/go-redis/redis/v8"
	"github.com/labstack/echo/v4"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/url"
	"strings"
	"time"
)

var (
	KeyNotProvided         = errors.New("no short_url argument provided")
	KeyNotExist            = errors.New("short url does not exist")
	UnexpectedError        = errors.New("an unexpected error occurred")
	UnexpectedParsingError = errors.New("an unexpected parsing error occurred")
	InvalidUrl             = errors.New("invalid URL argument was passed")
	DbSaveError            = errors.New("could not save to DB")
)

type Config struct {
	RDS *redis.Client
	DB  *gorm.DB
	Address string
}

func (c *Config) GetKeyValue(ctx echo.Context) error {
	shortKey := ctx.Param("short_key")
	value, err := c.getExistingKeyValue(shortKey)
	if err == nil {
		if h := ctx.Request().Header.Get("No-Redirect"); h == "true" {
			return ctx.JSON(http.StatusOK, value)
		}
		return ctx.Redirect(http.StatusSeeOther, value)
	}
	log.Error("GetKeyValue:", err)
	switch err {
	case KeyNotProvided:
		return ctx.JSON(http.StatusUnprocessableEntity, err)
	case KeyNotExist:
		return ctx.JSON(http.StatusNotFound, err)
	default:
		return ctx.JSON(http.StatusInternalServerError, UnexpectedError)
	}
}

func (c *Config) CreateNewKeyValue(ctx echo.Context) error {
	// parse out the original url that's the user wants to convert
	b, err := ioutil.ReadAll(ctx.Request().Body)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, UnexpectedError)
	}

	shortURLPath, err := c.createNewKeyValue(string(b))
	if err == nil {
		// TODO: Make this a proper url that's correctly sanitized, escaped, etc
		return ctx.JSON(http.StatusOK, c.Address + "/" + shortURLPath)
	}

	log.Error("CreateNewKeyValue:", err)
	switch err {
	case InvalidUrl:
		return ctx.JSON(http.StatusUnprocessableEntity, err)
	default:
		return ctx.JSON(http.StatusInternalServerError, UnexpectedError)
	}
}

// CreateNewKeyValue creates a new entry in our cache/db, if it doesn't already exist.
func (c *Config) createNewKeyValue(inputURL string) (string, error) {
	// ensure that it's a valid URL
	u, err := url.Parse(inputURL)
	if err != nil {
		return "", InvalidUrl
	}

	// check to see if key already exists and return if it does
	existingShortURL, err := c.getExistingKeyValue(u.String())
	if err == nil {
		return existingShortURL, nil
	}

	// TODO: Make this length vary based on redis cache (+ entries in psql) size
	shortURL := getRandomString(5)

	// TODO: make the timeout length be variable?
	_, err = c.RDS.Set(context.Background(), shortURL, inputURL, 7*24*time.Hour).Result()
	go func() {
		// If this below errors out, it's considered a critical error atm. Log it and move on
		err = c.saveURLToDB(inputURL, shortURL)
		if err != nil {
			log.Error("rip:", DbSaveError, err)
		}
	}()

	return shortURL, nil
}

func (c *Config) getExistingKeyValue(shortKey string) (string, error) {
	// if we STILL don't have a short key, return an error since we can't proceed
	if shortKey == "" {
		return "", KeyNotProvided
	}

	// see if the value exists in the redis cache, else check the postgres DB
	val, err := c.RDS.Get(context.Background(), shortKey).Result()
	if err == redis.Nil {
		// check the psql db to see if the key exists there
		fullURL, err := c.getFullURLFromDB(shortKey)
		if err == nil {
			// it was found in the db! return that
			return fullURL, nil
		}
		// key doesn't exist 😔
		return "", KeyNotExist
	} else if err != nil {
		return "", UnexpectedError
	}

	// parse the url to make it simpler to append the scheme without issues
	u, err := url.Parse(val)
	if err != nil {
		return "", UnexpectedParsingError
	}
	// assume https // TODO: make this conditional, maybe based on which port is being hit?
	u.Scheme = "https"

	// TODO: Update counter on url metrics here

	return u.String(), nil
}

const alphabet = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz012356789"

var alphLen = len(alphabet)

func getRandomString(len int) string {
	var s strings.Builder
	for i := 0; i < len; i++ {
		ch := alphabet[rand.Intn(alphLen)]
		s.WriteByte(ch)
	}
	return s.String()
}
