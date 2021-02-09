package pow

import (
	"bbcsyncer/infra"
	"encoding/json"
	"io"
	"net/http"

	"github.com/dabankio/civil"
)

func NewHandler(repo *Repo) *Handler { return &Handler{repo: repo} }

type Handler struct {
	infra.BaseHandler
	repo *Repo
}

func (h *Handler) CreateUnlockedBlocks(w http.ResponseWriter, req *http.Request) {
	// TODO 客户端认证

	b, err := io.ReadAll(req.Body)
	if err != nil {
		h.WriteErr(w, err)
		return
	}

	var reqModel ReqUnlockedBlocks
	err = json.Unmarshal(b, &reqModel)
	if err != nil {
		h.WriteErr(w, err)
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
	err = h.repo.InsertUnlockedBlocks(items)
	if err != nil {
		h.WriteErr(w, err)
		return
	}
	h.WriteJSON(w, "ok")
}
