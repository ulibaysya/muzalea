package compensate

import (
	"sync"
)

type StorageUnit struct {
	Func Func
	Meta []any
}

type Storage struct {
	units []StorageUnit
	mu    sync.Mutex
}

func NewStorage() *Storage {
	return &Storage{}
}

func (s *Storage) Append(u StorageUnit) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.units = append(s.units, u)
}

func (s *Storage) Take() []StorageUnit {
	s.mu.Lock()
	defer s.mu.Unlock()
	units := s.units
	s.units = nil
	return units
}
