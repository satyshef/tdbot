// Package eventman - менеджер событий
package eventman

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/polihoster/tdbot/events/event"
	"github.com/polihoster/tdbot/events/eventman/store/leveldb"
)

// Manager ....
type Manager struct {
	Store     *leveldb.Store
	WatchList []string
}

// New ...
func New(path string, watchList []string) *Manager {
	return &Manager{
		Store:     leveldb.New(path),
		WatchList: watchList,
	}

}

// AddToWatch ...
func (m *Manager) AddToWatch(evName ...string) {
	m.WatchList = append(m.WatchList, evName...)
}

// Write сохранить событие в хранилище
func (m *Manager) Write(ev *event.Event) error {
	for _, name := range m.WatchList {
		if name == ev.Name || name == "*" {
			if err := m.Store.Write(ev); err != nil {
				log.Fatalf("Write Event : %s\n", err)
				os.Exit(1)
			} else {
				return nil
			}
		}
	}
	return fmt.Errorf("Event not observed")
}

// SearchByTime найти события за интервал времени
func (m *Manager) SearchByTime(evType, evName string, minTime, maxTime int64) ([]*event.Event, error) {
	return m.Store.SearchByTime(evType, evName, minTime, maxTime)

}

// Search найти события определенного типа
func (m *Manager) Search(evType, evName string) ([]*event.Event, error) {
	return m.Store.Search(evType, evName)
}

// Count количество событий
func (m *Manager) Count(evType, evName string) (int, error) {

	evns, err := m.Store.Search(evType, evName)
	if err != nil {
		return 0, err
	}

	return len(evns), nil
}

// GetStats статистика события
// @offset - если указан, то статистика считается по времени начиная с этого смещения назад. Указывается в секундах
func (m *Manager) GetStats(evType, evName string, offset int32) (int, error) {

	var evns []*event.Event
	var err error

	if offset != 0 {
		//Проверить!!!
		minTime := time.Now().Unix() - int64(offset)
		evns, err = m.SearchByTime(evType, evName, minTime, 0)
	} else {
		evns, err = m.Search(evType, evName)
	}

	if err != nil {
		return 0, err
	}

	return len(evns), nil

}
