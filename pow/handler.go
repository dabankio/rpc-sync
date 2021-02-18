package pow

import (
	"bbcsyncer/infra"
	"net/http"

	"github.com/dabankio/civil"
	"github.com/jmoiron/sqlx"
)

func NewHandler(repo *Repo) *Handler { return &Handler{repo: repo} }

type Handler struct {
	infra.BaseHandler
	repo *Repo
}

func (h *Handler) CreateUnlockedBlocks(w http.ResponseWriter, req *http.Request) {
	// TODO 客户端认证

	var reqModel ReqUnlockedBlocks
	if !h.BindBody(w, req, &reqModel) {
		return
	}

	var items []UnlockedBlock
	for _, x := range reqModel.BalanceLst {
		items = append(items, UnlockedBlock{
			AddrFrom: x.AddrFrom,
			AddrTo:   x.AddrTo,
			Balance:  x.Balance,
			TimeSpan: x.TimeSpan,
			Day:      civil.DateOf(x.Date),
			Height:   x.Height,
		})
	}
	err := infra.RunInTx(h.repo.DB, func(tx *sqlx.Tx) error {
		return h.repo.InsertUnlockedBlocks(items, tx)
	})
	if err != nil {
		h.WriteErr(w, err)
		return
	}
	h.WriteJSON(w, infra.LegacyResult{
		Success: true,
		Code:    infra.LegacyCodeSuccess,
		Message: "OK",
	})
}

func (h *Handler) Query(w http.ResponseWriter, req *http.Request) {
	addrFrom := req.URL.Query().Get("addrFrom")
	dateStr := req.URL.Query().Get("date")

	if addrFrom == "" || dateStr == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("empty param"))
		return
	}
}
