package controller

import (
	"github.com/gofrs/uuid"
	"github.com/labstack/echo/v4"
	"olivermdelgado/url-shortener/pkg/model"
)

func (c *Config) saveURLToDB(original, short string) error {
	su := model.ShortURL{
		OriginalURL: original,
		ShortURL:    short,
	}
	uuid, err := uuid.NewV4()
	if err != nil {
		return err
	}
	su.UUID = uuid.String()

	err = c.DB.Save(&su).Error
	if err != nil {
		return err
	}
	return nil
}

func (c *Config) getFullURLFromDB(shortKey string) (string, error) {
	su := model.ShortURL{}
	err := c.DB.Where("short_url = ?", shortKey).First(&su).Error
	if err != nil {
		return "", err
	}
	return su.OriginalURL, nil
}

// TODO: Write business logic for unique visitors
// TODO: Flesh this out. something like `Select unique(ip_address) from short_url_metrics where short_url="<short_url>"`
func (c *Config) getURLMetrics(ctx echo.Context) error {
	return nil
}
