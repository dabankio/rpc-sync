package pow

import (
	"bbcsyncer/infra"
	"bbcsyncer/reward"
	"bbcsyncer/sync"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/dabankio/civil"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
)

func NewHandler(repo *Repo, infraRapo *infra.Repo, syncRepo *sync.Repo) *Handler {
	return &Handler{repo: repo, infraRapo: infraRapo, syncRepo: syncRepo}
}

type Handler struct {
	infra.BaseHandler
	repo      *Repo
	infraRapo *infra.Repo
	syncRepo  *sync.Repo
}

func (h *Handler) CreateUnlockedBlocks(w http.ResponseWriter, req *http.Request) {
	// TODO 客户端认证

	var reqModel ReqUnlockedBlocks
	// if !h.BindBody(w, req, &reqModel) {
	// 	return
	// }

	b, err := io.ReadAll(req.Body)
	if err != nil {
		h.WriteErr(w, err)
		return
	}
	err = h.infraRapo.CreateApiLog(infra.ApiLog{
		Time:   time.Now(),
		URL:    req.URL.String(),
		Method: req.Method,
		IP:     h.RealIP(req),
		Body:   string(b),
	})
	if err != nil {
		log.Println("[warn] write request to db failed ,unlocked blocks request:", string(b))
	}

	err = json.Unmarshal(b, &reqModel)
	if err != nil {
		h.WriteErr(w, err)
		return
	}

	var heights []uint64
	for i := 0; i < len(reqModel.BalanceLst); i++ {
		heights = append(heights, reqModel.BalanceLst[i].Height)
	}
	blocks, err := h.syncRepo.BlocksInHeight(heights...)
	if err != nil {
		h.WriteErr(w, err)
		return
	}

	blocksMap := sync.NewHeightBlockMap(blocks)

	// 从实际的api数据来看，date字段似乎不可靠（不准确，存在不确定的时区歧义
	// time_span 似乎和实际的区块时间有1分钟的偏差
	// 考虑到height是可靠的，不存在歧义偏差的，故使用height结合本地数据来确认奖励的日期

	// 一个潜在的问题是，如果api调用时区块同步没有完成，这样可能导致查不到区块数据从而无法确定时间，处理方案：用数据库缓存 api调用数据，当发现数据有问题时 等待区块同步然后手工重放请求写入数据
	// 缓存api调用数据有另一个好处，可以追溯api调用记录，方便查找问题
	// 另一个潜在的问题是，如果区块因为分叉被丢掉了，那么pow这个块的奖励数据也应该删掉(是不是考虑用外键约束的级连删除实现, 级连删除也有问题，高度是始终存在的，区块hash才是分叉丢弃的)(或者直接不管，出问题了再手动删掉)

	var items []UnlockedBlock
	for _, x := range reqModel.BalanceLst {
		block, ok := blocksMap[x.Height]
		if !ok {
			e := errors.Errorf("block of height %d not found", x.Height)
			log.Println("[err]", e)
			h.WriteErr(w, e)
			return
		}
		items = append(items, UnlockedBlock{
			AddrFrom:  x.AddrFrom,
			AddrTo:    x.AddrTo,
			Balance:   x.Balance,
			TimeSpan:  x.TimeSpan, //虽然不准确，还是保存下来吧，方便比较数据
			Day:       x.Date,     //虽然不准确，还是保存下来吧，方便比较数据
			RewardDay: civil.DateOf(block.Time.In(reward.ZoneBeijingTime)),
			Height:    x.Height,
		})
	}
	err = infra.RunInTx(h.repo.DB, func(tx *sqlx.Tx) error {
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
