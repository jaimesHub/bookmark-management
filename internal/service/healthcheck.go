// Package service contains core business logic and service implementations.
package service

import "context"

// HealthCheck defines the interface for health check service operations.
//
//go:generate mockery --name HealthCheck --filename healthcheck.go
type HealthCheck interface {
	// Check performs a health check and returns the service health status.
	Check() (Response, error)
}

//go:generate mockery --name Pinger --filename pinger.go
type Pinger interface {
	Ping(ctx context.Context) error
}

// healthCheckService implements the HealthCheck interface.
type healthCheckService struct {
	serviceName string
	instanceID  string
	pinger      Pinger
}

// NewHealthCheck creates and returns a new HealthCheck service instance.
func NewHealthCheck(cfg *Config, pinger Pinger) HealthCheck {
	return &healthCheckService{
		serviceName: cfg.ServiceName,
		instanceID:  cfg.InstanceID,
		pinger:      pinger,
	}
}

// Response represents the health check response data.
type Response struct {
	Message     string `json:"message"`
	ServiceName string `json:"service_name"`
	InstanceID  string `json:"instance_id"`
}

// Check returns the health status of the service with its name and instance ID.
func (s *healthCheckService) Check() (Response, error) {
	if err := s.pinger.Ping(context.Background()); err != nil {
		return Response{}, err
	}

	res := Response{
		Message:     "OK",
		ServiceName: s.serviceName,
		InstanceID:  s.instanceID,
	}
	return res, nil
}
