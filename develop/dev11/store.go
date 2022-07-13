package main

import (
	"errors"
	"fmt"
	"sort"
	"sync"
	"time"
)

type Store struct {
	sync.RWMutex
	eventCounter uint64
	users        map[uint64]bool
	// Используем мапу для хранения списка эвентов за месяц. Ключем служит
	// строка вида "userId_year-month".
	eventBuckets map[string][]*Event
	// Маппинг id эвентов на ключ вида "userId_year-month".
	mapping map[uint64]string
}

func (s *Store) IsUserExist(userId uint64) bool {
	s.RLock()
	defer s.RUnlock()
	return s.users[userId]
}

func (s *Store) getEventBucketKey(userId uint64, t time.Time) string {
	return fmt.Sprintf("%d_%d-%d", userId, t.Year(), int(t.Month()))
}

func (s *Store) GetEventsForDay(userId uint64, t time.Time) []*Event {
	s.RLock()
	defer s.RUnlock()

	key := s.getEventBucketKey(userId, t)
	_, ok := s.eventBuckets[key]
	if !ok {
		return []*Event{}
	}

	events := []*Event{}
	for _, e := range s.eventBuckets[key] {
		if e.Date.Day() == t.Day() {
			events = append(events, e)
		}
	}

	return events
}

func (s *Store) GetEventsForWeek(userId uint64, t time.Time) []*Event {
	monday := t
	// Убавляем день, пока не получим понедельник.
	for monday.Weekday() != time.Monday {
		monday = monday.AddDate(0, 0, -1)
	}

	sunday := monday.AddDate(0, 0, 6)
	eventsForMonth := []*Event{}
	if monday.Month() != sunday.Month() {
		// Начало и конец недели находятся в разных месяцах.
		eventsForMonth = append(eventsForMonth, s.GetEventsForMonth(userId, monday)...)
		eventsForMonth = append(eventsForMonth, s.GetEventsForMonth(userId, sunday)...)

		sort.Slice(eventsForMonth, func(i, j int) bool {
			return eventsForMonth[i].Id < eventsForMonth[j].Id
		})
	} else {
		eventsForMonth = append(eventsForMonth, s.GetEventsForMonth(userId, t)...)
	}

	events := []*Event{}
	for _, e := range eventsForMonth {
		et := e.Date.Time
		if (et == monday || et.After(monday)) && (et == sunday || et.Before(sunday)) {
			events = append(events, e)
		}
	}

	return events
}

func (s *Store) GetEventsForMonth(userId uint64, t time.Time) []*Event {
	s.RLock()
	defer s.RUnlock()

	key := s.getEventBucketKey(userId, t)
	_, ok := s.eventBuckets[key]
	if !ok {
		return []*Event{}
	}

	events := make([]*Event, len(s.eventBuckets[key]))
	// Копируем и возвращаем копию, а не исходный слайс.
	copy(events, s.eventBuckets[key])

	return events
}

func newEvent(name, desc string, t time.Time) *Event {
	return &Event{
		Name: name,
		Desc: desc,
		Date: Date{t},
	}
}

func (s *Store) CreateEvent(f *Form) *Event {
	s.Lock()
	defer s.Unlock()

	event := newEvent(f.EventName, f.EventDesc, f.EventDate)
	s.eventCounter++
	event.Id = s.eventCounter

	key := s.getEventBucketKey(f.UserId, f.EventDate)
	s.eventBuckets[key] = append(s.eventBuckets[key], event)
	s.mapping[event.Id] = key

	return event
}

func (s *Store) UpdateEvent(eventId uint64, f *Form) (*Event, error) {
	updated := newEvent(f.EventName, f.EventDesc, f.EventDate)
	updated.Id = eventId

	if _, err := s.DeleteEvent(f.UserId, eventId); err != nil {
		return nil, err
	}

	// Берем лок только после удаления старого эвента, чтобы не получить дедлок.
	s.Lock()
	defer s.Unlock()

	key := s.getEventBucketKey(f.UserId, f.EventDate)
	s.eventBuckets[key] = append(s.eventBuckets[key], updated)
	s.mapping[updated.Id] = key

	sort.Slice(s.eventBuckets[key], func(i, j int) bool {
		return s.eventBuckets[key][i].Id < s.eventBuckets[key][j].Id
	})

	return updated, nil
}

var ErrEventNotExist = errors.New("event does not exist")

func (s *Store) DeleteEvent(userId uint64, eventId uint64) (*Event, error) {
	s.Lock()
	defer s.Unlock()

	key, ok := s.mapping[eventId]
	if !ok {
		return nil, ErrEventNotExist
	}

	var event *Event
	length := len(s.eventBuckets[key])
	for i, e := range s.eventBuckets[key] {
		if e.Id == eventId {
			event = e
			s.eventBuckets[key][i] = s.eventBuckets[key][length-1]
			s.eventBuckets[key][length-1] = nil
			s.eventBuckets[key] = s.eventBuckets[key][:length-1]
			break
		}
	}

	delete(s.mapping, eventId)

	sort.Slice(s.eventBuckets[key], func(i, j int) bool {
		return s.eventBuckets[key][i].Id < s.eventBuckets[key][j].Id
	})

	return event, nil
}
