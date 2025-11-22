package delivery

import (
	"log/slog"
	"time"
)

// Service is a facade over the Coordinator and per-context Managers.
// It is the main entry point for delivery-related operations from server/bot code.
type Service struct {
	coord *Coordinator
	// queuedStart holds start-delivery requests for supervisors that are not yet running.
	queuedStart map[string]StartRequest
}

// NewService creates a new Service instance using the provided logger.
func NewService(logger *slog.Logger) *Service {
	return &Service{
		coord:       NewCoordinator(logger),
		queuedStart: make(map[string]StartRequest),
	}
}

// StartRequest represents a pending delivery start (room/password) that should
// be applied to a manager/context once the supervisor is available.
type StartRequest struct {
	Room     string
	Password string
}

// AttachManager wires a per-supervisor Manager into the delivery system.
// The Manager instance itself must be created by the caller.
func (s *Service) AttachManager(supervisorName string, mgr *Manager) {
	if mgr == nil {
		return
	}

	// Apply filters and callbacks on this manager
	s.coord.ApplyInitialFilters(supervisorName, mgr)
	s.coord.ConfigureCallbacks(supervisorName, mgr)

	// Apply any queued Express start-delivery request so that the first game
	// for this supervisor will run delivery before normal runs.
	if req, ok := s.consumeQueuedStart(supervisorName); ok {
		mgr.RequestDelivery(req.Room, req.Password)
	}
}

// SetFilters updates the filter configuration for a supervisor and
// immediately applies it to the given Manager if present.
func (s *Service) SetFilters(supervisor string, filters Filters, mgr *Manager) {
	s.coord.SetFilters(supervisor, filters, mgr)
}

// GetFilters returns the current filters configured for the given supervisor.
func (s *Service) GetFilters(supervisor string) (Filters, bool) {
	return s.coord.GetFilters(supervisor)
}

// SetClearServerFilterCallback registers a callback to clear server-side filter state.
func (s *Service) SetClearServerFilterCallback(callback func(supervisor string)) {
	s.coord.SetClearServerFilterCallback(callback)
}

// SetDeliveryResultCallback registers a callback to report delivery results back to the server.
func (s *Service) SetDeliveryResultCallback(callback func(supervisorName, room, result string, itemsDelivered int, duration time.Duration, errorMsg string)) {
	s.coord.SetDeliveryResultCallback(callback)
}

// QueueStartDelivery stores a start-delivery request for a supervisor that is
// currently not running. When its Manager is later attached, the request will
// be applied automatically.
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

// consumeQueuedStart retrieves and removes any queued start-delivery request
// for the given supervisor.
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
