package trace

import (
	"context"
	"errors"
	"log"
	"sync"
	"testing"
	"time"

	"github.com/SkyAPM/go2sky"
	"github.com/SkyAPM/go2sky/reporter"
)

func MockExtractor() (c string, e error) {
	return
}

func MockInjector(string) (e error) {
	return
}

func CrossChannelInjector(string) (e error) {
	return
}

func MockInjector2(string) (e error) {
	return errors.New("MockInjector2 error")
}

func Sleep(sec int64) {
	time.Sleep(time.Second * time.Duration(sec))
}

func TestSkyWalking001(t *testing.T) {
	r, err := reporter.NewGRPCReporter("127.0.0.1:11800")
	if err != nil {
		log.Fatalf("new reporter error %v \n", err)
	}
	defer r.Close()
	tracer, err := go2sky.NewTracer("example", go2sky.WithReporter(r))
	// create with sampler
	// tracer, err := go2sky.NewTracer("example", go2sky.WithReporter(r), go2sky.WithSampler(0.5))

	ctx := context.Background()
	span, ctx, _ := tracer.CreateEntrySpan(ctx, "entry", MockExtractor)
	subSpan, ctx, err := tracer.CreateLocalSpan(ctx)
	eSpan, _ := tracer.CreateExitSpan(ctx, "exit", "localhost:8080", MockInjector)
	// eSpan2, _ := tracer.CreateExitSpan(ctx, "exit2", "localhost:8081", MockInjector2)
	// eSpan2.End()
	eSpan.End()
	subSpan.End()
	span.End()
	for true {
		log.Printf("sleeping...")
		time.Sleep(time.Duration(1) * time.Second)
	}
	log.Printf("finish")
}

func TestSkyWalking002(t *testing.T) {
	r, err := reporter.NewGRPCReporter("127.0.0.1:11800")
	if err != nil {
		log.Fatalf("new reporter error %v \n", err)
	}
	defer r.Close()

	func() {
		tracer, _ := go2sky.NewTracer("example", go2sky.WithReporter(r))

		ctx := context.Background()
		span, ctx, _ := tracer.CreateEntrySpan(ctx, "entry", MockExtractor)
		defer span.End()

		exitWg := &sync.WaitGroup{}
		exitWg.Add(3)

		go func() {
			subSpan, ctx, _ := tracer.CreateLocalSpan(ctx)
			subSpan.SetOperationName("local_process_1")
			defer subSpan.End()
			Sleep(1)
			eSpan, _ := tracer.CreateExitSpan(ctx, "call_database", "database:6437", MockInjector)
			defer eSpan.End()
			Sleep(1)
			log.Printf("call_database finish")
			exitWg.Done()
		}()
		go func() {
			subSpan, ctx, _ := tracer.CreateLocalSpan(ctx)
			subSpan.SetOperationName("local_process_2")
			defer subSpan.End()
			Sleep(1)
			eSpan, _ := tracer.CreateExitSpan(ctx, "send_mq_type_k", "k:9092", MockInjector)
			defer eSpan.End()
			Sleep(1)
			log.Printf("send_mq_type_k finish")
			exitWg.Done()
		}()
		go func() {
			subSpan, ctx, _ := tracer.CreateLocalSpan(ctx)
			subSpan.SetOperationName("local_process_3")
			defer subSpan.End()
			Sleep(1)
			eSpan, _ := tracer.CreateExitSpan(ctx, "call_other_service", "other_service:8080", MockInjector)
			defer eSpan.End()
			Sleep(1)
			log.Printf("call_other_service finish")
			exitWg.Done()
		}()
		exitWg.Wait()
		Sleep(1)
	}()

	for true {
		log.Printf("sleeping...")
		time.Sleep(time.Duration(1) * time.Second)
	}
	log.Printf("finish")
}

