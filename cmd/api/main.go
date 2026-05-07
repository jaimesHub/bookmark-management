package main

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/jaimesHub/bookmark-management/internal/api"
	"github.com/jaimesHub/bookmark-management/internal/service"
	"github.com/kelseyhightower/envconfig"
)

func main() {
	// create api config
	cfg, err := api.NewConfig()
	if err != nil {
		panic(err)
	}

	// create service config
	var svcCfg service.Config
	err = envconfig.Process("", &svcCfg)
	if err != nil {
		panic(err)
	}

	// Log configuration
	fmt.Printf("APP_PORT: %s\n", cfg.AppPort)
	fmt.Printf("SERVICE_NAME: %s\n", svcCfg.ServiceName)
	fmt.Printf("INSTANCE_ID: %s\n", svcCfg.InstanceID)

	// UUID fallback
	if svcCfg.InstanceID == "" {
		svcCfg.InstanceID = uuid.New().String()
	}

	app := api.NewEngine(cfg, &svcCfg)

	err = app.Start()
	if err != nil {
		panic(err)
	}
}
