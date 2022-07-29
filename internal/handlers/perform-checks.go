package handlers

import (
	"encoding/json"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/tsawler/vigilate/internal/models"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const (
	HTTP           = 1
	HTTPS          = 2
	SSLCertificate = 3
)

// jsonResp describes the JSON response sent back to client
type jsonResp struct {
	OK            bool      `json:"ok"`
	Message       string    `json:"message"`
	ServiceID     int       `json:"service_id"`
	HostServiceID int       `json:"host_service_id"`
	HostID        int       `json:"host_id"`
	OldStatus     string    `json:"old_status"`
	NewStatus     string    `json:"new_status"`
	LastCheck     time.Time `json:"last_check"`
}

// ScheduleCheck performs a scheduled check on a host service by id
func (repo *DBRepo) ScheduleCheck(hostServiceID int) {
	log.Println("********** Running check for ", hostServiceID)

	hs, err := repo.DB.GetHostServiceByID(hostServiceID)
	if err != nil {
		log.Println(err)
		return
	}

	h, err := repo.DB.GetHostByID(hs.HostID)
	if err != nil {
		log.Println(err)
		return
	}

	// tests the service
	newStatus, msg := repo.testServiceForHost(h, hs)

	hostServiceStatusChanged := false
	if newStatus != hs.Status {
		hostServiceStatusChanged = true
	}

	// if the host service status has changed, broadcast to all clients
	if hostServiceStatusChanged {
		data := make(map[string]string)
		data["message"] = fmt.Sprintf("host service %s on %s has changed to %s", hs.Service.ServiceName, h.HostName, newStatus)

		repo.broadcastMessage("public-channel", "host-service-status-changed", data)
		// if appropriate, send email or SMS message
	}

	// update host service record in db with status (if changed) and
	// update last-check
	hs.Status = newStatus
	hs.LastCheck = time.Now()
	err = repo.DB.UpdateHostService(hs)
	if err != nil {
		log.Println(err)
		return
	}

	if hostServiceStatusChanged {
		pending, healthy, warning, problem, err := repo.DB.GetAllServiceStatusCounts()
		if err != nil {
			log.Println(err)
			return
		}

		data := make(map[string]string)
		data["healthy_count"] = strconv.Itoa(healthy)
		data["pending_count"] = strconv.Itoa(pending)
		data["warning_count"] = strconv.Itoa(warning)
		data["problem_count"] = strconv.Itoa(problem)
		repo.broadcastMessage("public-channel", "host-service-count-changed", data)
	}

	log.Println("New status is", newStatus, "and msg is", msg)
}

func (repo DBRepo) broadcastMessage(channel, messageType string, data map[string]string) {
	err := app.WsClient.Trigger(channel, messageType, data)
	if err != nil {
		log.Println(err)
	}
}

// TestCheck manually tests a host service and sends JSON response
func (repo *DBRepo) TestCheck(w http.ResponseWriter, r *http.Request) {
	hostServiceID, _ := strconv.Atoi(chi.URLParam(r, "id"))
	oldStatus := chi.URLParam(r, "oldStatus")
	okay := true

	// get host service
	hs, err := repo.DB.GetHostServiceByID(hostServiceID)
	if err != nil {
		log.Println(err)
		okay = false
	}

	// get host?
	h, err := repo.DB.GetHostByID(hs.HostID)
	if err != nil {
		log.Println(err)
		okay = false
	}

	// test the service
	newStatus, msg := repo.testServiceForHost(h, hs)

	// update the host service in the db with status (if changed) and last check
	hs.Status = newStatus
	hs.LastCheck = time.Now()
	hs.UpdatedAt = time.Now()

	err = repo.DB.UpdateHostService(hs)
	if err != nil {
		log.Println(err)
		okay = false
	}

	// broadcast service status changed event

	var resp jsonResp

	// create json
	if okay {
		resp = jsonResp{
			OK:            true,
			Message:       msg,
			ServiceID:     hs.ServiceID,
			HostServiceID: hs.ID,
			HostID:        hs.HostID,
			OldStatus:     oldStatus,
			NewStatus:     newStatus,
			LastCheck:     time.Now(),
		}
	} else {
		resp.OK = false
		resp.Message = "Something went wrong"
	}

	// send json to client
	out, _ := json.MarshalIndent(resp, "", "  ")
	w.Header().Set("Content-Type", "application/json")
	w.Write(out)
}

// testServiceForHost tests a service for a host
func (repo *DBRepo) testServiceForHost(h models.Host, hs models.HostService) (string, string) {
	var msg, newStatus string

	switch hs.ServiceID {
	case HTTP:
		msg, newStatus = testHTTPForHost(h.URL)
		break
	}

	return newStatus, msg
}

func testHTTPForHost(url string) (string, string) {
	if strings.HasSuffix(url, "/") {
		url = strings.TrimSuffix(url, "/")
	}

	url = strings.Replace(url, "https://", "http://", -1)

	resp, err := http.Get(url)
	if err != nil {
		return fmt.Sprintf("%s - %s", url, "error connecting"), "problem"
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Sprintf("%s - %s", url, resp.Status), "problem"
	}

	return fmt.Sprintf("%s - %s", url, resp.Status), "healthy"
}
