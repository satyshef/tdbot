// Package eventman - менеджер событий
package eventman

import (
	"fmt"

	"github.com/satyshef/tdbot/events/event"
	"github.com/satyshef/tdbot/events/eventman/store/leveldb"
)

// Manager ....
type Manager struct {
	Store      *leveldb.Store
	WatchList  []string
	IgnoreList []string
}

var ignoreList = []string{"botReady"}

// Create new storage manager
func New(path string, watchList []string) (*Manager, error) {
	store, err := leveldb.New(path)
	if err != nil {
		return nil, err
	}
	return &Manager{
			Store:      store,
			WatchList:  watchList,
			IgnoreList: ignoreList,
		},
		nil
}

// Добавить событие для наблюдения
func (m *Manager) AddToWatch(evName ...string) {
	m.WatchList = append(m.WatchList, evName...)
}

// Write сохранить событие в хранилище
func (m *Manager) Write(ev *event.Event) error {
	for _, name := range m.IgnoreList {
		if name == ev.Name {
			return nil
		}
	}
	//если в списке нет значений значит пишем все события
	if len(m.WatchList) == 0 {
		if err := m.Store.Write(ev); err != nil {
			return fmt.Errorf("Write Event Error: %s\nEvent: %#v\n\n", err, ev)

		}
		return nil
	}
	for _, name := range m.WatchList {
		if name == ev.Name {
			if err := m.Store.Write(ev); err != nil {
				return fmt.Errorf("Write Event Error: %s\n", err)
			}
			return nil
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

// Search найти все события
func (m *Manager) All() ([]*event.Event, error) {
	return m.Store.All()
}

// Count количество событий
func (m *Manager) Count(evType, evName string) (int, error) {

	evns, err := m.Store.Search(evType, evName)
	if err != nil {
		return 0, err
	}

	return len(evns), nil
}

/*
// GetStats статистика события
// @offset - если указан, то статистика считается по времени начиная с этого смещения назад. Указывается в секундах
func (m *Manager) GetStats1(evType, evName string, offset int32) (int, error) {

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
*/

func (m *Manager) Close() error {
	m.Store.Close()
	return nil
}
