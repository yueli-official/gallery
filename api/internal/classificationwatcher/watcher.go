package classificationwatcher

import (
	"context"
	"errors"
	"net"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/lib/pq"
)

const (
	notificationChannel   = "classification_catalog_changed"
	revisionCheckInterval = 30 * time.Second
)

type Watcher struct {
	listener *pq.Listener
	refresh  chan struct{}
	cancel   context.CancelFunc
	done     chan struct{}
	once     sync.Once
}

func Start(config *gdb.ConfigNode, onChange func()) (*Watcher, error) {
	if onChange == nil {
		return nil, errors.New("classification watcher requires an onChange callback")
	}
	connection, err := connectionURL(config)
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithCancel(context.Background())
	watcher := &Watcher{refresh: make(chan struct{}, 1), cancel: cancel, done: make(chan struct{})}
	watcher.listener = pq.NewListener(connection, time.Second, 30*time.Second, func(event pq.ListenerEventType, _ error) {
		if event == pq.ListenerEventReconnected {
			watcher.signal()
		}
	})
	if err := watcher.listener.Listen(notificationChannel); err != nil {
		cancel()
		_ = watcher.listener.Close()
		return nil, err
	}
	go watcher.run(ctx, onChange)
	watcher.signal()
	return watcher, nil
}

func (w *Watcher) run(ctx context.Context, onChange func()) {
	defer close(w.done)
	ticker := time.NewTicker(revisionCheckInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			w.signal()
		case notification, open := <-w.listener.Notify:
			if !open {
				return
			}
			if notification == nil || relevantPayload(notification.Extra) {
				w.signal()
			}
		case <-w.refresh:
			onChange()
		}
	}
}

func (w *Watcher) signal() {
	select {
	case w.refresh <- struct{}{}:
	default:
	}
}

func (w *Watcher) Close() error {
	var closeErr error
	w.once.Do(func() {
		w.cancel()
		closeErr = w.listener.Close()
		<-w.done
	})
	return closeErr
}

func relevantPayload(payload string) bool {
	prefix, revision, found := strings.Cut(strings.TrimSpace(payload), ":")
	if !found || prefix != "gallery" {
		return false
	}
	value, err := strconv.ParseUint(revision, 10, 64)
	return err == nil && value > 0
}

func connectionURL(config *gdb.ConfigNode) (string, error) {
	if config == nil {
		return "", errors.New("classification watcher requires database configuration")
	}
	host := strings.TrimSpace(config.Host)
	name := strings.TrimSpace(config.Name)
	if host == "" || name == "" {
		return "", errors.New("classification watcher requires database host and name")
	}
	port := strings.TrimSpace(config.Port)
	if port == "" {
		port = "5432"
	}
	query := make(url.Values)
	query.Set("sslmode", "disable")
	if strings.TrimSpace(config.Extra) != "" {
		extra, err := url.ParseQuery(config.Extra)
		if err != nil {
			return "", err
		}
		for key, values := range extra {
			query.Del(key)
			for _, value := range values {
				query.Add(key, value)
			}
		}
	}
	if query.Get("application_name") == "" {
		query.Set("application_name", "gallery-classification-listener")
	}
	user := url.User(strings.TrimSpace(config.User))
	if config.Pass != "" {
		user = url.UserPassword(strings.TrimSpace(config.User), config.Pass)
	}
	return (&url.URL{
		Scheme:   "postgresql",
		User:     user,
		Host:     net.JoinHostPort(host, port),
		Path:     "/" + name,
		RawQuery: query.Encode(),
	}).String(), nil
}
