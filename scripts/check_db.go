package main

import (
	"fmt"

	"github.com/gibran/go-gin-boilerplate/config"
	"github.com/gibran/go-gin-boilerplate/database"
	"github.com/gibran/go-gin-boilerplate/internal/model"
)

func main() {
	cfg := config.Load()
	db := database.Connect(cfg)

	var devices []model.Device
	db.Find(&devices)

	fmt.Println("=== REGISTERED DEVICES ===")
	if len(devices) == 0 {
		fmt.Println("(no devices found)")
	}
	for _, d := range devices {
		fmt.Printf("MAC: %-20s | Name: %-20s | Approved: %v | UserID: %s | CertThumb: %s\n", d.MacAddress, d.Name, d.IsApproved, d.UserID, d.CertThumb)
	}
}
