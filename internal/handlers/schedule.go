package handlers

import (
	"fmt"
	"github.com/CloudyKit/jet/v6"
	"github.com/tsawler/vigilate/internal/helpers"
	"github.com/tsawler/vigilate/internal/models"
	"log"
	"net/http"
	"sort"
)

type ByHost []models.Schedule

// Len is used to sort by host
func (h ByHost) Len() int {
	return len(h)
}

// Less is used to sort by host
func (h ByHost) Less(i, j int) bool {
	return h[i].Host < h[j].Host
}

// Swap is used to sort by host
func (h ByHost) Swap(i, j int) {
	h[i], h[j] = h[j], h[i]
}

// ListEntries lists schedule entries
func (repo *DBRepo) ListEntries(w http.ResponseWriter, r *http.Request) {
	var items []models.Schedule

	for k, v := range repo.App.MonitorMap {
		var item models.Schedule

		item.ID = k
		item.EntryID = v
		item.Entry = repo.App.Scheduler.Entry(v)
		hs, err := repo.DB.GetHostServiceByID(k)
		if err != nil {
			log.Println(err)
			return
		}

		item.ScheduleText = fmt.Sprintf("@every %d%s", hs.ScheduleNumber, hs.ScheduleUnit)
		item.LastRunFromHS = hs.LastCheck
		item.Host = hs.HostName
		item.Service = hs.Service.ServiceName
		items = append(items, item)
	}

	// sort the slices
	sort.Sort(ByHost(items))

	data := make(jet.VarMap)

	data.Set("items", items)

	err := helpers.RenderPage(w, r, "schedule", data, nil)
	if err != nil {
		printTemplateError(w, err)
	}
}