func TestSkyWalking003(t *testing.T) {
	r, err := reporter.NewGRPCReporter("127.0.0.1:11800")
	if err != nil {
		log.Fatalf("new reporter error %v \n", err)
	}
	defer r.Close()

	func() {
		tracer, _ := go2sky.NewTracer("example", go2sky.WithReporter(r))

		ctx := context.Background()
		span, ctx, _ := tracer.CreateEntrySpan(ctx, "entry", MockExtractor)
		defer span.End()

		go func() {
			subSpan, ctx, _ := tracer.CreateEntrySpan(ctx, "local_process_1", MockExtractor)
			defer subSpan.End()
			Sleep(1)
			eSpan, _ := tracer.CreateExitSpan(ctx, "call_database", "database:6437", MockInjector)
			defer eSpan.End()
			Sleep(1)
			log.Printf("call_database finish")
		}()
		go func() {
			subSpan, ctx, _ := tracer.CreateEntrySpan(ctx, "local_process_2", MockExtractor)
			defer subSpan.End()
			Sleep(1)
			eSpan, _ := tracer.CreateExitSpan(ctx, "send_mq_type_k", "k:9092", MockInjector)
			defer eSpan.End()
			Sleep(1)
			log.Printf("send_mq_type_k finish")
		}()
		go func() {
			subSpan, ctx, _ := tracer.CreateEntrySpan(ctx, "local_process_3", MockExtractor)
			defer subSpan.End()
			Sleep(1)
			eSpan, _ := tracer.CreateExitSpan(ctx, "call_other_service", "other_service:8080", MockInjector)
			defer eSpan.End()
			Sleep(1)
			log.Printf("call_other_service finish")
		}()
		eSpan, _ := tracer.CreateExitSpan(ctx, "exit", "localhost", MockInjector)
		defer eSpan.End()
		Sleep(1)
	}()

	for true {
		log.Printf("sleeping...")
		time.Sleep(time.Duration(1) * time.Second)
	}
	log.Printf("finish")
}

func TestSkyWalking004(t *testing.T) {
	r, err := reporter.NewGRPCReporter("127.0.0.1:11800")
	if err != nil {
		log.Fatalf("new reporter error %v \n", err)
	}
	defer r.Close()

	func() {
		tracer, _ := go2sky.NewTracer("example5", go2sky.WithReporter(r))

		ctx := context.Background()
		span, ctx, _ := tracer.CreateEntrySpan(ctx, "TestSkyWalking004", MockExtractor)
		defer span.End()

		subSpan, ctx, _ := tracer.CreateLocalSpan(ctx)
		subSpan.SetOperationName("local_process_0")
		defer subSpan.End()

		var exitHeader string
		eSpan, _ := tracer.CreateExitSpan(ctx, "goroutines", "in_process", func(header string) error {
			log.Printf("header[%s]", header)
			exitHeader = header
			return nil
		})
		defer eSpan.End()

		go func() {
			subSpan, ctx, _ := tracer.CreateEntrySpan(context.Background(), "local_process_1", func() (s string, e error) {
				return exitHeader, nil
			})
			defer subSpan.End()
			Sleep(1)
			eSpan, _ := tracer.CreateExitSpan(ctx, "call_database", "database:6437", MockInjector)
			defer eSpan.End()
			Sleep(1)
			log.Printf("call_database finish")
		}()
		go func() {
			subSpan, ctx, _ := tracer.CreateEntrySpan(context.Background(), "local_process_2", func() (s string, e error) {
				return exitHeader, nil
			})
			defer subSpan.End()
			Sleep(1)
			eSpan, _ := tracer.CreateExitSpan(ctx, "send_mq_type_k", "k:9092", MockInjector)
			defer eSpan.End()
			Sleep(1)
			log.Printf("send_mq_type_k finish")
		}()
		go func() {
			subSpan, ctx, _ := tracer.CreateEntrySpan(context.Background(), "local_process_3", func() (s string, e error) {
				return exitHeader, nil
			})
			defer subSpan.End()
			Sleep(1)
			eSpan, _ := tracer.CreateExitSpan(ctx, "call_other_service", "other_service:8080", MockInjector)
			defer eSpan.End()
			Sleep(1)
			log.Printf("call_other_service finish")
		}()

		Sleep(1)
	}()

	for true {
		log.Printf("sleeping...")
		time.Sleep(time.Duration(1) * time.Second)
	}
	log.Printf("finish")
}

