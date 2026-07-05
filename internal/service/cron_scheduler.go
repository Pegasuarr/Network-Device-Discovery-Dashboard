package service

import (
	"context"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"github.com/user/network-monitoring/internal/model"
	"github.com/user/network-monitoring/internal/repository"
)

type CronScheduler struct {
	discoveryService *DiscoveryService
}

func NewCronScheduler(ds *DiscoveryService) *CronScheduler {
	return &CronScheduler{discoveryService: ds}
}

func (cs *CronScheduler) Start(ctx context.Context) {
	slog.Info("Starting Cron Scan Schedule background worker...")
	
	// Check every minute
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			slog.Info("Stopping Cron Scan Schedule background worker...")
			return
		case <-ticker.C:
			cs.checkAndRun(ctx)
		}
	}
}

func (cs *CronScheduler) checkAndRun(ctx context.Context) {
	var schedules []model.ScanSchedule
	err := repository.DB.Where("enabled = ?", true).Find(&schedules).Error
	if err != nil {
		slog.Error("Cron worker failed to fetch active schedules", "error", err)
		return
	}

	now := time.Now()
	for _, sched := range schedules {
		if shouldRunCron(sched.CronExpression, now) {
			slog.Info("Cron rule triggered active scan", "rule", sched.Name, "target", sched.Target)
			_, err := cs.discoveryService.StartScan(sched.OrganizationID, sched.Target, sched.ScanProfile, "scheduled")
			if err != nil {
				slog.Error("Cron failed to launch scan", "rule", sched.Name, "error", err)
			}
		}
	}
}

// shouldRunCron matches a cron expression against a given timestamp.
func shouldRunCron(expr string, t time.Time) bool {
	parts := strings.Fields(expr)
	if len(parts) != 5 {
		return false
	}

	minute := t.Minute()
	hour := t.Hour()
	dom := t.Day()
	month := int(t.Month())
	dow := int(t.Weekday()) // Sunday = 0

	return matchCronField(parts[0], minute) &&
		matchCronField(parts[1], hour) &&
		matchCronField(parts[2], dom) &&
		matchCronField(parts[3], month) &&
		matchCronField(parts[4], dow)
}

func matchCronField(field string, value int) bool {
	if field == "*" {
		return true
	}
	
	// Handles steps, e.g. "*/5" (every 5 minutes/hours/etc)
	if strings.HasPrefix(field, "*/") {
		stepStr := strings.TrimPrefix(field, "*/")
		step, err := strconv.Atoi(stepStr)
		if err == nil && step > 0 {
			return value%step == 0
		}
	}
	
	// Handles single integers, e.g. "30"
	if val, err := strconv.Atoi(field); err == nil {
		return val == value
	}
	
	return false
}
