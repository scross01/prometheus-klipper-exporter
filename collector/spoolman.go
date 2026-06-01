package collector

// Spoolman collector for Moonraker's Spoolman proxy integration
// https://moonraker.readthedocs.io/en/latest/external_api/integrations/#spoolman

import (
	"strconv"

	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
)

// SpoolmanSpool represents a single spool returned by the Spoolman API
// via Moonraker's /server/spoolman/spool proxy endpoint.
type SpoolmanSpool struct {
	ID              int     `json:"id"`
	RemainingWeight float64 `json:"remaining_weight"`
	UsedWeight      float64 `json:"used_weight"`
	RemainingLength float64 `json:"remaining_length"`
	UsedLength      float64 `json:"used_length"`
	Filament        struct {
		Name     string `json:"name"`
		Material string `json:"material"`
		ColorHex string `json:"color_hex"`
	} `json:"filament"`
}

func (c Collector) collectSpoolman(ch chan<- prometheus.Metric) {
	var spools []SpoolmanSpool
	if err := c.fetchFromMoonraker("/server/spoolman/spool", &spools); err != nil {
		log.Error(err)
		return
	}

	// klipper_spoolman_spool_info{spool_id, filament_name, material, color} = 1
	spoolInfoLabels := []string{"spool_id", "filament_name", "material", "color"}
	spoolInfoDesc := prometheus.NewDesc(
		"klipper_spoolman_spool_info",
		"Spoolman spool information (always 1).",
		spoolInfoLabels, nil,
	)

	// klipper_spoolman_remaining_weight{spool_id}
	remainingWeightDesc := prometheus.NewDesc(
		"klipper_spoolman_remaining_weight",
		"Remaining filament weight on the spool in grams.",
		[]string{"spool_id"}, nil,
	)

	// klipper_spoolman_used_weight{spool_id}
	usedWeightDesc := prometheus.NewDesc(
		"klipper_spoolman_used_weight",
		"Used filament weight from the spool in grams.",
		[]string{"spool_id"}, nil,
	)

	// klipper_spoolman_remaining_length{spool_id}
	remainingLengthDesc := prometheus.NewDesc(
		"klipper_spoolman_remaining_length",
		"Remaining filament length on the spool in millimetres.",
		[]string{"spool_id"}, nil,
	)

	// klipper_spoolman_used_length{spool_id}
	usedLengthDesc := prometheus.NewDesc(
		"klipper_spoolman_used_length",
		"Used filament length from the spool in millimetres.",
		[]string{"spool_id"}, nil,
	)

	for _, spool := range spools {
		spoolID := strconv.Itoa(spool.ID)

		// Info metric
		ch <- prometheus.MustNewConstMetric(
			spoolInfoDesc, prometheus.GaugeValue, 1.0,
			spoolID,
			GetValidLabelName(spool.Filament.Name),
			GetValidLabelName(spool.Filament.Material),
			GetValidLabelName(spool.Filament.ColorHex),
		)

		// Weight and length metrics
		ch <- prometheus.MustNewConstMetric(remainingWeightDesc, prometheus.GaugeValue, spool.RemainingWeight, spoolID)
		ch <- prometheus.MustNewConstMetric(usedWeightDesc, prometheus.GaugeValue, spool.UsedWeight, spoolID)
		ch <- prometheus.MustNewConstMetric(remainingLengthDesc, prometheus.GaugeValue, spool.RemainingLength, spoolID)
		ch <- prometheus.MustNewConstMetric(usedLengthDesc, prometheus.GaugeValue, spool.UsedLength, spoolID)
	}
}
