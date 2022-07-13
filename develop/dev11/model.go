package main

import (
	"encoding/json"
	"strings"
	"time"
)

type Event struct {
	Id   uint64 `json:"event_id"`
	Name string `json:"name"`
	Desc string `json:"description"`
	Date Date   `json:"date"`
}

type Date struct {
	time.Time
}

func (d *Date) MarshalJSON() ([]byte, error) {
	s := d.Format(dateLayout)
	return json.Marshal(s)
}

func (d *Date) UnmarshalJSON(data []byte) error {
	t, err := time.Parse(dateLayout, strings.Trim(string(data), "\""))
	if err != nil {
		return err
	}

	d.Time = t
	return nil
}

type EventResponse struct {
	Result *Event `json:"result"`
}

type EventsListResponse struct {
	Result []*Event `json:"result"`
}

type ErrorResponse struct {
	Msg string `json:"error"`
}
