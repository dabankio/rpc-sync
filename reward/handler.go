package reward

import (
	"bbcsyncer/infra"
	"net/http"

	"github.com/dabankio/civil"
)

func NewHandler(repo *Repo) *Handler { return &Handler{repo: repo} }

type Handler struct {
	infra.BaseHandler
	repo *Repo
}

func (h *Handler) GetDailyDposRewards(w http.ResponseWriter, r *http.Request) {
	day, err := civil.ParseDate(r.URL.Query().Get("day"))
	if err != nil {
		h.WriteErr(w, err)
		return
	}
	items, err := h.repo.DailyRewardsOfDay(day)
	if err != nil {
		h.WriteErr(w, err)
		return
	}
	h.WriteJSON(w, items)
}
