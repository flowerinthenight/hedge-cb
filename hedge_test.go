package hedge

import (
	"context"
	"log/slog"
	"testing"
	"time"

	"cloud.google.com/go/spanner"
	gaxv2 "github.com/googleapis/gax-go/v2"
)

const (
	db = "projects/test-project/instances/test-instance/databases/testdb"
)

func TestAny(t *testing.T) {
	// TODO:
	d := []byte{}
	d = append(d, []byte("hello")...)
	d = append(d, []byte("world")...)
	d = append(d, []byte("stats")...)
	d = append(d, []byte("one")...)
	slog.Info("next:", "val", d[0:5])
	slog.Info("next:", "val", d[5:10])
	slog.Info("next:", "val", d[10:15])
	slog.Info("next:", "val", d[15:20])
}

func TestBasic(t *testing.T) {
	ctx := context.Background()
	client, err := spanner.NewClient(ctx, db)
	if err != nil {
		t.Error(err)
		return
	}

	defer client.Close()
	op := New(client, ":8080", "locktable", "mylock", "logtable",
		WithLeaderHandler(
			nil,
			func(data interface{}, msg []byte) ([]byte, error) {
				t.Log("[send] received:", string(msg))
				return []byte("send " + string(msg)), nil
			},
		),
		WithBroadcastHandler(
			nil,
			func(data interface{}, msg []byte) ([]byte, error) {
				t.Log("[broadcast/semaphore] received:", string(msg))
				return nil, nil
			},
		),
	)

	done := make(chan error, 1)
	quit, cancel := context.WithCancel(ctx)
	go op.Run(quit, done)

	var cnt int
	bo := gaxv2.Backoff{
		Initial:    time.Second,
		Max:        time.Second * 30,
		Multiplier: 2,
	}

	for {
		cnt++
		locked, _ := op.HasLock()
		switch {
		case locked:
			t.Log("got lock")
			break
		default:
			t.Log("didn't get lock, retry")
			time.Sleep(bo.Pause())
			continue
		}

		if cnt >= 10 {
			t.Fatalf("can't get lock")
		}

		break
	}

	cancel()
	<-done
}
