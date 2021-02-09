package infra

import (
	"fmt"
	"os"
	"runtime"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	r "github.com/stretchr/testify/require"

	_ "github.com/lib/pq"
)

// eg:  5432;unittest;unittest;pwd
// sql: create user unittest with password 'pwd';create database unittest with owner unittest;
// alter user unittest with createdb;
const TestEnvLocalDB = "DEV_DB"
const TestEnvLocalDBKeep = "DEV_DB_KEEP"

// 创建一个用于测试的数据库，测试清理时删除数据库
func MustNewTestPGDB(t *testing.T) *sqlx.DB {
	t.Log("create test database using local env:", os.Getenv(TestEnvLocalDB))
	seedInfo := strings.Split(os.Getenv(TestEnvLocalDB), ";") //port;database;user;password
	r.Len(t, seedInfo, 4)
	port, err := strconv.Atoi(seedInfo[0])
	r.NoError(t, err)
	now := time.Now()

	caller := strings.Replace(getFrame(1).Function, "Test", "", 1)
	caller = strings.Replace(caller, "bbcsyncer/", "", -1)
	caller = strings.Replace(caller, "/", "_", -1)
	caller = strings.Replace(caller, ".", "_", -1)
	caller = strings.ToLower(caller)
	dbName := fmt.Sprintf("ut_%s_%s", caller, now.Format("20060102_150405"))
	user, password := seedInfo[2], seedInfo[3]

	db, err := sqlx.Connect("postgres", fmt.Sprintf("postgres://%s:%s@localhost:%d/%s?sslmode=disable", user, password, port, seedInfo[1]))
	r.NoError(t, err)

	createDB := fmt.Sprintf("create database %s owner %s", dbName, user)
	t.Log("createDB:", createDB)
	_, err = db.Exec(createDB)
	r.NoError(t, err)

	newDB, err := sqlx.Connect("postgres", fmt.Sprintf("postgres://%s:%s@localhost:%d/%s?sslmode=disable", user, password, port, dbName))
	r.NoError(t, err)
	setDBMapper(newDB)
	t.Cleanup(func() {
		defer db.Close()
		if os.Getenv(TestEnvLocalDBKeep) != "keep" {
			_ = newDB.Close()
			if _, e := db.Exec(fmt.Sprintf("drop database if exists %s", dbName)); e != nil {
				t.Log("[ERR] drop test db failed", e)
			}
		}
	})
	return newDB
}

func MustMigrateDB(t *testing.T, db *sqlx.DB) {
	for _, sql := range strings.Split(schemaSQL, ";") {
		if strings.HasPrefix(strings.TrimSpace(sql), "--") {
			continue
		}
		// t.Log("migrate sql:", sql)
		_, err := db.Exec(sql)
		r.NoError(t, err)
	}
}

func getFrame(skipFrames int) runtime.Frame {
	// We need the frame at index skipFrames+2, since we never want runtime.Callers and getFrame
	targetFrameIndex := skipFrames + 2

	// Set size to targetFrameIndex+2 to ensure we have room for one more caller than we need
	programCounters := make([]uintptr, targetFrameIndex+2)
	n := runtime.Callers(0, programCounters)

	frame := runtime.Frame{Function: "unknown"}
	if n > 0 {
		frames := runtime.CallersFrames(programCounters[:n])
		for more, frameIndex := true, 0; more && frameIndex <= targetFrameIndex; frameIndex++ {
			var frameCandidate runtime.Frame
			frameCandidate, more = frames.Next()
			if frameIndex == targetFrameIndex {
				frame = frameCandidate
			}
		}
	}
	return frame
}
