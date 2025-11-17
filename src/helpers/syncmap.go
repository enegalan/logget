package helpers

import "sync"

func LoadStringFromSyncMap(m *sync.Map, key string) (string, bool) {
	if val, ok := m.Load(key); ok {
		if str, ok := val.(string); ok {
			return str, true
		}
	}
	return "", false
}

func LoadFloat64FromSyncMap(m *sync.Map, key string) (float64, bool) {
	if val, ok := m.Load(key); ok {
		if f, ok := val.(float64); ok {
			return f, true
		}
	}
	return 0, false
}

type SyncMaps struct {
	RequestMethods    sync.Map
	RequestHeadersMap sync.Map
	RequestURLs       sync.Map
	RequestStartTimes sync.Map
	NetworkEntriesMap sync.Map
}

func NewSyncMaps() *SyncMaps { return &SyncMaps{} }
