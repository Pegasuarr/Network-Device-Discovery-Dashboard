package service

import (
	"encoding/csv"
	"fmt"
	"io"
	"strconv"
	"time"

	"github.com/google/uuid"

	"github.com/user/network-monitoring/internal/model"
	"github.com/user/network-monitoring/internal/repository"
)

type DeviceService struct {
	deviceRepo *repository.DeviceRepository
}

func NewDeviceService(deviceRepo *repository.DeviceRepository) *DeviceService {
	return &DeviceService{deviceRepo: deviceRepo}
}

func (s *DeviceService) CreateDevice(d *model.Device) error {
	d.Status = "unknown"
	d.CreatedAt = time.Now()
	d.UpdatedAt = time.Now()
	return s.deviceRepo.Create(d)
}

func (s *DeviceService) GetDevice(id uuid.UUID, orgID uuid.UUID) (*model.Device, error) {
	return s.deviceRepo.FindByID(id, orgID)
}

func (s *DeviceService) ListDevices(orgID uuid.UUID) ([]model.Device, error) {
	return s.deviceRepo.ListByOrg(orgID)
}

func (s *DeviceService) UpdateDevice(d *model.Device) error {
	d.UpdatedAt = time.Now()
	return s.deviceRepo.Update(d)
}

func (s *DeviceService) DeleteDevice(id uuid.UUID, orgID uuid.UUID) error {
	// First reset parent_id of any children of this device to nil
	repository.DB.Model(&model.Device{}).
		Where("parent_id = ? AND organization_id = ?", id, orgID).
		Update("parent_id", nil)

	return s.deviceRepo.Delete(id, orgID)
}

// Groups CRUD
func (s *DeviceService) CreateGroup(g *model.DeviceGroup) error {
	return s.deviceRepo.CreateGroup(g)
}

func (s *DeviceService) ListGroups(orgID uuid.UUID) ([]model.DeviceGroup, error) {
	return s.deviceRepo.ListGroups(orgID)
}

func (s *DeviceService) DeleteGroup(id uuid.UUID, orgID uuid.UUID) error {
	return s.deviceRepo.DeleteGroup(id, orgID)
}

// CSV Export
func (s *DeviceService) ExportCSV(orgID uuid.UUID, w io.Writer) error {
	devices, err := s.deviceRepo.ListByOrg(orgID)
	if err != nil {
		return err
	}

	csvWriter := csv.NewWriter(w)
	defer csvWriter.Flush()

	// Write CSV Header
	header := []string{"Name", "Hostname", "IPAddress", "MACAddress", "DeviceType", "OS", "Vendor", "Location", "MonitoringInterval", "Tags", "Notes", "Enabled"}
	if err := csvWriter.Write(header); err != nil {
		return err
	}

	for _, d := range devices {
		row := []string{
			d.Name,
			d.Hostname,
			d.IPAddress,
			d.MACAddress,
			d.DeviceType,
			d.OS,
			d.Vendor,
			d.Location,
			strconv.Itoa(d.MonitoringInterval),
			d.Tags,
			d.Notes,
			strconv.FormatBool(d.Enabled),
		}
		if err := csvWriter.Write(row); err != nil {
			return err
		}
	}

	return nil
}

// CSV Import
func (s *DeviceService) ImportCSV(orgID uuid.UUID, r io.Reader) (int, error) {
	csvReader := csv.NewReader(r)
	rows, err := csvReader.ReadAll()
	if err != nil {
		return 0, err
	}

	if len(rows) < 2 {
		return 0, fmt.Errorf("CSV is empty or missing data rows")
	}

	importedCount := 0
	for i, row := range rows {
		if i == 0 {
			continue // skip header
		}

		if len(row) < 12 {
			continue // skip invalid rows
		}

		interval, _ := strconv.Atoi(row[8])
		if interval <= 0 {
			interval = 60
		}
		enabled, _ := strconv.ParseBool(row[11])

		d := &model.Device{
			ID:                 uuid.New(),
			OrganizationID:     orgID,
			Name:               row[0],
			Hostname:           row[1],
			IPAddress:          row[2],
			MACAddress:         row[3],
			DeviceType:         row[4],
			OS:                 row[5],
			Vendor:             row[6],
			Location:           row[7],
			MonitoringInterval: interval,
			Tags:               row[9],
			Notes:              row[10],
			Enabled:            enabled,
			Status:             "unknown",
			CreatedAt:          time.Now(),
			UpdatedAt:          time.Now(),
		}

		if err := s.deviceRepo.Create(d); err == nil {
			importedCount++
		}
	}

	return importedCount, nil
}
