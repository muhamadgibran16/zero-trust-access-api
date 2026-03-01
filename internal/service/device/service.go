package device

import (
	"errors"

	"github.com/gibran/go-gin-boilerplate/internal/model"
	repository "github.com/gibran/go-gin-boilerplate/internal/repository/device"
	"github.com/google/uuid"
)

// DeviceService handles MDM logic
type DeviceService struct {
	repo *repository.DeviceRepository
}

// NewDeviceService creates a new DeviceService
func NewDeviceService(repo *repository.DeviceRepository) *DeviceService {
	return &DeviceService{repo: repo}
}

type RegisterDeviceRequest struct {
	MacAddress string `json:"macAddress" binding:"required"`
	Name       string `json:"name" binding:"required"`
}

// RegisterDevice registers a new device pending admin approval
func (s *DeviceService) RegisterDevice(userID uuid.UUID, req RegisterDeviceRequest) (*model.Device, error) {
	// Check if already registered
	existing, _ := s.repo.FindByMAC(req.MacAddress)
	if existing != nil {
		if existing.UserID != userID {
			return nil, errors.New("this hardware identifier is already registered to another user")
		}
		return existing, nil
	}

	device := &model.Device{
		UserID:     userID,
		MacAddress: req.MacAddress,
		Name:       req.Name,
		IsApproved: false, // Default to false in Zero Trust
	}

	err := s.repo.Create(device)
	return device, err
}

// GetUserDevices returns all devices for a given user
func (s *DeviceService) GetUserDevices(userID uuid.UUID) ([]model.Device, error) {
	return s.repo.FindByUserID(userID)
}

// GetAllDevices returns all devices globally for Admin view
func (s *DeviceService) GetAllDevices() ([]model.Device, error) {
	return s.repo.FindAll()
}

// ApproveDevice sets the trust flag for a device to True
func (s *DeviceService) ApproveDevice(mac string) error {
	device, err := s.repo.FindByMAC(mac)
	if err != nil {
		return errors.New("device not found")
	}

	device.IsApproved = true
	return s.repo.Update(device)
}

// RejectDevice sets the trust flag for a device to False
func (s *DeviceService) RejectDevice(mac string) error {
	device, err := s.repo.FindByMAC(mac)
	if err != nil {
		return errors.New("device not found")
	}

	device.IsApproved = false
	return s.repo.Update(device)
}
