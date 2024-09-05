package main

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/nikoksr/notify"
	"github.com/nikoksr/notify/service/slack"
	"gopkg.in/yaml.v3"
)

var notifiers = map[string]func(map[string]interface{}) error{
	"slack": setupSlack,
}

func setupNotifiers() error {
	configBytes, err := os.ReadFile(*notifyServicesPath)
	if err != nil {
		return fmt.Errorf("Failed to read services file: %w", err)
	}
	config := make(map[string]map[string]interface{})
	err = yaml.Unmarshal(configBytes, &config)

	if err != nil {
		return fmt.Errorf("Failed to read services file: %w", err)
	}

	for k, v := range notifiers {
		if cfg, ok := config[k]; ok {
			err = v(cfg)
			if err != nil {
				slog.Warn("Error setting up service", "service", "slack", "err", err)
			}
		}
	}
	return nil
}

func setupSlack(cfg map[string]interface{}) error {
	token, ok := cfg["token"].(string)
	if !ok {
		return fmt.Errorf("Invalid value for slack.token: %v", cfg["token"])
	}
	svc := slack.New(token)
	receiversIface, ok := cfg["channels"].([]interface{})
	if !ok {
		return fmt.Errorf("Invalid value for slack.channels: %v", cfg["channels"])
	}
	var receivers []string
	for idx, v := range receiversIface {
		val, ok := v.(string)
		if !ok {
			return fmt.Errorf("Invalid value for slack.channels[%d]: %v", idx, v)
		}
		receivers = append(receivers, val)
	}
	svc.AddReceivers(receivers...)
	notify.UseServices(svc)
	return nil
}
