package tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"

	effectivemobileapp "github.com/xoticdsign/effectivemobile/internal/app/effectivemobile"
	effectivemobileservice "github.com/xoticdsign/effectivemobile/internal/service/effectivemobile"
	storage "github.com/xoticdsign/effectivemobile/internal/storage/postgresql"
	"github.com/xoticdsign/effectivemobile/tests/suite"
)

func TestDeleteByID_Functional(t *testing.T) {
	s := suite.New(t)

	h := effectivemobileapp.Handlers{
		Service: effectivemobileservice.UnimplementedHandlers{},
		Log:     s.Log.Log,
		Config:  s.Config.EffectiveMobile,
	}

	f := fiber.New()
	defer f.Shutdown()

	f.Delete(fmt.Sprintf("/%s/%s", effectivemobileapp.DeleteByIDHanlder, effectivemobileapp.DeleteByIDParameters), h.DeleteByID)

	cases := []struct {
		name         string
		inMethod     string
		inBody       effectivemobileapp.DeleteByIDRequest
		inTarget     string
		expectedErr  error
		expectedCode int
		expectedBody effectivemobileapp.DeleteByIDResponse
	}{
		{
			name:         "happy case",
			inMethod:     http.MethodDelete,
			inBody:       effectivemobileapp.DeleteByIDRequest{},
			inTarget:     fmt.Sprintf("/%s/1", effectivemobileapp.DeleteByIDHanlder),
			expectedErr:  nil,
			expectedCode: fiber.StatusOK,
			expectedBody: effectivemobileapp.DeleteByIDResponse{
				Status:  fiber.StatusOK,
				Message: effectivemobileapp.DeleteByIDSuccess,
			},
		},
		{
			name:         "not found case",
			inMethod:     http.MethodDelete,
			inBody:       effectivemobileapp.DeleteByIDRequest{},
			inTarget:     fmt.Sprintf("/%s", effectivemobileapp.DeleteByIDHanlder),
			expectedErr:  nil,
			expectedCode: fiber.StatusNotFound,
			expectedBody: effectivemobileapp.DeleteByIDResponse{},
		},
		{
			name:         "wrong method case",
			inMethod:     http.MethodGet,
			inBody:       effectivemobileapp.DeleteByIDRequest{},
			inTarget:     fmt.Sprintf("/%s/1", effectivemobileapp.DeleteByIDHanlder),
			expectedErr:  nil,
			expectedCode: fiber.StatusMethodNotAllowed,
			expectedBody: effectivemobileapp.DeleteByIDResponse{},
		},
		{
			name:         "storage not found case",
			inMethod:     http.MethodDelete,
			inBody:       effectivemobileapp.DeleteByIDRequest{},
			inTarget:     fmt.Sprintf("/%s/s404", effectivemobileapp.DeleteByIDHanlder),
			expectedErr:  nil,
			expectedCode: fiber.StatusNotFound,
			expectedBody: effectivemobileapp.DeleteByIDResponse{},
		},
		{
			name:         "storage internal case",
			inMethod:     http.MethodDelete,
			inBody:       effectivemobileapp.DeleteByIDRequest{},
			inTarget:     fmt.Sprintf("/%s/s500", effectivemobileapp.DeleteByIDHanlder),
			expectedErr:  nil,
			expectedCode: fiber.StatusInternalServerError,
			expectedBody: effectivemobileapp.DeleteByIDResponse{},
		},
	}

	for _, c := range cases {
		s.T.Run(c.name, func(t *testing.T) {
			b, _ := json.Marshal(c.inBody)

			r := httptest.NewRequest(c.inMethod, c.inTarget, bytes.NewBuffer(b))
			r.Header.Set("Content-Type", "application/json")

			resp, err := f.Test(r, int(s.Config.EffectiveMobile.Client.Timeout))
			if err != nil {
				assert.Equal(t, c.expectedErr, err)
			}
			defer resp.Body.Close()

			assert.Equal(t, c.expectedCode, resp.StatusCode)

			if resp.StatusCode == fiber.StatusOK {
				var body effectivemobileapp.DeleteByIDResponse

				rb, err := io.ReadAll(resp.Body)
				assert.NoError(t, err)

				err = json.Unmarshal(rb, &body)
				assert.NoError(t, err)

				assert.Equal(t, c.expectedBody, body)
			}
		})
	}
}

