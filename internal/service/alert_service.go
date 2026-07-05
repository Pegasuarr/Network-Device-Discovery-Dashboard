package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/smtp"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/user/network-monitoring/internal/model"
	"github.com/user/network-monitoring/internal/repository"
)

type AlertService struct {
	AlertRepo *repository.AlertRepository
}

func NewAlertService(alertRepo *repository.AlertRepository) *AlertService {
	return &AlertService{AlertRepo: alertRepo}
}

func (s *AlertService) CreateAlert(alert *model.Alert) error {
	alert.CreatedAt = time.Now()
	alert.Status = "active"
	err := s.AlertRepo.Create(alert)
	if err == nil {
		// Run notification dispatch in background
		go s.DispatchAlertNotifications(alert)
	}
	return err
}

func (s *AlertService) ListActiveAlerts(orgID uuid.UUID) ([]model.Alert, error) {
	return s.AlertRepo.ListActive(orgID)
}

func (s *AlertService) ResolveAlert(id uuid.UUID, orgID uuid.UUID) error {
	return s.AlertRepo.Resolve(id, orgID)
}

func (s *AlertService) EvaluateRules(device *model.Device, res *model.MonitoringResult) {
	// 1. Fetch Alert Rules for this organization
	rules, err := s.AlertRepo.ListRules(device.OrganizationID)
	if err != nil {
		return
	}

	for _, rule := range rules {
		if !rule.Enabled {
			continue
		}

		// If rule is specific to a device, check matching ID
		if rule.DeviceID != nil && *rule.DeviceID != device.ID {
			continue
		}

		triggered := false
		var metricVal float64
		var currentValStr string

		switch rule.Metric {
		case "latency_ms":
			metricVal = res.LatencyMS
			triggered = s.compareFloat(metricVal, rule.Operator, rule.Value)
			currentValStr = fmt.Sprintf("%.2f ms", metricVal)
		case "packet_loss":
			metricVal = res.PacketLossPct
			triggered = s.compareFloat(metricVal, rule.Operator, rule.Value)
			currentValStr = fmt.Sprintf("%.1f%%", metricVal)
		case "response_time":
			metricVal = res.ResponseTimeMS
			triggered = s.compareFloat(metricVal, rule.Operator, rule.Value)
			currentValStr = fmt.Sprintf("%.2f ms", metricVal)
		case "cpu":
			metricVal = res.CPUUsage
			triggered = s.compareFloat(metricVal, rule.Operator, rule.Value)
			currentValStr = fmt.Sprintf("%.1f%%", metricVal)
		case "ram":
			metricVal = res.RAMUsage
			triggered = s.compareFloat(metricVal, rule.Operator, rule.Value)
			currentValStr = fmt.Sprintf("%.1f%%", metricVal)
		case "disk":
			metricVal = res.DiskUsage
			triggered = s.compareFloat(metricVal, rule.Operator, rule.Value)
			currentValStr = fmt.Sprintf("%.1f%%", metricVal)
		case "ssl_days":
			metricVal = float64(res.SSLDaysRemaining)
			if res.SSLDaysRemaining > 0 { // Only trigger if TLS was successful
				triggered = s.compareFloat(metricVal, rule.Operator, rule.Value)
			}
			currentValStr = fmt.Sprintf("%d days remaining", res.SSLDaysRemaining)
		case "status":
			// For status, check device status
			statusOffline := device.Status == "offline"
			if rule.Operator == "==" && rule.Value == 0 && statusOffline {
				triggered = true
			}
			currentValStr = device.Status
		}

		if triggered {
			// Check if we should suppress the alert due to parent dependency
			if rule.Metric == "status" && device.ParentID != nil {
				var parent model.Device
				err := repository.DB.Where("id = ?", *device.ParentID).First(&parent).Error
				if err == nil && parent.Status == "offline" {
					// Parent is offline, suppress warning/critical alert for child
					// Update child status to unreachable
					repository.DB.Model(&model.Device{}).Where("id = ?", device.ID).Update("status", "unreachable")
					continue
				}
			}

			// Check if alert is already active
			var activeCount int64
			repository.DB.Model(&model.Alert{}).
				Where("device_id = ? AND rule_id = ? AND status = ?", device.ID, rule.ID, "active").
				Count(&activeCount)

			if activeCount == 0 {
				// Fire new alert
				alertMsg := fmt.Sprintf("Rule '%s' triggered. Metric '%s' is %s (Threshold: %s %v)",
					rule.Name, rule.Metric, currentValStr, rule.Operator, rule.Value)

				alert := &model.Alert{
					ID:             uuid.New(),
					OrganizationID: device.OrganizationID,
					DeviceID:       device.ID,
					RuleID:         &rule.ID,
					Type:           rule.Metric,
					Message:        alertMsg,
					Level:          rule.Level,
					Status:         "active",
					CreatedAt:      time.Now(),
				}

				s.CreateAlert(alert)
			}
		} else {
			// If not triggered, auto-resolve active alerts for this rule
			s.AlertRepo.ResolveActiveByDeviceAndType(device.ID, rule.Metric)
		}
	}
}

func (s *AlertService) compareFloat(val float64, op string, target float64) bool {
	switch op {
	case ">":
		return val > target
	case "<":
		return val < target
	case "==":
		return val == target
	case "!=":
		return val != target
	}
	return false
}

