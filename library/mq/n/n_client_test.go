package n

import (
	"context"
	"strconv"
	"sync/atomic"
	"testing"
	"time"
)

func TestPubSub(t *testing.T) {
	url := "n://derek:T0pS3cr3t@127.0.0.1:4222"
	c := NewNClient(url)
	c.Start()
	ctx := context.Background()
	count := 0
	c.Sub(ctx, "test", func(ctx context.Context, input []byte) error {
		t.Logf("i got")
		count++
		return nil
	})
	c.Sub(ctx, "test", func(ctx context.Context, input []byte) error {
		count++
		t.Logf("i got")
		return nil
	})

	c.Pub(context.Background(), "test", []byte("hello"))
	time.Sleep(3 * time.Second)
	if count != 2 {
		t.Errorf("TestPubSub not pass")
	}
}

func TestPubQueueSub(t *testing.T) {
	url := "n://derek:T0pS3cr3t@127.0.0.1:4222"
	c := NewNClient(url)
	c.Start()
	ctx := context.Background()
	sameGroupCount := 0
	notSameGroupCount := 0
	c.QueueSub(ctx, "test", "g1", func(ctx context.Context, input []byte) error {
		sameGroupCount++
		notSameGroupCount++
		t.Logf("i got g1")
		return nil
	})
	c.QueueSub(ctx, "test", "g1", func(ctx context.Context, input []byte) error {
		sameGroupCount++
		notSameGroupCount++
		t.Logf("i got g1")
		return nil
	})
	c.QueueSub(ctx, "test", "g2", func(ctx context.Context, input []byte) error {
		notSameGroupCount++
		t.Logf("i got g2")
		return nil
	})
	c.Pub(context.Background(), "test", []byte("hello"))
	time.Sleep(3 * time.Second)
	if sameGroupCount != 1 || notSameGroupCount != 2 {
		t.Errorf("TestPubQueueSub not pass")
	}
}

func TestReconnect(t *testing.T) {
	url := "n://derek:T0pS3cr3t@127.0.0.1:4222"
	c := NewNClient(url)
	c.Start()
	ctx := context.Background()
	ch := make(chan int64)
	var count int64 = 0
	c.Sub(ctx, "test", func(ctx context.Context, input []byte) error {
		t.Logf("i got msg:%s", string(input))
		atomic.AddInt64(&count, 1)
		ch <- count
		return nil
	})
	c.Pub(context.Background(), "test", []byte("hello"+strconv.FormatInt(count, 10)))
	time.Sleep(10 * time.Second)
	<-ch
	c.Pub(context.Background(), "test", []byte("hello"+strconv.FormatInt(count, 10)))
	time.Sleep(10 * time.Second)
	<-ch
	c.Pub(context.Background(), "test", []byte("hello"+strconv.FormatInt(count, 10)))
	time.Sleep(10 * time.Second)
	<-ch
	c.Pub(context.Background(), "test", []byte("hello"+strconv.FormatInt(count, 10)))
	time.Sleep(10 * time.Second)
	<-ch
	if count != 4 {
		t.Errorf("TestReconnect not pass")
	}
}
