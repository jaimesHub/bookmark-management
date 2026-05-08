// Package service contains core business logic and service implementations.
package service

type Config struct {
	ServiceName string `default:"bookmark_service" envconfig:"SERVICE_NAME"`
	InstanceID  string `default:"" envconfig:"INSTANCE_ID"`
}