// Notifications Dispatcher
func (s *AlertService) DispatchAlertNotifications(alert *model.Alert) {
	// Preload device details
	repository.DB.Preload("Device").First(alert)

	// Fetch Settings for the organization
	var settings []model.Setting
	repository.DB.Where("organization_id = ?", alert.OrganizationID).Find(&settings)

	setMap := make(map[string]string)
	for _, s := range settings {
		setMap[s.Key] = s.Value
	}

	// 1. Email SMTP Notification
	if setMap["smtp_host"] != "" && setMap["smtp_username"] != "" {
		go s.sendEmail(alert, setMap)
	}

	// 2. Slack Webhook Notification
	if setMap["slack_webhook"] != "" {
		go s.sendSlack(alert, setMap["slack_webhook"])
	}

	// 3. Telegram Bot Notification
	if setMap["telegram_token"] != "" && setMap["telegram_chat_id"] != "" {
		go s.sendTelegram(alert, setMap["telegram_token"], setMap["telegram_chat_id"])
	}

	// 4. Discord Webhook Notification
	if setMap["discord_webhook"] != "" {
		go s.sendDiscord(alert, setMap["discord_webhook"])
	}
}

func (s *AlertService) logNotification(alertID uuid.UUID, channel, status, errStr string) {
	log := &model.NotificationLog{
		ID:           uuid.New(),
		AlertID:      alertID,
		Channel:      channel,
		Status:       status,
		ErrorMessage: errStr,
		SentAt:       time.Now(),
	}
	s.AlertRepo.CreateNotificationLog(log)
}

func (s *AlertService) sendEmail(alert *model.Alert, settings map[string]string) {
	host := settings["smtp_host"]
	port := settings["smtp_port"]
	user := settings["smtp_username"]
	pass := settings["smtp_password"]
	to := settings["smtp_to"]

	if to == "" {
		to = "admin@enterprise.local" // Fallback
	}

	auth := smtp.PlainAuth("", user, pass, host)
	msg := []byte(fmt.Sprintf("Subject: [%s] Alert: Device %s is %s\r\n\r\n"+
		"Alert Details:\r\n"+
		"Device: %s (%s)\r\n"+
		"Severity: %s\r\n"+
		"Message: %s\r\n"+
		"Time: %s\r\n",
		strings.ToUpper(alert.Level), alert.Device.Name, alert.Type, alert.Device.Name, alert.Device.IPAddress, alert.Level, alert.Message, alert.CreatedAt.Format(time.RFC1123)))

	err := smtp.SendMail(host+":"+port, auth, user, []string{to}, msg)
	if err != nil {
		s.logNotification(alert.ID, "email", "failed", err.Error())
	} else {
		s.logNotification(alert.ID, "email", "sent", "")
	}
}

func (s *AlertService) sendSlack(alert *model.Alert, webhookURL string) {
	payload := map[string]interface{}{
		"text": fmt.Sprintf("⚠️ *[%s Alert]* Device *%s* has fired an alert: %s (Metric: %s)",
			strings.ToUpper(alert.Level), alert.Device.Name, alert.Message, alert.Type),
	}

	body, _ := json.Marshal(payload)
	resp, err := http.Post(webhookURL, "application/json", bytes.NewBuffer(body))
	if err != nil {
		s.logNotification(alert.ID, "slack", "failed", err.Error())
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		s.logNotification(alert.ID, "slack", "failed", "HTTP status "+resp.Status)
	} else {
		s.logNotification(alert.ID, "slack", "sent", "")
	}
}

func (s *AlertService) sendTelegram(alert *model.Alert, token, chatID string) {
	msgText := fmt.Sprintf("⚠️ *[%s ALERT]*\n*Device*: %s (%s)\n*Message*: %s\n*Time*: %s",
		strings.ToUpper(alert.Level), alert.Device.Name, alert.Device.IPAddress, alert.Message, alert.CreatedAt.Format("2006-01-02 15:04:05"))

	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", token)
	payload := map[string]string{
		"chat_id":    chatID,
		"text":       msgText,
		"parse_mode": "Markdown",
	}

	body, _ := json.Marshal(payload)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(body))
	if err != nil {
		s.logNotification(alert.ID, "telegram", "failed", err.Error())
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		s.logNotification(alert.ID, "telegram", "failed", "HTTP status "+resp.Status)
	} else {
		s.logNotification(alert.ID, "telegram", "sent", "")
	}
}

func (s *AlertService) sendDiscord(alert *model.Alert, webhookURL string) {
	color := 16711680 // red for critical
	if alert.Level == "warning" {
		color = 16776960 // yellow
	} else if alert.Level == "info" {
		color = 65535 // blue
	}

	payload := map[string]interface{}{
		"embeds": []map[string]interface{}{
			{
				"title":       fmt.Sprintf("⚠️ Network Alert: %s", alert.Device.Name),
				"description": alert.Message,
				"color":       color,
				"fields": []map[string]interface{}{
					{"name": "Level", "value": alert.Level, "inline": true},
					{"name": "IP Address", "value": alert.Device.IPAddress, "inline": true},
					{"name": "Type", "value": alert.Type, "inline": true},
				},
				"timestamp": alert.CreatedAt.Format(time.RFC3339),
			},
		},
	}

	body, _ := json.Marshal(payload)
	resp, err := http.Post(webhookURL, "application/json", bytes.NewBuffer(body))
	if err != nil {
		s.logNotification(alert.ID, "discord", "failed", err.Error())
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		s.logNotification(alert.ID, "discord", "failed", "HTTP status "+resp.Status)
	} else {
		s.logNotification(alert.ID, "discord", "sent", "")
	}
}
