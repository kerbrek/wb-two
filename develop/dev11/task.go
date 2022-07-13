package main

/*
=== HTTP server ===

Реализовать HTTP сервер для работы с календарем. В рамках задания необходимо работать строго со
стандартной HTTP библиотекой.
В рамках задания необходимо:
	1. Реализовать вспомогательные функции для сериализации объектов доменной области в JSON.
	2. Реализовать вспомогательные функции для парсинга и валидации параметров методов /create_event
	и /update_event.
	3. Реализовать HTTP обработчики для каждого из методов API, используя вспомогательные функции и
	объекты доменной области.
	4. Реализовать middleware для логирования запросов
Методы API:
	POST /create_event
	POST /update_event
	POST /delete_event
	GET /events_for_day
	GET /events_for_week
	GET /events_for_month
Параметры передаются в виде www-url-form-encoded (т.е. обычные user_id=3&date=2019-09-09).
В GET методах параметры передаются через queryString, в POST через тело запроса.
В результате каждого запроса должен возвращаться JSON документ содержащий либо {"result": "..."}
в случае успешного выполнения метода, либо {"error": "..."} в случае ошибки бизнес-логики.

В рамках задачи необходимо:
	1. Реализовать все методы.
	2. Бизнес логика НЕ должна зависеть от кода HTTP сервера.
	3. В случае ошибки бизнес-логики сервер должен возвращать HTTP 503. В случае ошибки входных
		данных (невалидный int например) сервер должен возвращать HTTP 400. В случае остальных ошибок
		сервер должен возвращать HTTP 500. Web-сервер должен запускаться на порту указанном в конфиге
		и выводить в лог каждый обработанный запрос.
	4. Код должен проходить проверки go vet и golint.
*/

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/felixge/httpsnoop"
)

var storage = Store{
	eventCounter: 0,
	users:        map[uint64]bool{1: true},
	eventBuckets: make(map[string][]*Event),
	mapping:      make(map[uint64]string),
}

func main() {
	config, err := parseConfig("config.json")
	if err != nil {
		log.Fatal(err)
	}

	router := setupRouter()
	addr := fmt.Sprintf(":%d", config.Port)
	log.Fatal(http.ListenAndServe(addr, advancedLoggingMiddleware(router)))
}

type Config struct {
	Port int `json:"port"`
}

func parseConfig(filename string) (*Config, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	config := new(Config)
	if err = json.Unmarshal(data, config); err != nil {
		return nil, err
	}

	return config, nil
}

func setupRouter() *http.ServeMux {
	router := http.NewServeMux()
	router.HandleFunc("/", rootHandler)
	router.HandleFunc("/create_event", createEventHandler)
	router.HandleFunc("/update_event", updateEventHandler)
	router.HandleFunc("/delete_event", deleteEventHandler)
	router.HandleFunc("/events_for_day", getEventsForDayHandler)
	router.HandleFunc("/events_for_week", getEventsForWeekHandler)
	router.HandleFunc("/events_for_month", getEventsForMonthHandler)
	return router
}

//lint:ignore U1000 Ignore unused code
func simpleLoggingMiddleware(handler http.Handler) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			handler.ServeHTTP(w, r)
			log.Printf(
				"%s - %s %s",
				r.RemoteAddr,
				r.Method,
				r.URL,
			)
		},
	)
}

func advancedLoggingMiddleware(handler http.Handler) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			// https://github.com/felixge/httpsnoop#why-this-package-exists
			m := httpsnoop.CaptureMetrics(handler, w, r)
			log.Printf(
				"%s - %s %s (code=%d dt=%s written=%d)",
				r.RemoteAddr,
				r.Method,
				r.URL,
				m.Code,
				m.Duration,
				m.Written,
			)
		},
	)
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "", http.StatusInternalServerError)
}

