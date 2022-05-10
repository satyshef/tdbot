package leveldb

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/polihoster/tdbot/events/event"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/util"
)

// Store ...
type Store struct {
	db   *leveldb.DB
	path string
}

// Create new store
func New(path string) (*Store, error) {
	db, err := initDB(path)
	if err != nil {
		return nil, err
	}
	return &Store{
			db:   db,
			path: path,
		},
		nil
}

func initDB(path string) (db *leveldb.DB, err error) {
	db, err = leveldb.OpenFile(path, nil)
	if err != nil {
		db, err = leveldb.RecoverFile(path, nil)
		if err != nil {
			return nil, fmt.Errorf("Level DB : %s", err)
		}
	}
	return db, nil
}

// Write event to base
func (s *Store) Write(ev *event.Event) error {
	key := []byte(fmt.Sprintf("%s-%s-%d", ev.Type, ev.Name, ev.Time))
	value := []byte(ev.Data)
	return s.db.Put(key, value, nil)
}

// SearchByTime найти события за указанный интервал времени. Интервал int64(time.Unix())
// @evType - тип события
// @evName - имя события
// @minTime - с какого времени искать
// @maxTime - по какое время
func (s *Store) SearchByTime(evType, evName string, minTime, maxTime int64) (evns []*event.Event, err error) {
	if evType == "" || evName == "" {
		return nil, fmt.Errorf("%s", "Event type or name not specified")
	}
	if maxTime == 0 {
		maxTime = time.Now().Unix()
	}
	if minTime >= maxTime {
		return nil, fmt.Errorf("Не верно указан интервал времени (min - %d ; max - %d)", minTime, maxTime)
	}
	//Обрезаем наносекунды
	startTime := int32(minTime)
	endTime := int32(maxTime)
	from := []byte(fmt.Sprintf("%s-%s-%d", evType, evName, startTime))
	to := []byte(fmt.Sprintf("%s-%s-%d", evType, evName, endTime))
	iter := s.db.NewIterator(&util.Range{Start: from, Limit: to}, nil)
	for iter.Next() {
		if ev := parseRecord(iter.Key(), iter.Value()); ev != nil {
			evns = append(evns, ev)
		}
	}
	iter.Release()
	if err = iter.Error(); err != nil {
		log.Fatalf("SearchByTime error : %s\n", err)
		os.Exit(1)
	}
	return evns, err
}

// Search найти события указанного типа
// @evType - тип события
// @evName - имя события
func (s *Store) Search(evType, evName string) (evns []*event.Event, err error) {
	if evType == "" && evName != "" {
		return nil, fmt.Errorf("%s", "Event name set and type not specified")
	}
	var rang *util.Range
	//Если не указан тип и имя тогда возвращаем все значения(range = nil)
	if evType != "" && evName != "" {
		prefix := []byte(fmt.Sprintf("%s-%s-", evType, evName))
		rang = util.BytesPrefix(prefix)
	} else if evType != "" {
		prefix := []byte(fmt.Sprintf("%s-", evType))
		rang = util.BytesPrefix(prefix)
	}
	iter := s.db.NewIterator(rang, nil)
	for iter.Next() {
		if ev := parseRecord(iter.Key(), iter.Value()); ev != nil {
			evns = append(evns, ev)
		}
	}
	iter.Release()
	err = iter.Error()
	return evns, err
}

func (s *Store) All() (evns []*event.Event, err error) {
	iter := s.db.NewIterator(nil, nil)
	for iter.Next() {
		if ev := parseRecord(iter.Key(), iter.Value()); ev != nil {
			evns = append(evns, ev)
		}
	}
	iter.Release()
	err = iter.Error()
	return evns, err
}

// распарсить key и value в событие
func parseRecord(keyByte, valueByte []byte) *event.Event {

	key := string(keyByte)
	data := string(valueByte)

	r := strings.Split(key, "-")

	if len(r) < 3 {
		return nil
	}

	var evTime int64
	n, err := strconv.ParseInt(r[2], 10, 64)
	if err == nil {
		evTime = n
	}

	return event.New(r[0], r[1], evTime, data)
}

// Close ...
func (s *Store) Close() {
	s.db.Close()
}

//==========================================================================
/*
// Read ...
func (s *Store) Read(name string) error {

	from := []byte(name)
	//iter := s.db.NewIterator(&util.Range{Start: []byte("-1613252945"), Limit: []byte("-1613254297")}, nil)
	iter := s.db.NewIterator(util.BytesPrefix(from), nil)

	for iter.Next() {
		key := iter.Key()
		value := iter.Value()
		fmt.Printf("key: %s | value: %s\n", key, value)
	}

	iter.Release()
	err := iter.Error()
	return err
}

*/
