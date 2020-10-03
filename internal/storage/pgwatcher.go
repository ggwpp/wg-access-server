package storage

import (
	"encoding/json"

	"github.com/pkg/errors"
	"github.com/place1/pg-events/pkg/pgevents"
	"github.com/sirupsen/logrus"
)

type PgWatcher struct {
	*pgevents.Listener
}

func NewPgWatcher(connectionString string, table string) (*PgWatcher, error) {
	listener, err := pgevents.OpenListener(connectionString)
	if err != nil {
		return nil, errors.Wrap(err, "failed to open pg listener")
	}

	if err := listener.Attach(table); err != nil {
		return nil, errors.Wrapf(err, "failed to attach listener to table: %s", table)
	}

	return &PgWatcher{
		Listener: listener,
	}, nil
}

func (w *PgWatcher) OnAdd(cb Callback) {
	w.Listener.OnEvent(func(event *pgevents.TableEvent) {
		if event.Action == "UPDATE" || event.Action == "INSERT" {
			w.emit(cb, event)
		}
	})
}

func (w *PgWatcher) OnDelete(cb Callback) {
	w.Listener.OnEvent(func(event *pgevents.TableEvent) {
		if event.Action == "DELETE" {
			w.emit(cb, event)
		}
	})
}

func (w *PgWatcher) emit(cb Callback, event *pgevents.TableEvent) {
	device := &Device{}
	if err := json.Unmarshal([]byte(event.Data), device); err != nil {
		logrus.Error(errors.Wrap(err, "failed to unmarshal postgres event data into device struct"))
	} else {
		cb(device)
	}
}
