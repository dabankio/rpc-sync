package infra

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/dabankio/bbrpc"
)

type Conf struct {
	DB          string `json:"db,omitempty"`
	RPCUrl      string `json:"rpc_url,omitempty"`
	RPCUsr      string `json:"rpc_usr"`
	RPCPassword string `json:"rpc_password,omitempty"`
}

func NewBBCClient(conf Conf) (*bbrpc.Client, error) {
	httpClient := &http.Client{
		Timeout:   30 * time.Second,
		Transport: &http.Transport{MaxConnsPerHost: 1},
	}
	return bbrpc.NewClientWith(&bbrpc.ConnConfig{
		Host:       conf.RPCUrl,
		User:       conf.RPCUsr,
		Pass:       conf.RPCPassword,
		DisableTLS: true,
	}, httpClient)
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
