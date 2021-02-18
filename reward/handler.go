package reward

import (
	"bbcsyncer/infra"
	"net/http"

	"github.com/dabankio/civil"
	"github.com/shopspring/decimal"
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

	dposAddr := r.URL.Query().Get("dpos_addr")

	items, err := h.repo.DailyRewardsOfDay(day, dposAddr)
	if err != nil {
		h.WriteErr(w, err)
		return
	}
	var resp RespDailyRewardEDPoS
	for _, x := range items {
		resp.Data = append(resp.Data, DailyRewardEDPoS{
			DposAddr:     x.Delegate,
			ClientAddr:   x.Voter,
			PaymentDate:  x.Day.String(),
			PaymentMoney: x.Amount,
		})
	}
	h.WriteJSON(w, resp)
}

type RespDailyRewardEDPoS struct {
	Data []DailyRewardEDPoS `json:"Data"`
}

type DailyRewardEDPoS struct {
	ID           int             `json:"id"` //id 无用，此处只是为了看起来和bbcrewarder一样
	DposAddr     string          `json:"dpos_addr"`
	ClientAddr   string          `json:"client_addr,omitempty"`
	Txid         string          `json:"txid,omitempty"` //无用
	PaymentDate  string          `json:"payment_date,omitempty"`
	PaymentMoney decimal.Decimal `json:"payment_money"`
}