func TestSkyWalking005(t *testing.T) {
	r, err := reporter.NewGRPCReporter("127.0.0.1:11800")
	if err != nil {
		log.Fatalf("new reporter error %v \n", err)
	}
	defer r.Close()

	func() {
		tracer, _ := go2sky.NewTracer("k", go2sky.WithReporter(r))

		ctx := context.Background()
		span, ctx, _ := tracer.CreateEntrySpan(ctx, "TestSkyWalking005", MockExtractor)
		defer span.End()

		subSpan, ctx, _ := tracer.CreateLocalSpan(ctx)
		subSpan.SetOperationName("local_process_0")
		defer subSpan.End()

		var exitHeader string
		eSpan, _ := tracer.CreateExitSpan(ctx, "goroutines", "in_process", func(header string) error {
			log.Printf("header[%s]", header)
			exitHeader = header
			return nil
		})
		defer eSpan.End()

		go func() {
			subSpan, ctx, _ := tracer.CreateEntrySpan(context.Background(), "local_process_1", func() (s string, e error) {
				return exitHeader, nil
			})
			defer subSpan.End()
			Sleep(1)
			eSpan, _ := tracer.CreateExitSpan(ctx, "call_database", "database2:6437", MockInjector)
			defer eSpan.End()
			Sleep(1)
			log.Printf("call_database finish")
		}()
		go func() {
			subSpan, ctx, _ := tracer.CreateEntrySpan(context.Background(), "local_process_2", func() (s string, e error) {
				return exitHeader, nil
			})
			defer subSpan.End()
			Sleep(1)
			eSpan, _ := tracer.CreateExitSpan(ctx, "send_mq_type_k", "kafka2:9092", MockInjector)
			defer eSpan.End()
			Sleep(1)
			log.Printf("send_mq_type_k finish")
		}()
		go func() {
			subSpan, ctx, _ := tracer.CreateEntrySpan(context.Background(), "local_process_3", func() (s string, e error) {
				return exitHeader, nil
			})
			defer subSpan.End()
			Sleep(1)
			eSpan, _ := tracer.CreateExitSpan(ctx, "call_other_service", "other_service2:8080", MockInjector)
			defer eSpan.End()
			Sleep(1)
			log.Printf("call_other_service finish")
		}()

		Sleep(1)
	}()

	for true {
		log.Printf("sleeping...")
		time.Sleep(time.Duration(1) * time.Second)
	}
	log.Printf("finish")
}

func TestSkyWalking006(t *testing.T) {
	r, err := reporter.NewGRPCReporter("127.0.0.1:11800")
	if err != nil {
		log.Fatalf("new reporter error %v \n", err)
	}
	defer r.Close()

	func() {
		tracer, _ := go2sky.NewTracer("TestSkyWalking006", go2sky.WithReporter(r))

		ctx := context.Background()
		span, ctx, _ := tracer.CreateEntrySpan(ctx, "start", MockExtractor)
		defer span.End()

		go func() {
			subSpan, ctx, _ := tracer.CreateEntrySpan(ctx, "local_process_1", MockExtractor)
			defer subSpan.End()
			Sleep(1)
			eSpan, _ := tracer.CreateExitSpan(ctx, "call_database", "database:6437", MockInjector)
			defer eSpan.End()
			Sleep(1)
			log.Printf("call_database finish")
		}()
		go func() {
			subSpan, ctx, _ := tracer.CreateEntrySpan(ctx, "local_process_1", MockExtractor)
			defer subSpan.End()
			Sleep(1)
			eSpan, _ := tracer.CreateExitSpan(ctx, "call_database", "database:6437", MockInjector)
			defer eSpan.End()
			Sleep(1)
			log.Printf("send_mq_type_k finish")
		}()
		go func() {
			subSpan, ctx, _ := tracer.CreateEntrySpan(ctx, "local_process_1", MockExtractor)
			defer subSpan.End()
			Sleep(1)
			eSpan, _ := tracer.CreateExitSpan(ctx, "call_database", "database:6437", MockInjector)
			defer eSpan.End()
			Sleep(1)
			log.Printf("call_other_service finish")
		}()

		Sleep(1)
	}()

	for true {
		log.Printf("sleeping...")
		time.Sleep(time.Duration(1) * time.Second)
	}
	log.Printf("finish")
}