func TestUpdateByID_Functional(t *testing.T) {
	s := suite.New(t)

	h := effectivemobileapp.Handlers{
		Service: effectivemobileservice.UnimplementedHandlers{},
		Log:     s.Log.Log,
		Config:  s.Config.EffectiveMobile,
	}

	f := fiber.New()
	defer f.Shutdown()

	f.Put(fmt.Sprintf("/%s/%s", effectivemobileapp.UpdateByIDHandler, effectivemobileapp.UpdateByIDParameters), h.UpdateByID)

	cases := []struct {
		name         string
		inMethod     string
		inBody       effectivemobileapp.UpdateByIDRequest
		inTarget     string
		expectedErr  error
		expectedCode int
		expectedBody effectivemobileapp.UpdateByIDResponse
	}{
		{
			name:     "happy case",
			inMethod: http.MethodPut,
			inBody: effectivemobileapp.UpdateByIDRequest{
				Name: "test",
			},
			inTarget:     fmt.Sprintf("/%s/1", effectivemobileapp.UpdateByIDHandler),
			expectedErr:  nil,
			expectedCode: fiber.StatusOK,
			expectedBody: effectivemobileapp.UpdateByIDResponse{
				Status:  fiber.StatusOK,
				Message: effectivemobileapp.UpdateByIDSuccess,
			},
		},
		{
			name:         "not found case",
			inMethod:     http.MethodPut,
			inBody:       effectivemobileapp.UpdateByIDRequest{},
			inTarget:     fmt.Sprintf("/%s", effectivemobileapp.UpdateByIDHandler),
			expectedErr:  nil,
			expectedCode: fiber.StatusNotFound,
			expectedBody: effectivemobileapp.UpdateByIDResponse{},
		},
		{
			name:         "worng method case",
			inMethod:     http.MethodGet,
			inBody:       effectivemobileapp.UpdateByIDRequest{},
			inTarget:     fmt.Sprintf("/%s/1", effectivemobileapp.UpdateByIDHandler),
			expectedErr:  nil,
			expectedCode: fiber.StatusMethodNotAllowed,
			expectedBody: effectivemobileapp.UpdateByIDResponse{},
		},
		{
			name:         "storage not found case",
			inMethod:     http.MethodPut,
			inBody:       effectivemobileapp.UpdateByIDRequest{},
			inTarget:     fmt.Sprintf("/%s/s404", effectivemobileapp.UpdateByIDHandler),
			expectedErr:  nil,
			expectedCode: fiber.StatusNotFound,
			expectedBody: effectivemobileapp.UpdateByIDResponse{},
		},
		{
			name:         "storage bad request case",
			inMethod:     http.MethodPut,
			inBody:       effectivemobileapp.UpdateByIDRequest{},
			inTarget:     fmt.Sprintf("/%s/s400", effectivemobileapp.UpdateByIDHandler),
			expectedErr:  nil,
			expectedCode: fiber.StatusBadRequest,
			expectedBody: effectivemobileapp.UpdateByIDResponse{},
		},
		{
			name:         "storage internal case",
			inMethod:     http.MethodPut,
			inBody:       effectivemobileapp.UpdateByIDRequest{},
			inTarget:     fmt.Sprintf("/%s/s500", effectivemobileapp.UpdateByIDHandler),
			expectedErr:  nil,
			expectedCode: fiber.StatusInternalServerError,
			expectedBody: effectivemobileapp.UpdateByIDResponse{},
		},
	}

	for _, c := range cases {
		s.T.Run(c.name, func(t *testing.T) {
			b, _ := json.Marshal(c.inBody)

			r := httptest.NewRequest(c.inMethod, c.inTarget, bytes.NewBuffer(b))
			r.Header.Set("Content-Type", "application/json")

			resp, err := f.Test(r, int(s.Config.EffectiveMobile.Client.Timeout))
			if err != nil {
				assert.Equal(t, c.expectedErr, err)
			}
			defer resp.Body.Close()

			assert.Equal(t, c.expectedCode, resp.StatusCode)

			if resp.StatusCode == fiber.StatusOK {
				var body effectivemobileapp.UpdateByIDResponse

				rb, err := io.ReadAll(resp.Body)
				assert.NoError(t, err)

				err = json.Unmarshal(rb, &body)
				assert.NoError(t, err)

				assert.Equal(t, c.expectedBody, body)
			}
		})
	}
}

