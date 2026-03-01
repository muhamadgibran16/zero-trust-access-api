package device

import (
	"github.com/gibran/go-gin-boilerplate/internal/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// DeviceRepository defines database operations for MDM devices
type DeviceRepository struct {
	db *gorm.DB
}

// NewDeviceRepository creates a new DeviceRepository
func NewDeviceRepository(db *gorm.DB) *DeviceRepository {
	return &DeviceRepository{db: db}
}

// Create inserts a new device
func (r *DeviceRepository) Create(device *model.Device) error {
	return r.db.Create(device).Error
}

// FindByMAC finds a device by its MAC address
func (r *DeviceRepository) FindByMAC(mac string) (*model.Device, error) {
	var device model.Device
	err := r.db.Where("mac_address = ?", mac).First(&device).Error
	if err != nil {
		return nil, err
	}
	return &device, nil
}

// FindByUserID retrieves all devices registered by a user
func (r *DeviceRepository) FindByUserID(userID uuid.UUID) ([]model.Device, error) {
	var devices []model.Device
	err := r.db.Where("user_id = ?", userID).Find(&devices).Error
	return devices, err
}

// FindAll retrieves all devices globally (for admin view)
func (r *DeviceRepository) FindAll() ([]model.Device, error) {
	var devices []model.Device
	err := r.db.Preload("User").Find(&devices).Error
	return devices, err
}

// Update modifies device details (e.g. IsApproved)
func (r *DeviceRepository) Update(device *model.Device) error {
	return r.db.Save(device).Error
}