func TestSkyWalking007(t *testing.T) {
	r, err := reporter.NewGRPCReporter("127.0.0.1:11800")
	if err != nil {
		log.Fatalf("new reporter error %v \n", err)
	}
	defer r.Close()

	func() {
		tracer, _ := go2sky.NewTracer("example5", go2sky.WithReporter(r))

		ctx := context.Background()
		span, ctx, _ := tracer.CreateEntrySpan(ctx, "TestSkyWalking007", MockExtractor)
		defer span.End()

		subSpan, ctx, _ := tracer.CreateLocalSpan(ctx)
		subSpan.SetOperationName("local_process_0")
		defer subSpan.End()

		go func() {
			subSpan, ctx, _ := tracer.CreateEntrySpan(context.Background(), "local_process_1", MockExtractor)
			defer subSpan.End()
			Sleep(1)
			eSpan, _ := tracer.CreateExitSpan(ctx, "call_database", "database:6437", MockInjector)
			defer eSpan.End()
			Sleep(1)
			log.Printf("call_database finish")
		}()
		go func() {
			subSpan, ctx, _ := tracer.CreateEntrySpan(context.Background(), "local_process_2", MockExtractor)
			defer subSpan.End()
			Sleep(1)
			eSpan, _ := tracer.CreateExitSpan(ctx, "send_mq_type_k", "k:9092", MockInjector)
			defer eSpan.End()
			Sleep(1)
			log.Printf("send_mq_type_k finish")
		}()
		go func() {
			subSpan, ctx, _ := tracer.CreateEntrySpan(context.Background(), "local_process_3", MockExtractor)
			defer subSpan.End()
			Sleep(1)
			eSpan, _ := tracer.CreateExitSpan(ctx, "call_other_service", "other_service:8080", MockInjector)
			defer eSpan.End()
			Sleep(1)
			log.Printf("call_other_service finish")
		}()

		Sleep(1)
	}()

	for true {
		log.Printf("sleeping...")
		time.Sleep(time.Duration(1) * time.Second)
	}
	log.Printf("finish")
}

func TestSkyWalking008(t *testing.T) {
	r, err := reporter.NewGRPCReporter("127.0.0.1:11800")
	if err != nil {
		log.Fatalf("new reporter error %v \n", err)
	}
	defer r.Close()

	func() {
		tracer, _ := go2sky.NewTracer("example5", go2sky.WithReporter(r))

		ctx := context.Background()
		span, ctx, _ := tracer.CreateEntrySpan(ctx, "TestSkyWalking008", MockExtractor)
		defer span.End()

		go func() {
			subSpan, ctx, _ := tracer.CreateEntrySpan(ctx, "local_process_1", MockExtractor)
			defer subSpan.End()
			Sleep(1)
			eSpan, _ := tracer.CreateExitSpan(ctx, "call_database", "database:6437", MockInjector)
			defer eSpan.End()
			Sleep(1)
			log.Printf("call_database finish")
		}()
		go func() {
			subSpan, ctx, _ := tracer.CreateEntrySpan(ctx, "local_process_2", MockExtractor)
			defer subSpan.End()
			Sleep(1)
			eSpan, _ := tracer.CreateExitSpan(ctx, "send_mq_type_k", "k:9092", MockInjector)
			defer eSpan.End()
			Sleep(1)
			log.Printf("send_mq_type_k finish")
		}()
		go func() {
			subSpan, ctx, _ := tracer.CreateEntrySpan(ctx, "local_process_3", MockExtractor)
			defer subSpan.End()
			Sleep(1)
			eSpan, _ := tracer.CreateExitSpan(ctx, "call_other_service", "other_service:8080", MockInjector)
			defer eSpan.End()
			Sleep(1)
			log.Printf("call_other_service finish")
		}()

		Sleep(1)
	}()

	for true {
		log.Printf("sleeping...")
		time.Sleep(time.Duration(1) * time.Second)
	}
	log.Printf("finish")
}

func TestSkyWalking009(t *testing.T) {
	r, err := reporter.NewGRPCReporter("127.0.0.1:11800")
	if err != nil {
		log.Fatalf("new reporter error %v \n", err)
	}
	defer r.Close()

	func() {
		tracer, _ := go2sky.NewTracer("example5", go2sky.WithReporter(r))

		ctx := context.Background()
		span, ctx, _ := tracer.CreateEntrySpan(ctx, "TestSkyWalking009", MockExtractor)
		defer span.End()

		go func() {
			eSpan, _ := tracer.CreateExitSpan(ctx, "call_database", "database:6437", MockInjector)
			defer eSpan.End()
			Sleep(1)
			log.Printf("call_database finish")
		}()
		go func() {
			eSpan, _ := tracer.CreateExitSpan(ctx, "send_mq_type_k", "k:9092", MockInjector)
			defer eSpan.End()
			Sleep(1)
			log.Printf("send_mq_type_k finish")
		}()
		go func() {
			eSpan, _ := tracer.CreateExitSpan(ctx, "call_other_service", "other_service:8080", MockInjector)
			defer eSpan.End()
			Sleep(1)
			log.Printf("call_other_service finish")
		}()

		Sleep(1)
	}()

	for true {
		log.Printf("sleeping...")
		time.Sleep(time.Duration(1) * time.Second)
	}
	log.Printf("finish")
}

