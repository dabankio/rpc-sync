package pow

import (
	"bbcsyncer/infra"
	"os"
	"testing"
	"time"

	"github.com/dabankio/civil"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
)

func TestRepo_InsertUnlockedBlocks(t *testing.T) {
	// os.Setenv(infra.TestEnvLocalDB, "5432;unittest;unittest;pwd")

	db := infra.MustNewTestPGDB(t)
	infra.MustMigrateDB(t, db)
	repo := NewRepo(db)

	blocks := []UnlockedBlock{
		{
			AddrFrom: "abc",
			AddrTo:   "123",
			Balance:  decimal.NewFromFloat(2.3),
			TimeSpan: 2,
			Day:      civil.DateOf(time.Now()),
		},
		{
			AddrFrom: "abc2",
			AddrTo:   "1234",
			Balance:  decimal.NewFromFloat(2.4),
			TimeSpan: 3,
			Day:      civil.DateOf(time.Now()),
		},
	}
	err := repo.InsertUnlockedBlocks(blocks)
	require.NoError(t, err)
}
