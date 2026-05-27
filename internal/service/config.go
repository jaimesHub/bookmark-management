// Package service contains core business logic and service implementations.
package service

// Config holds service-level configuration loaded from environment variables.
type Config struct {
	ServiceName string `default:"bookmark_service" envconfig:"SERVICE_NAME"`
	InstanceID  string `default:"" envconfig:"INSTANCE_ID"`
	// Hostname định danh instance khi deploy nhiều VM. Nếu env không set,
	// main.go fallback về os.Hostname() để app vẫn chạy được local.
	Hostname string `envconfig:"APP_HOSTNAME"`
}
