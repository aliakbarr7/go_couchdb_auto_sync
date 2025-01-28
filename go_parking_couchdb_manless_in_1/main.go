package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/fjl/go-couchdb"
)

const (
	couchDBURL      = "http://ali:1234@127.0.0.1:5984"
	dbName          = "parking-system"
	gateType        = "manless_in_1"
	ratePerHour int = 2000
)

var db *couchdb.DB

func init() {
	client, err := couchdb.NewClient(couchDBURL, nil)
	if err != nil {
		log.Fatalf("Gagal menghubungkan ke CouchDB: %v", err)
	}

	if _, err := client.CreateDB(dbName); err != nil {
		if !strings.Contains(err.Error(), "file_exists") {
			log.Fatalf("Gagal membuat database: %v", err)
		}
	}
	db = client.DB(dbName)
}

type Vehicle struct {
	ID              string  `json:"_id,omitempty"`
	Rev             string  `json:"_rev,omitempty"`
	CarNumber       string  `json:"car_number"`
	InTime          string  `json:"in_time"`
	OutTime         *string `json:"out_time"`
	Status          string  `json:"status"`
	GateIn          string  `json:"gate_in"`
	GateOut         *string `json:"gate_out"`
	ParkingDuration *int    `json:"parking_duration"`
	ParkingRates    *int    `json:"parking_rates"`
}

func main() {
	showMenu()
}

func showMenu() {
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Println()
		fmt.Println("=== Sistem Parkir ===")
		fmt.Println("1. Tampilkan semua database")
		fmt.Println("2. Tampilkan database berdasarkan ID")
		fmt.Println("3. Tambahkan kendaraan masuk")
		fmt.Println("4. Perbarui kendaraan keluar")
		fmt.Println("5. Hapus kendaraan berdasarkan ID")
		fmt.Println("6. Keluar")
		fmt.Println()

		fmt.Print("Pilih opsi (1-6): ")
		choice, _ := reader.ReadString('\n')
		choice = strings.TrimSpace(choice)

		switch choice {
		case "1":
			showAllDatabase()
		case "2":
			fmt.Print("Masukkan ID kendaraan: ")
			id, _ := reader.ReadString('\n')
			id = strings.TrimSpace(id)
			showDatabaseByID(id)
		case "3":
			fmt.Print("Masukkan nomor kendaraan: ")
			carNumber, _ := reader.ReadString('\n')
			carNumber = strings.TrimSpace(carNumber)
			addVehicle(carNumber)
		case "4":
			fmt.Print("Masukkan ID kendaraan: ")
			id, _ := reader.ReadString('\n')
			id = strings.TrimSpace(id)
			updateVehicle(id)
		case "5":
			fmt.Print("Masukkan ID kendaraan yang ingin dihapus: ")
			id, _ := reader.ReadString('\n')
			id = strings.TrimSpace(id)
			deleteVehicle(id)
		case "6":
			fmt.Println("Keluar dari sistem parkir.")
			return
		default:
			fmt.Println("Opsi tidak valid. Silakan pilih lagi.")
		}
	}
}

func showAllDatabase() {
	var rows struct {
		Rows []struct {
			ID  string  `json:"id"`
			Doc Vehicle `json:"doc"`
		} `json:"rows"`
	}

	err := db.AllDocs(&rows, couchdb.Options{"include_docs": true})
	if err != nil {
		fmt.Printf("Gagal mengambil data: %v\n", err)
		return
	}

	fmt.Println()
	fmt.Println("=== Isi Database ===")
	fmt.Println()
	for _, row := range rows.Rows {
		prettyPrintJSON(row.Doc)
		fmt.Println()
	}
	fmt.Println("====================")
	fmt.Println()

	promptReturnToMenu()
}

func showDatabaseByID(id string) {
	var vehicle Vehicle
	if err := db.Get(id, &vehicle, nil); err != nil {
		fmt.Printf("Gagal mendapatkan data: %v\n", err)
	} else {
		fmt.Println("=== Data Kendaraan ===")
		prettyPrintJSON(vehicle)
	}

	promptReturnToMenu()
}

