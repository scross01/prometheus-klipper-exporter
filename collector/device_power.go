package collector

// Power Devices collector for Moonraker device power management
// https://moonraker.readthedocs.io/en/latest/external_api/devices/

import (
	"strings"

	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
)

// MoonrakerPowerDevicesResponse wraps /machine/device_power/devices
type MoonrakerPowerDevicesResponse struct {
	Result struct {
		Devices []MoonrakerPowerDevice `json:"devices"`
	} `json:"result"`
}

// MoonrakerPowerDevice represents a single power device
type MoonrakerPowerDevice struct {
	Device              string `json:"device"`
	Status              string `json:"status"`
	LockedWhilePrinting bool   `json:"locked_while_printing"`
	Type                string `json:"type"`
}

// MoonrakerPowerStatusResponse wraps /machine/device_power/status
type MoonrakerPowerStatusResponse struct {
	Result map[string]string `json:"result"`
}

func (c Collector) collectPowerDevices(ch chan<- prometheus.Metric) {
	log.Infof("Collecting device_power for %s", c.target)

	// Fetch list of power devices
	var devicesResult MoonrakerPowerDevicesResponse
	if err := c.fetchFromMoonraker("/machine/device_power/devices", &devicesResult); err != nil {
		log.Error(err)
		return
	}

	// Emit klipper_power_device_info{device, type} = 1 for each device
	deviceInfoLabels := []string{"device", "type"}
	deviceInfoDesc := prometheus.NewDesc(
		"klipper_power_device_info",
		"Power device information (always 1).",
		deviceInfoLabels, nil,
	)
	for _, d := range devicesResult.Result.Devices {
		ch <- prometheus.MustNewConstMetric(deviceInfoDesc, prometheus.GaugeValue, 1, GetValidLabelName(d.Device), d.Type)
	}

	// Build status URL with device names as query parameters
	deviceNames := make([]string, 0, len(devicesResult.Result.Devices))
	for _, d := range devicesResult.Result.Devices {
		deviceNames = append(deviceNames, d.Device)
	}
	statusURL := "/machine/device_power/status"
	if len(deviceNames) > 0 {
		statusURL += "?" + strings.Join(deviceNames, "&")
	}

	// Fetch device statuses
	var statusResult MoonrakerPowerStatusResponse
	if err := c.fetchFromMoonraker(statusURL, &statusResult); err != nil {
		log.Error(err)
		return
	}

	// Emit klipper_power_device_status{device} (1=on, 0=off/error/init)
	statusLabels := []string{"device"}
	statusDesc := prometheus.NewDesc(
		"klipper_power_device_status",
		"Power device on/off status (1=on, 0=off/error/init).",
		statusLabels, nil,
	)

	// Emit klipper_power_device_state_info{device, state} = 1
	stateInfoLabels := []string{"device", "state"}
	stateInfoDesc := prometheus.NewDesc(
		"klipper_power_device_state_info",
		"Power device state information (always 1).",
		stateInfoLabels, nil,
	)

	for device, state := range statusResult.Result {
		status := 0.0
		if state == "on" {
			status = 1.0
		}
		ch <- prometheus.MustNewConstMetric(statusDesc, prometheus.GaugeValue, status, GetValidLabelName(device))
		ch <- prometheus.MustNewConstMetric(stateInfoDesc, prometheus.GaugeValue, 1, GetValidLabelName(device), state)
	}
}