func TestCreate_Functional(t *testing.T) {
	s := suite.New(t)

	h := effectivemobileapp.Handlers{
		Service: effectivemobileservice.UnimplementedHandlers{},
		Log:     s.Log.Log,
		Config:  s.Config.EffectiveMobile,
	}

	f := fiber.New()
	defer f.Shutdown()

	f.Post(fmt.Sprintf("/%s", effectivemobileapp.CreateHandler), h.Create)

	cases := []struct {
		name         string
		inMethod     string
		inBody       effectivemobileapp.CreateRequest
		inTarget     string
		expectedErr  error
		expectedCode int
		expectedBody effectivemobileapp.CreateResponse
	}{
		{
			name:     "happy case",
			inMethod: http.MethodPost,
			inBody: effectivemobileapp.CreateRequest{
				Name:    "test",
				Surname: "test",
			},
			inTarget:     fmt.Sprintf("/%s", effectivemobileapp.CreateHandler),
			expectedErr:  nil,
			expectedCode: fiber.StatusOK,
			expectedBody: effectivemobileapp.CreateResponse{
				Status:  fiber.StatusOK,
				Message: effectivemobileapp.CreateSuccess,
			},
		},
		{
			name:         "bad request case",
			inMethod:     http.MethodPost,
			inBody:       effectivemobileapp.CreateRequest{},
			inTarget:     fmt.Sprintf("/%s", effectivemobileapp.CreateHandler),
			expectedErr:  nil,
			expectedCode: fiber.StatusBadRequest,
			expectedBody: effectivemobileapp.CreateResponse{},
		},
		{
			name:         "worng method case",
			inMethod:     http.MethodGet,
			inBody:       effectivemobileapp.CreateRequest{},
			inTarget:     fmt.Sprintf("/%s", effectivemobileapp.CreateHandler),
			expectedErr:  nil,
			expectedCode: fiber.StatusMethodNotAllowed,
			expectedBody: effectivemobileapp.CreateResponse{},
		},
		{
			name:     "storage not found case",
			inMethod: http.MethodPost,
			inBody: effectivemobileapp.CreateRequest{
				Name:    "s404",
				Surname: "test",
			},
			inTarget:     fmt.Sprintf("/%s", effectivemobileapp.CreateHandler),
			expectedErr:  nil,
			expectedCode: fiber.StatusNotFound,
			expectedBody: effectivemobileapp.CreateResponse{},
		},
		{
			name:     "client not found case",
			inMethod: http.MethodPost,
			inBody: effectivemobileapp.CreateRequest{
				Name:    "c404",
				Surname: "test",
			},
			inTarget:     fmt.Sprintf("/%s", effectivemobileapp.CreateHandler),
			expectedErr:  nil,
			expectedCode: fiber.StatusNotFound,
			expectedBody: effectivemobileapp.CreateResponse{},
		},
		{
			name:     "internal case",
			inMethod: http.MethodPost,
			inBody: effectivemobileapp.CreateRequest{
				Name:    "500",
				Surname: "test",
			},
			inTarget:     fmt.Sprintf("/%s", effectivemobileapp.CreateHandler),
			expectedErr:  nil,
			expectedCode: fiber.StatusInternalServerError,
			expectedBody: effectivemobileapp.CreateResponse{},
		},
	}

	for _, c := range cases {
		s.T.Run(c.name, func(t *testing.T) {
			b, _ := json.Marshal(c.inBody)

			r := httptest.NewRequest(c.inMethod, c.inTarget, bytes.NewBuffer(b))
			r.Header.Set("Content-Type", "application/json")

			resp, err := f.Test(r, int(s.Config.EffectiveMobile.Client.Timeout))
			if err != nil {
				assert.Equal(t, c.expectedErr, err)
			}
			defer resp.Body.Close()

			assert.Equal(t, c.expectedCode, resp.StatusCode)

			if resp.StatusCode == fiber.StatusOK {
				var body effectivemobileapp.CreateResponse

				rb, err := io.ReadAll(resp.Body)
				assert.NoError(t, err)

				err = json.Unmarshal(rb, &body)
				assert.NoError(t, err)

				assert.Equal(t, c.expectedBody, body)
			}
		})
	}
}

