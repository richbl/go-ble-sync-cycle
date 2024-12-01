package ble

import (
	"context"
	"encoding/binary"
	"errors"
	"log"
	"math"
	"time"

	config "ble-sync-cycle/internal/configuration"
	speed "ble-sync-cycle/internal/speed"

	"tinygo.org/x/bluetooth"
)

// BLEController represents the BLE central controller component
type BLEController struct {
	bleConfig   config.BLEConfig
	speedConfig config.SpeedConfig
	bleAdapter  bluetooth.Adapter
}

var (
	// CSC speed tracking variables
	lastWheelRevs uint32
	lastWheelTime uint16
	lastCrankRevs uint16
	lastCrankTime uint16
)

// NewBLEController creates a new BLE central controller for accessing a BLE peripheral
func NewBLEController(bleConfig config.BLEConfig, speedConfig config.SpeedConfig) (*BLEController, error) {

	bleAdapter := bluetooth.DefaultAdapter
	if err := bleAdapter.Enable(); err != nil {
		return nil, err
	}

	log.Println("\\ Created new BLE central controller")

	return &BLEController{
		bleConfig:   bleConfig,
		speedConfig: speedConfig,
		bleAdapter:  *bleAdapter,
	}, nil

}

// GetBLECharacteristic scans for the BLE peripheral and returns CSC services/characteristics
func (m *BLEController) GetBLECharacteristic(ctx context.Context, speedController *speed.SpeedController) (*bluetooth.DeviceCharacteristic, error) {

	// Scan for BLE peripheral
	result, err := m.scanForBLEPeripheral(ctx)
	if err != nil {
		return nil, err
	}

	log.Println("\\ Connecting to BLE peripheral device", result.Address)

	// Connect to BLE peripheral device
	var device bluetooth.Device
	if device, err = m.bleAdapter.Connect(result.Address, bluetooth.ConnectionParams{}); err != nil {
		return nil, err
	}

	log.Println("\\ BLE peripheral device connected")
	log.Println("\\ Discovering CSC services", bluetooth.New16BitUUID(0x1816))

	// Find CSC service and characteristic
	svc, err := device.DiscoverServices([]bluetooth.UUID{bluetooth.New16BitUUID(0x1816)})
	if err != nil {
		log.Println("\\ CSC services discovery failed:", err)
		return nil, err
	}

	log.Println("\\ Found CSC service", svc[0].UUID().String())
	log.Println("\\ Discovering CSC characteristics", bluetooth.New16BitUUID(0x2A5B))

	char, err := svc[0].DiscoverCharacteristics([]bluetooth.UUID{bluetooth.New16BitUUID(0x2A5B)})
	if err != nil {
		log.Println("\\ CSC characteristics discovery failed:", err)
		return nil, err
	}

	log.Println("\\ Found CSC characteristic", char[0].UUID().String())

	return &char[0], nil

}

// GetBLEUpdates enables BLE peripheral monitoring to report real-time sensor data
func (m *BLEController) GetBLEUpdates(ctx context.Context, speedController *speed.SpeedController, char *bluetooth.DeviceCharacteristic) error {

	log.Println("\\ Starting real-time monitoring of BLE sensor notifications...")

	// Subscribe to live BLE sensor notifications
	if err := char.EnableNotifications(func(buf []byte) {
		speed := m.processBLESpeed(buf)
		speedController.UpdateSpeed(speed)
	}); err != nil {
		return err
	}

	<-ctx.Done()
	return nil

}

// scanForBLEPeripheral scans for the specified BLE peripheral UUID within the given timeout
func (m *BLEController) scanForBLEPeripheral(ctx context.Context) (bluetooth.ScanResult, error) {

	scanCtx, cancel := context.WithTimeout(ctx, time.Duration(m.bleConfig.ScanTimeoutSecs)*time.Second)
	defer cancel()

	found := make(chan bluetooth.ScanResult, 1)

	go func() {

		log.Println("\\ Now scanning the ether for BLE peripheral UUID of", m.bleConfig.SensorUUID, "...")

		err := m.bleAdapter.Scan(func(adapter *bluetooth.Adapter, result bluetooth.ScanResult) {

			if result.Address.String() == m.bleConfig.SensorUUID {

				if err := m.bleAdapter.StopScan(); err != nil {
					log.Println("\\ Failed to stop scan:", err)
				}

				found <- result
			}

		})

		if err != nil {
			log.Println("\\ Scan error:", err)
		}

	}()

	// Wait for the scan to complete or context cancellation
	select {
	case result := <-found:
		log.Println("\\ Found BLE peripheral", result.Address.String())
		return result, nil
	case <-scanCtx.Done():
		if err := m.bleAdapter.StopScan(); err != nil {
			log.Println("\\ Failed to stop scan:", err)
		}
		return bluetooth.ScanResult{}, errors.New("scanning time limit reached")
	}

}

// processBLESpeed processes raw BLE CSC speed data and returns the adjusted current speed
func (m *BLEController) processBLESpeed(data []byte) float64 {

	if len(data) < 1 {
		return 0.0
	}

	log.Println("| Processing speed data from BLE peripheral...")

	flags := data[0]
	hasWheelRev := (flags & 0x01) != 0
	hasCrankRev := (flags & 0x02) != 0
	offset := 1
	var speed float64

	// Calculate wheel speed
	if hasWheelRev && len(data) >= offset+6 {
		wheelRevs := binary.LittleEndian.Uint32(data[offset:])
		wheelEventTime := binary.LittleEndian.Uint16(data[offset+4:])

		// Calculate speed if we have previous measurements
		if lastWheelTime != 0 {

			timeDiff := uint16(wheelEventTime - lastWheelTime)

			if timeDiff != 0 {
				revDiff := wheelRevs - lastWheelRevs

				// Convert speed units (kph)
				speedConversion := 3.6

				// Convert speed units (mph)
				if m.speedConfig.SpeedUnits == "mph" {
					speedConversion = 2.23694
				}

				// Calculate speed
				speed = float64(revDiff) * float64(m.speedConfig.WheelCircumferenceMM) * speedConversion / float64(timeDiff)

				log.Println("| BLE sensor speed:", math.Round(speed*100)/100, m.speedConfig.SpeedUnits)
			}

		}

		lastWheelRevs = wheelRevs
		lastWheelTime = wheelEventTime
		offset += 6

	}

	// Calculate crank speed (future functionality)
	//
	if hasCrankRev && len(data) >= offset+4 {

		log.Println("Crank/cadence event!")
		crankRevs := binary.LittleEndian.Uint16(data[offset:])
		crankEventTime := binary.LittleEndian.Uint16(data[offset+2:])

		// Calculate cadence if we have previous measurements
		if lastCrankTime != 0 {

			// Handle timer wraparound (16-bit timer)
			timeDiff := uint16(crankEventTime - lastCrankTime)

			if timeDiff != 0 {
				revDiff := crankRevs - lastCrankRevs

				// Calculate cadence (RPM)
				cadence := float64(revDiff) * 60 * 1024 / float64(timeDiff)
				log.Println("| BLE sensor cadence:", math.Round(cadence*100)/100, "RPM")
			}

		}

		lastCrankRevs = crankRevs
		lastCrankTime = crankEventTime
	}

	return speed

}
