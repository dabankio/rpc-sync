package pow

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"
	"time"

	"github.com/shopspring/decimal"
	r "github.com/stretchr/testify/require"
)

func TestOriginAPI(t *testing.T) {
	// INSERT INTO `AppInfo` (id,appID,appName,addTime,secretKey,note) VALUES (1, 'som_app', 'app_name_x', now(), "abc_key", "notex")

	req := ReqUnlockedBlocks{}
	req.AppID = "som_app"
	req.TimeSpan = fmt.Sprintf("%d", time.Now().Unix())
	req.SignPlain = "any_plainx"
	{
		raw := fmt.Sprintf("%s:%s:%s", req.AppID, req.TimeSpan, req.SignPlain)
		fmt.Println("raw:", raw)
		h := hmac.New(sha256.New, []byte("abc_key"))
		h.Write([]byte(raw))
		req.RequestSign = hex.EncodeToString(h.Sum(nil))
	}
	req.BalanceLst = append(req.BalanceLst, UnlockBlock{
		UnlockBlockBase: UnlockBlockBase{
			AddrFrom: "from_addx",
			// Date:     "2021-02-07T02:46:25.948Z",
			Date: time.Now(),
		},
		AddrTo:   "123",
		Balance:  decimal.NewFromFloat(2.33),
		TimeSpan: time.Now().Unix(),
		Height:   999,
	})

	reqB, err := json.Marshal(req)
	r.NoError(t, err)
	fmt.Println("request bytes: ", string(reqB))

	// httpReq, err := http.NewRequest(http.MethodPost, "http://localhost:10003/api/UnlockBblock", bytes.NewReader(reqB))
	httpReq, err := http.NewRequest(http.MethodPost, "http://localhost:7777/api/UnlockBblock", bytes.NewReader(reqB))
	r.NoError(t, err)
	httpReq.Header.Add("Content-type", "application/json")

	resp, err := http.DefaultClient.Do(httpReq)
	r.NoError(t, err)
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	r.NoError(t, err)
	fmt.Println("resp:", string(body))
}
