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

type NetworkMaps struct {
	Methods        sync.Map
	URLs           sync.Map
	StartTimes     sync.Map
	NetworkEntries sync.Map
}

func GetNetworkMaps() *NetworkMaps {
	return &NetworkMaps{}
}
