package client

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"

	"github.com/xoticdsign/effectivemobile/internal/utils/config"
)

var (
	ErrNotFound = fmt.Errorf("клиент ничего не нашел")
	ErrInternal = fmt.Errorf("внутренняя ошибка")
)

const source = "client"

type Client struct {
	C C

	log    *slog.Logger
	config config.EffectiveMobileConfig
}

type Handlerer interface {
	GetAge(name string) (int, error)
	GetGender(name string) (string, error)
	GetNationality(name string) (string, error)
}

type C struct {
	Implementations http.Client
	Handlers        Handlerer
}

func New(config config.EffectiveMobileConfig, log *slog.Logger) *Client {
	client := http.Client{
		Timeout: config.Client.Timeout,
	}

	return &Client{
		C: C{
			Implementations: client,
			Handlers: handlers{
				Client: client,

				log:    log,
				config: config,
			},
		},
	}
}

func (c *Client) Shutdown() {
	c.C.Implementations.CloseIdleConnections()
}

type handlers struct {
	UnimplementedHandlers

	Client http.Client

	log    *slog.Logger
	config config.EffectiveMobileConfig
}

type GetAgeResponse struct {
	Count int    `json:"count"`
	Name  string `json:"name"`
	Age   int    `json:"age"`
}

func (h handlers) GetAge(name string) (int, error) {
	const op = "client.GetAge()"

	h.log.Debug(
		"данные получены клиентом",
		slog.String("source", source),
		slog.String("op", op),
	)

	r, err := http.NewRequest(http.MethodGet, "https://api.agify.io/?name="+name, nil)
	if err != nil {
		return 0, err
	}

	resp, err := h.Client.Do(r)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == http.StatusNotFound {
			return 0, ErrNotFound
		}
		return 0, ErrInternal
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, err
	}

	var b GetAgeResponse

	err = json.Unmarshal(body, &b)
	if err != nil {
		return 0, err
	}

	if b.Count == 0 {
		return 0, ErrNotFound
	}

	h.log.Debug(
		"данные обработаны клиентом",
		slog.String("source", source),
		slog.String("op", op),
	)

	return b.Age, nil
}

type GetGenderResponse struct {
	Count       int     `json:"count"`
	Name        string  `json:"name"`
	Gender      string  `json:"gender"`
	Probability float64 `json:"probability"`
}

func (h handlers) GetGender(name string) (string, error) {
	const op = "client.GetGender()"

	h.log.Debug(
		"данные получены клиентом",
		slog.String("source", source),
		slog.String("op", op),
	)

	r, err := http.NewRequest(http.MethodGet, "https://api.genderize.io/?name="+name, nil)
	if err != nil {
		return "", err
	}

	resp, err := h.Client.Do(r)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == http.StatusNotFound {
			return "", ErrNotFound
		}
		return "", ErrInternal
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var b GetGenderResponse

	err = json.Unmarshal(body, &b)
	if err != nil {
		return "", err
	}

	if b.Count == 0 {
		return "", ErrNotFound
	}

	h.log.Debug(
		"данные обработаны клиентом",
		slog.String("source", source),
		slog.String("op", op),
	)

	return b.Gender, nil
}

type GetNationalityResponse struct {
	Count   int       `json:"count"`
	Name    string    `json:"name"`
	Country []Country `json:"country"`
}

type Country struct {
	CountryID   string  `json:"country_id"`
	Probability float64 `json:"probability"`
}

func findMostProbableNationality(countries []Country) string {
	var (
		country string
		p       float64 = -1
	)

	for _, c := range countries {
		if c.Probability > p {
			p = c.Probability
			country = c.CountryID
		}
	}

	return country
}

func (h handlers) GetNationality(name string) (string, error) {
	const op = "client.GetNationality()"

	h.log.Debug(
		"данные получены клиентом",
		slog.String("source", source),
		slog.String("op", op),
	)

	r, err := http.NewRequest(http.MethodGet, "https://api.nationalize.io/?name="+name, nil)
	if err != nil {
		return "", err
	}

	resp, err := h.Client.Do(r)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == http.StatusNotFound {
			return "", ErrNotFound
		}
		return "", ErrInternal
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var b GetNationalityResponse

	err = json.Unmarshal(body, &b)
	if err != nil {
		return "", err
	}

	if b.Count == 0 {
		return "", ErrNotFound
	}

	nationality := findMostProbableNationality(b.Country)

	h.log.Debug(
		"данные обработаны клиентом",
		slog.String("source", source),
		slog.String("op", op),
	)

	return nationality, nil
}

type UnimplementedHandlers struct{}

func (u UnimplementedHandlers) GetAge(name string) (int, error) {
	return 0, nil
}

func (u UnimplementedHandlers) GetGender(name string) (string, error) {
	return "", nil
}

func (u UnimplementedHandlers) GetNationality(name string) (string, error) {
	return "", nil
}
