package delivery

import (
	"log/slog"
	"time"
)

// Service: main delivery entry point
type Service struct {
	coord *Coordinator
	// queued start-delivery requests per supervisor
	queuedStart map[string]StartRequest
}

// Service constructor
func NewService(logger *slog.Logger) *Service {
	return &Service{
		coord:       NewCoordinator(logger),
		queuedStart: make(map[string]StartRequest),
	}
}

// StartRequest: pending request to apply when supervisor is attached
type StartRequest struct {
	Room     string
	Password string
}

// Attach a Manager for a supervisor
func (s *Service) AttachManager(supervisorName string, mgr *Manager) {
	if mgr == nil {
		return
	}

	// Apply filters and callbacks
	s.coord.ApplyInitialFilters(supervisorName, mgr)
	s.coord.ConfigureCallbacks(supervisorName, mgr)

	// Apply queued Express request
	if req, ok := s.consumeQueuedStart(supervisorName); ok {
		mgr.RequestDelivery(req.Room, req.Password)
	}
}

// Set filters for a supervisor
func (s *Service) SetFilters(supervisor string, filters Filters, mgr *Manager) {
	s.coord.SetFilters(supervisor, filters, mgr)
}

// Get filters for a supervisor
func (s *Service) GetFilters(supervisor string) (Filters, bool) {
	return s.coord.GetFilters(supervisor)
}

// Register server filter clear callback
func (s *Service) SetClearServerFilterCallback(callback func(supervisor string)) {
	s.coord.SetClearServerFilterCallback(callback)
}

// Register delivery result callback
func (s *Service) SetDeliveryResultCallback(callback func(supervisorName, room, result string, itemsDelivered int, duration time.Duration, errorMsg string)) {
	s.coord.SetDeliveryResultCallback(callback)
}

// Store start-delivery request to apply when supervisor is attached
func (s *Service) QueueStartDelivery(supervisor, room, password string) {
	if s == nil {
		return
	}
	if s.queuedStart == nil {
		s.queuedStart = make(map[string]StartRequest)
	}
	s.queuedStart[supervisor] = StartRequest{
		Room:     room,
		Password: password,
	}
}

// Return and remove queued start-delivery request
func (s *Service) consumeQueuedStart(supervisor string) (StartRequest, bool) {
	if s == nil || s.queuedStart == nil {
		return StartRequest{}, false
	}
	req, ok := s.queuedStart[supervisor]
	if ok {
		delete(s.queuedStart, supervisor)
	}
	return req, ok
}