func TestSkyWalking010(t *testing.T) {
	r, err := reporter.NewGRPCReporter("127.0.0.1:11800")
	if err != nil {
		log.Fatalf("new reporter error %v \n", err)
	}
	defer r.Close()

	func() {
		tracer, _ := go2sky.NewTracer("example5", go2sky.WithReporter(r))

		ctx := context.Background()
		span, ctx, _ := tracer.CreateEntrySpan(ctx, "TestSkyWalking010", MockExtractor)
		defer span.End()

		eSpan, _ := tracer.CreateExitSpan(ctx, "call_database", "database:6437", MockInjector)
		defer eSpan.End()
		Sleep(1)
		log.Printf("call_database finish")
		eSpan, _ = tracer.CreateExitSpan(ctx, "send_mq_type_k", "k:9092", MockInjector)
		defer eSpan.End()
		Sleep(1)
		log.Printf("send_mq_type_k finish")
		eSpan, _ = tracer.CreateExitSpan(ctx, "call_other_service", "other_service:8080", MockInjector)
		defer eSpan.End()
		Sleep(1)
		log.Printf("call_other_service finish")

		Sleep(1)
	}()

	for true {
		log.Printf("sleeping...")
		time.Sleep(time.Duration(1) * time.Second)
	}
	log.Printf("finish")
}

func TestSkyWalking011(t *testing.T) {
	r, err := reporter.NewGRPCReporter("127.0.0.1:11800")
	if err != nil {
		log.Fatalf("new reporter error %v \n", err)
	}
	defer r.Close()

	func() {
		tracer, _ := go2sky.NewTracer("example5", go2sky.WithReporter(r))

		ctx := context.Background()
		span, ctx, _ := tracer.CreateEntrySpan(ctx, "TestSkyWalking011", MockExtractor)
		defer span.End()

		eSpan, _ := tracer.CreateExitSpan(ctx, "call_database", "database:6437", MockInjector)
		defer eSpan.End()
		Sleep(1)
		log.Printf("call_database finish")
		eSpan, _ = tracer.CreateExitSpan(ctx, "call_database", "database:6437", MockInjector)
		defer eSpan.End()
		Sleep(1)
		log.Printf("call_database finish")
		eSpan, _ = tracer.CreateExitSpan(ctx, "call_database", "database:6437", MockInjector)
		defer eSpan.End()
		Sleep(1)
		log.Printf("call_database finish")

		Sleep(1)
	}()

	for true {
		log.Printf("sleeping...")
		time.Sleep(time.Duration(1) * time.Second)
	}
	log.Printf("finish")
}

// func TestSkyWalking012(t *testing.T) {
//	skyWalkingUrl := "127.0.0.1:11800"
//	serverName := "TestSkyWalking012"
//	initSkyWalking(skyWalkingUrl, serverName, 10, true)
//
//	func() {
//		ctx := context.Background()
//		span, ctx, _ := skyTracer().CreateEntrySpan(ctx, "TestSkyWalking012", MockExtractor)
//		defer span.End()
//
//		eSpan, _ := skyTracer().CreateExitSpan(ctx, "call_database", "database:6437", MockInjector)
//		defer eSpan.End()
//		Sleep(1)
//		log.Printf("call_database finish")
//		eSpan, _ = skyTracer().CreateExitSpan(ctx, "send_mq_type_k", "k:9092", MockInjector)
//		defer eSpan.End()
//		Sleep(1)
//		log.Printf("send_mq_type_k finish")
//		eSpan, _ = skyTracer().CreateExitSpan(ctx, "call_other_service", "other_service:8080", MockInjector)
//		defer eSpan.End()
//		Sleep(1)
//		log.Printf("call_other_service finish")
//
//		Sleep(1)
//	}()
//
//	for true {
//		log.Printf("sleeping...")
//		time.Sleep(time.Duration(1) * time.Second)
//	}
//	log.Printf("finish")
// }
