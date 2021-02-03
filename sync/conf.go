package sync

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"log"

	"github.com/dabankio/bbrpc"
)

type Conf struct {
	DB          string `json:"db,omitempty"`
	RPCUrl      string `json:"rpc_url,omitempty"`
	RPCUsr      string `json:"rpc_usr"`
	RPCPassword string `json:"rpc_password,omitempty"`
}

func NewBBCClient(conf Conf) (*bbrpc.Client, error) {
	return bbrpc.NewClient(&bbrpc.ConnConfig{
		Host:       conf.RPCUrl,
		User:       conf.RPCUsr,
		Pass:       conf.RPCPassword,
		DisableTLS: true,
	})
}

func ParseConf() (c Conf) {
	var confFile string
	if !flag.Parsed() {
		flag.StringVar(&confFile, "conf", "./conf.json", "-conf=/etc/sync_conf.json")
		flag.Parse()
	}
	b, err := ioutil.ReadFile(confFile)
	PanicErr(err)
	PanicErr(json.Unmarshal(b, &c))
	log.Println("conf loaded from : ", confFile)
	return
}
