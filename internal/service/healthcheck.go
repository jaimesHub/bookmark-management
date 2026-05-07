package service

// HealthCheck defines the interface for health check service operations.
//
//go:generate mockery --name HealthCheck --filename healthcheck.go
type HealthCheck interface {
	// Check performs a health check and returns the service health status.
	Check() (Response, error)
}

type healthCheckService struct {
	serviceName string
	instanceID  string
}

// NewHealthCheck creates and returns a new HealthCheck service instance.
func NewHealthCheck(cfg *Config) HealthCheck {
	return &healthCheckService{
		serviceName: cfg.ServiceName,
		instanceID:  cfg.InstanceID,
	}
}

// Moving to model/dto later
type Response struct {
	Message     string `json:"message"`
	ServiceName string `json:"service_name"`
	InstanceID  string `json:"instance_id"`
}

func (s *healthCheckService) Check() (Response, error) {
	res := Response{
		Message:     "OK",
		ServiceName: s.serviceName,
		InstanceID:  s.instanceID,
	}
	return res, nil
}
