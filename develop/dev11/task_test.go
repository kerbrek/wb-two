package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestApiRoutes(t *testing.T) {
	assert := assert.New(t)
	router := setupRouter()

	d1, _ := time.Parse(dateLayout, "2022-07-01")
	event1 := &Event{Id: 1, Name: "Some event", Desc: "Some description", Date: Date{d1}}

	d2, _ := time.Parse(dateLayout, "2022-06-29")
	event2 := &Event{Id: 2, Name: "Another event", Desc: "", Date: Date{d2}}

	d2upd, _ := time.Parse(dateLayout, "2022-06-28")
	event2upd := &Event{Id: 2, Name: "Updated event", Desc: "Updated description", Date: Date{d2upd}}

	t.Run("POST /create_event #1", func(t *testing.T) {
		body := &bytes.Buffer{}
		body.WriteString("user_id=1&date=2022-07-01&name=Some+event&desc=Some+description")

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/create_event", body)
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		router.ServeHTTP(w, req)
		assert.Equal(http.StatusCreated, w.Code)

		res := new(EventResponse)
		err := json.Unmarshal(w.Body.Bytes(), res)
		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(event1, res.Result)
	})

	t.Run("POST /create_event #2", func(t *testing.T) {
		body := &bytes.Buffer{}
		body.WriteString("user_id=1&date=2022-06-29&name=Another+event")

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/create_event", body)
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		router.ServeHTTP(w, req)
		assert.Equal(http.StatusCreated, w.Code)

		res := new(EventResponse)
		err := json.Unmarshal(w.Body.Bytes(), res)
		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(event2, res.Result)
	})

	t.Run("POST /update_event", func(t *testing.T) {
		body := &bytes.Buffer{}
		body.WriteString("user_id=1&event_id=2&date=2022-06-28&name=Updated+event&desc=Updated+description")

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/update_event", body)
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		router.ServeHTTP(w, req)
		assert.Equal(http.StatusOK, w.Code)

		res := new(EventResponse)
		err := json.Unmarshal(w.Body.Bytes(), res)
		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(event2upd, res.Result)
	})

	t.Run("GET /events_for_day", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/events_for_day?user_id=1&date=2022-07-01", nil)
		router.ServeHTTP(w, req)
		assert.Equal(http.StatusOK, w.Code)

		res := new(EventsListResponse)
		err := json.Unmarshal(w.Body.Bytes(), res)
		if err != nil {
			t.Fatal(err)
		}

		expected := []*Event{event1}
		assert.Equal(expected, res.Result)
	})

	t.Run("GET /events_for_week", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/events_for_week?user_id=1&date=2022-07-01", nil)
		router.ServeHTTP(w, req)
		assert.Equal(http.StatusOK, w.Code)

		res := new(EventsListResponse)
		err := json.Unmarshal(w.Body.Bytes(), res)
		if err != nil {
			t.Fatal(err)
		}

		expected := []*Event{event1, event2upd}
		assert.Equal(expected, res.Result)
	})

	t.Run("GET /events_for_month #1", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/events_for_month?user_id=1&date=2022-07-01", nil)
		router.ServeHTTP(w, req)
		assert.Equal(http.StatusOK, w.Code)

		res := new(EventsListResponse)
		err := json.Unmarshal(w.Body.Bytes(), res)
		if err != nil {
			t.Fatal(err)
		}

		expected := []*Event{event1}
		assert.Equal(expected, res.Result)
	})

	t.Run("GET /events_for_month #2", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/events_for_month?user_id=1&date=2022-06-01", nil)
		router.ServeHTTP(w, req)
		assert.Equal(http.StatusOK, w.Code)

		res := new(EventsListResponse)
		err := json.Unmarshal(w.Body.Bytes(), res)
		if err != nil {
			t.Fatal(err)
		}

		expected := []*Event{event2upd}
		assert.Equal(expected, res.Result)
	})

	t.Run("POST /delete_event", func(t *testing.T) {
		body := &bytes.Buffer{}
		body.WriteString("user_id=1&event_id=2")

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/delete_event", body)
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		router.ServeHTTP(w, req)
		assert.Equal(http.StatusOK, w.Code)

		res := new(EventResponse)
		err := json.Unmarshal(w.Body.Bytes(), res)
		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(event2upd, res.Result)
	})

	t.Run("GET /events_for_week", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/events_for_week?user_id=1&date=2022-07-01", nil)
		router.ServeHTTP(w, req)
		assert.Equal(http.StatusOK, w.Code)

		res := new(EventsListResponse)
		err := json.Unmarshal(w.Body.Bytes(), res)
		if err != nil {
			t.Fatal(err)
		}

		expected := []*Event{event1}
		assert.Equal(expected, res.Result)
	})
}

func TestApiErrors(t *testing.T) {
	assert := assert.New(t)
	router := setupRouter()

	t.Run("POST /create_event", func(t *testing.T) {
		body := &bytes.Buffer{}
		body.WriteString("user_id=1&date=2022-07-01")

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/create_event", body)
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		router.ServeHTTP(w, req)
		assert.Equal(http.StatusBadRequest, w.Code)
	})

	t.Run("POST /create_event", func(t *testing.T) {
		body := &bytes.Buffer{}
		body.WriteString("user_id=100500&date=2022-07-01&name=Some+event")

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/create_event", body)
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		router.ServeHTTP(w, req)
		assert.Equal(http.StatusServiceUnavailable, w.Code)
	})

	t.Run("POST /update_event", func(t *testing.T) {
		body := &bytes.Buffer{}
		body.WriteString("user_id=1&event_id=100500&date=2022-06-28&name=Updated+event&desc=Updated+description")

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/update_event", body)
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		router.ServeHTTP(w, req)
		assert.Equal(http.StatusServiceUnavailable, w.Code)
	})

	t.Run("POST /delete_event", func(t *testing.T) {
		body := &bytes.Buffer{}
		body.WriteString("user_id=1&event_id=100500")

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/delete_event", body)
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		router.ServeHTTP(w, req)
		assert.Equal(http.StatusServiceUnavailable, w.Code)
	})

	t.Run("GET /events_for_month", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/events_for_month?user_id=100500&date=2022-07-01", nil)
		router.ServeHTTP(w, req)
		assert.Equal(http.StatusServiceUnavailable, w.Code)
	})
}