func addVehicle(carNumber string) {

	id := fmt.Sprintf("%02d%s", rand.Intn(100), carNumber)

	var rows struct {
		Rows []struct {
			Doc Vehicle `json:"doc"`
		} `json:"rows"`
	}

	if err := db.AllDocs(&rows, couchdb.Options{"include_docs": true}); err == nil {
		for _, row := range rows.Rows {
			if row.Doc.CarNumber == carNumber && row.Doc.Status == "in" {
				fmt.Printf("Kendaraan dengan plat nomor %s sudah masuk dan belum keluar.\n", carNumber)
				promptReturnToMenu()
				return
			}
		}
	}

	vehicle := Vehicle{
		ID:              id,
		CarNumber:       carNumber,
		InTime:          time.Now().Format(time.RFC3339),
		Status:          "in",
		GateIn:          gateType,
		GateOut:         nil,
		OutTime:         nil,
		ParkingDuration: nil,
		ParkingRates:    nil,
	}

	if _, err := db.Put(vehicle.ID, &vehicle, ""); err != nil {
		fmt.Printf("Gagal menambahkan kendaraan: %v\n", err)
	} else {
		fmt.Printf("Kendaraan %s berhasil masuk melalui %s.\n", carNumber, gateType)
		prettyPrintJSON(vehicle)
	}

	promptReturnToMenu()
}

func updateVehicle(id string) {
	var vehicle Vehicle
	if err := db.Get(id, &vehicle, nil); err != nil {
		fmt.Printf("Gagal menemukan kendaraan dengan ID %s: %v\n", id, err)
		return
	}

	if vehicle.Status == "out" {
		fmt.Printf("Kendaraan dengan ID %s sudah keluar sebelumnya.\n", id)
		promptReturnToMenu()
		return
	}

	now := time.Now()
	vehicle.OutTime = stringPtr(now.Format(time.RFC3339))
	vehicle.Status = "out"
	vehicle.GateOut = stringPtr(gateType)

	inTime, err := time.Parse(time.RFC3339, vehicle.InTime)
	if err != nil {
		fmt.Printf("Gagal mem-parsing waktu masuk: %v\n", err)
		return
	}

	durationMinutes := int(now.Sub(inTime).Minutes())
	vehicle.ParkingDuration = &durationMinutes

	hours := (durationMinutes + 59) / 60
	rates := hours * ratePerHour
	vehicle.ParkingRates = &rates

	if _, err := db.Put(vehicle.ID, &vehicle, vehicle.Rev); err != nil {
		fmt.Printf("Gagal memperbarui kendaraan: %v\n", err)
	} else {
		fmt.Printf("Kendaraan %s berhasil keluar melalui %s.\n", id, gateType)
		prettyPrintJSON(vehicle)
	}

	promptReturnToMenu()
}

func deleteVehicle(id string) {
	var vehicle Vehicle
	if err := db.Get(id, &vehicle, nil); err != nil {
		fmt.Printf("Kendaraan tidak ditemukan: %v\n", err)
		return
	}

	if _, err := db.Delete(vehicle.ID, vehicle.Rev); err != nil {
		fmt.Printf("Gagal menghapus kendaraan: %v\n", err)
	} else {
		fmt.Printf("Kendaraan %s berhasil dihapus.\n", vehicle.CarNumber)
	}

	promptReturnToMenu()
}

func prettyPrintJSON(data interface{}) {
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		fmt.Printf("Gagal memformat JSON: %v\n", err)
		return
	}
	fmt.Println(string(jsonData))
}

func promptReturnToMenu() {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Kembali ke menu (y/n)? ")
	choice, _ := reader.ReadString('\n')
	choice = strings.TrimSpace(choice)

	if strings.ToLower(choice) == "y" {
		showMenu()
	} else {
		fmt.Println("Keluar dari sistem parkir.")
		os.Exit(0)
	}
}

func stringPtr(s string) *string {
	return &s
}
