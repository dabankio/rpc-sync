package pow

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestHandler_CreateUnlockedBlocks(t *testing.T) {
	b := []byte(` {"balanceLst":[{"addrTo":"1dq62d8y4fz20sfg63zzy4h4ayksswv1fgqjzvegde306bxxg5zygc27q","balance":1,"timeSpan":1613797027,"height":658820,"addrFrom":"20g0epy7jerpbc542a15f99b00mzvex4g3rrkj04crdgzb30b7bp9ncfj","date":"2021-02-20"},{"addrTo":"1r8wv5wg42ftf4gxs38r8vrthkg688b7kretfc2x6p48p88jpm0gfk91k","balance":8.642997,"timeSpan":1613797027,"height":658820,"addrFrom":"20g0epy7jerpbc542a15f99b00mzvex4g3rrkj04crdgzb30b7bp9ncfj","date":"2021-02-20"},{"addrTo":"1jfebkztesj8cw583fch0dyy94x5m6v1f4e3v4vj7188bywcq04e88eh1","balance":0.344817,"timeSpan":1613797027,"height":658820,"addrFrom":"20g0epy7jerpbc542a15f99b00mzvex4g3rrkj04crdgzb30b7bp9ncfj","date":"2021-02-20"}],"requestSign":"9bf0f01f00f3788698181e82b01665c353a395ed03b6150e90555e03179d49b5","appID":"aad7e952-f4ce-4217-909a-903d9d7aa5ed","signPlain":"46691","timeSpan":"1613800715"}`)

	var reqModel ReqUnlockedBlocks
	err := json.Unmarshal(b, &reqModel)
	require.NoError(t, err)
	t.Log(len(reqModel.BalanceLst))
	for i := 0; i < len(reqModel.BalanceLst); i++ {
		t.Logf("%#v\n", reqModel.BalanceLst[i])
	}
}
