package main

import (
	"testing"
)

func Test_getVote(t *testing.T) {
	type args struct {
		tplHex string
	}
	tests := []struct {
		name  string
		args  args
		want  string
		want1 string
	}{
		// TODO: Add test cases.
		{
			name:  "tx:5fc73379ae525b4e3e225f291c3ccd4ee20a1f50d325c6038f4c0dbf8e9c2a33",
			args:  args{"0205001e207e335a8937e89334c61de06efa1ea60a541783a1f2b744428892698901cf7298557dc9eaf58284ee29c9f06ee05bf82634d94ce025b56b438d016bbf3c"},
			want:  "20m01w83y6dd8jdz8jctcc7f0dvx1x9gaagbr78fjpx24524jd64jkbs5",
			want1: "1sxs9gnbxs7nfb0m4xrmwkw3ew1dzg9hmv56e09dndd1rt0bbqwy9f6gv",
		},
		{
			name:  "tx:5f2a4dcd5d177e2e59738ce620e6b262d898a68e6e5b9c7511662c82253a723d",
			args:  args{"020500d0593057abcb9a68096722551c46b5e6b1e8050c7cc9961ff32f524f59c2020200036a246dcae027949a9f86686bb5480deddba262b6df33b80739100657bd"},
			want:  "20m0d0p9gaynwq6k815kj4n8w8ttydcf80m67sjcp3zsjymjfb7162307",
			want1: "208006th4dq5e09wmkafrct3bpn40vvevm9hbdqskq03kj406ayyx9y8m",
		},
		{
			name:  "tx:5eccb2cd018407e2684db21ab8a309996f905f6ea37912cd7f1251d262596c5d",
			args:  args{"0205004a355729edbdea02783c7b8a0b43d271c4430c363a46a3ae1ff86a65b4da0110c8531d13e1683d2720fcc423b2183bfcea0ce368fb640b7aab4f40c688c54e"},
			want:  "20m04mdaq57pvvtg2f0y7q2gb8f973h231gv3mhn3nrfzgtk5pkd1m2c7",
			want1: "12345678kw5m3t9s0zk227cgr7fyem373d3xp82vtnd7m1hm8rn76z4mz",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := getVote(tt.args.tplHex)
			// t.Log(got, got1)
			if got != tt.want {
				t.Errorf("getVote() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("getVote() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}