func sendResponse(w http.ResponseWriter, resp any, statusCode int) {
	b, err := json.Marshal(resp)
	if err != nil {
		log.Print(err)
		http.Error(w, "", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	w.Write(b)
}

type Form struct {
	UserId    uint64
	EventDate time.Time
	EventName string
	EventDesc string
}

var ErrBlankEventName = errors.New("event name can not be blank")

func (f Form) Validate() error {
	if f.EventName == "" {
		return ErrBlankEventName
	}

	return nil
}

const dateLayout = "2006-01-02"

func parseFormData(r *http.Request) (*Form, error) {
	userId, err := strconv.ParseUint(r.PostFormValue("user_id"), 10, 64)
	if err != nil {
		return nil, err
	}

	eventDate, err := time.Parse(dateLayout, r.PostFormValue("date"))
	if err != nil {
		return nil, err
	}

	eventName := strings.TrimSpace(r.PostFormValue("name"))
	eventDesc := strings.TrimSpace(r.PostFormValue("desc"))

	form := &Form{
		UserId:    userId,
		EventDate: eventDate,
		EventName: eventName,
		EventDesc: eventDesc,
	}

	return form, nil
}

var ErrUserNotExist = errors.New("user does not exist")

func createEventHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "", http.StatusInternalServerError)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "", http.StatusBadRequest)
		return
	}

	form, err := parseFormData(r)
	if err != nil {
		http.Error(w, "", http.StatusBadRequest)
		fmt.Println("parseFormData:", form)
		return
	}

	if err := form.Validate(); err != nil {
		http.Error(w, "", http.StatusBadRequest)
		return
	}

	if !storage.IsUserExist(form.UserId) {
		resp := ErrorResponse{Msg: ErrUserNotExist.Error()}
		sendResponse(w, resp, http.StatusServiceUnavailable)
		return
	}

	event := storage.CreateEvent(form)
	resp := EventResponse{Result: event}
	sendResponse(w, resp, http.StatusCreated)
}

func updateEventHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "", http.StatusInternalServerError)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "", http.StatusBadRequest)
		return
	}

	form, err := parseFormData(r)
	if err != nil {
		http.Error(w, "", http.StatusBadRequest)
		return
	}

	if vErr := form.Validate(); vErr != nil {
		http.Error(w, "", http.StatusBadRequest)
		return
	}

	eventId, err := strconv.ParseUint(r.PostFormValue("event_id"), 10, 64)
	if err != nil {
		http.Error(w, "", http.StatusBadRequest)
		return
	}

	if !storage.IsUserExist(form.UserId) {
		resp := ErrorResponse{Msg: ErrUserNotExist.Error()}
		sendResponse(w, resp, http.StatusServiceUnavailable)
		return
	}

	event, err := storage.UpdateEvent(eventId, form)
	if err != nil {
		resp := ErrorResponse{Msg: err.Error()}
		sendResponse(w, resp, http.StatusServiceUnavailable)
		return
	}

	resp := EventResponse{Result: event}
	sendResponse(w, resp, http.StatusOK)
}

func deleteEventHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "", http.StatusInternalServerError)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "", http.StatusBadRequest)
		return
	}

	userId, err := strconv.ParseUint(r.PostFormValue("user_id"), 10, 64)
	if err != nil {
		http.Error(w, "", http.StatusBadRequest)
		return
	}

	eventId, err := strconv.ParseUint(r.PostFormValue("event_id"), 10, 64)
	if err != nil {
		http.Error(w, "", http.StatusBadRequest)
		return
	}

	if !storage.IsUserExist(userId) {
		resp := ErrorResponse{Msg: ErrUserNotExist.Error()}
		sendResponse(w, resp, http.StatusServiceUnavailable)
		return
	}

	event, err := storage.DeleteEvent(userId, eventId)
	if err != nil {
		resp := ErrorResponse{Msg: err.Error()}
		sendResponse(w, resp, http.StatusServiceUnavailable)
		return
	}

	resp := EventResponse{Result: event}
	sendResponse(w, resp, http.StatusOK)
}

func getEventsForDayHandler(w http.ResponseWriter, r *http.Request) {
	getEvents(w, r, storage.GetEventsForDay)
}

func getEventsForWeekHandler(w http.ResponseWriter, r *http.Request) {
	getEvents(w, r, storage.GetEventsForWeek)
}

func getEventsForMonthHandler(w http.ResponseWriter, r *http.Request) {
	getEvents(w, r, storage.GetEventsForMonth)
}

func getEvents(w http.ResponseWriter, r *http.Request, fn func(uint64, time.Time) []*Event) {
	if r.Method != http.MethodGet {
		http.Error(w, "", http.StatusInternalServerError)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "", http.StatusInternalServerError)
		return
	}

	userId, err := strconv.ParseUint(r.FormValue("user_id"), 10, 64)
	if err != nil {
		http.Error(w, "", http.StatusBadRequest)
		return
	}

	date, err := time.Parse(dateLayout, r.FormValue("date"))
	if err != nil {
		http.Error(w, "", http.StatusBadRequest)
		return
	}

	if !storage.IsUserExist(userId) {
		resp := ErrorResponse{Msg: ErrUserNotExist.Error()}
		sendResponse(w, resp, http.StatusServiceUnavailable)
		return
	}

	events := fn(userId, date)
	resp := EventsListResponse{Result: events}
	sendResponse(w, resp, http.StatusOK)
}