func TestSelect_Functional(t *testing.T) {
	s := suite.New(t)

	h := effectivemobileapp.Handlers{
		Service: effectivemobileservice.UnimplementedHandlers{},
		Log:     s.Log.Log,
		Config:  s.Config.EffectiveMobile,
	}

	f := fiber.New()
	defer f.Shutdown()

	f.Get(fmt.Sprintf("/%s/%s", effectivemobileapp.SelectHandler, effectivemobileapp.SelectParameters), h.Select)

	cases := []struct {
		name         string
		inMethod     string
		inBody       *effectivemobileapp.SelectRequest
		inTarget     string
		expectedErr  error
		expectedCode int
		expectedBody effectivemobileapp.SelectResponse
	}{
		{
			name:         "happy case",
			inMethod:     http.MethodGet,
			inBody:       &effectivemobileapp.SelectRequest{},
			inTarget:     fmt.Sprintf("/%s/1", effectivemobileapp.SelectHandler),
			expectedErr:  nil,
			expectedCode: fiber.StatusOK,
			expectedBody: effectivemobileapp.SelectResponse{
				Status:  fiber.StatusOK,
				Message: effectivemobileapp.SelectSuccess,
				Result:  []storage.Row{},
			},
		},
		{
			name:         "bad request case",
			inMethod:     http.MethodGet,
			inBody:       nil,
			inTarget:     fmt.Sprintf("/%s", effectivemobileapp.SelectHandler),
			expectedErr:  nil,
			expectedCode: fiber.StatusBadRequest,
			expectedBody: effectivemobileapp.SelectResponse{},
		},
		{
			name:         "worng method case",
			inMethod:     http.MethodPost,
			inBody:       &effectivemobileapp.SelectRequest{},
			inTarget:     fmt.Sprintf("/%s/1", effectivemobileapp.SelectHandler),
			expectedErr:  nil,
			expectedCode: fiber.StatusMethodNotAllowed,
			expectedBody: effectivemobileapp.SelectResponse{},
		},
		{
			name:         "storage not found case",
			inMethod:     http.MethodGet,
			inBody:       &effectivemobileapp.SelectRequest{},
			inTarget:     fmt.Sprintf("/%s/s404", effectivemobileapp.SelectHandler),
			expectedErr:  nil,
			expectedCode: fiber.StatusNotFound,
			expectedBody: effectivemobileapp.SelectResponse{},
		},
		{
			name:         "storage internal case",
			inMethod:     http.MethodGet,
			inBody:       &effectivemobileapp.SelectRequest{},
			inTarget:     fmt.Sprintf("/%s/s500", effectivemobileapp.SelectHandler),
			expectedErr:  nil,
			expectedCode: fiber.StatusInternalServerError,
			expectedBody: effectivemobileapp.SelectResponse{},
		},
	}

	for _, c := range cases {
		s.T.Run(c.name, func(t *testing.T) {
			var b []byte

			if c.inBody != nil {
				b, _ = json.Marshal(c.inBody)
			}

			r := httptest.NewRequest(c.inMethod, c.inTarget, bytes.NewBuffer(b))
			r.Header.Set("Content-Type", "application/json")

			resp, err := f.Test(r, int(s.Config.EffectiveMobile.Client.Timeout))
			if err != nil {
				assert.Equal(t, c.expectedErr, err)
			}
			defer resp.Body.Close()

			assert.Equal(t, c.expectedCode, resp.StatusCode)

			if resp.StatusCode == fiber.StatusOK {
				var body effectivemobileapp.SelectResponse

				rb, err := io.ReadAll(resp.Body)
				assert.NoError(t, err)

				err = json.Unmarshal(rb, &body)
				assert.NoError(t, err)

				assert.Equal(t, c.expectedBody, body)
			}
		})
	}
}
