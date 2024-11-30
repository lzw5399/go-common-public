package trace

import (
	"context"
	"fmt"
	"hash/fnv"
	"log"
	"testing"
	"time"

	"gopkg.in/go-playground/assert.v1"
)

func TestClient01(t *testing.T) {
	build := CreateBuild("127.0.0.1:11800", "TestClient01", 1, true)
	BuildApmClient(build)
	defer CloseApmClient()

	func() {
		ctx := context.Background()
		span, ctx := ApmClient().CreateEntrySpan(ctx, "TestClient01", MockExtractor)
		defer span.End()

		grpcESpan, ctx := ApmClient().CreateGRpcExitSpan(ctx, "grpc_exit", "method", "thirdPart")
		defer grpcESpan.End()
		Sleep(1)
		log.Printf("grpc finish")

		esESpan := ApmClient().CreateESExitSpan(ctx, "es_exit", "es_address")
		defer esESpan.End()
		Sleep(1)
		log.Printf("es finish")

		kEsSpan, ctx := ApmClient().CreateKEntrySpan(ctx, "k_exit", "method", nil)
		defer kEsSpan.End()
		Sleep(1)
		log.Printf("k finish")

		mongoSpan := ApmClient().CreateMongoExitSpan(ctx, "mongo_exit", "address", "dbname")
		defer mongoSpan.End()
		Sleep(1)
		log.Printf("mongo finish")

		Sleep(1)
	}()

	for true {
		log.Printf("sleeping...")
		time.Sleep(time.Duration(1) * time.Second)
	}
	log.Printf("finish")
}

func TestClient_NoTraceContext(t *testing.T) {
	build := CreateBuild("127.0.0.1:11800", "TestClient01", 1, true)
	BuildApmClient(build)
	defer CloseApmClient()

	ctx := ApmClient().NoTraceContext(nil)
	assert.Equal(t, true, ApmClient().isNoTraceContext(nil))
	assert.Equal(t, true, ApmClient().isNoTraceContext(ctx))

	ctx = context.Background()
	assert.Equal(t, false, ApmClient().isNoTraceContext(ctx))

	ctx = ApmClient().NoTraceContext(ctx)
	assert.Equal(t, true, ApmClient().isNoTraceContext(ctx))

}

func TestClient_Enable(t *testing.T) {
	func() {
		build := CreateBuild("127.0.0.1:11800", "TestClient_Enable", 1, false)
		BuildApmClient(build)
		defer CloseApmClient()

		ctx := context.Background()
		span := ApmClient().CreateExitSpan(ctx, "TestClient_Enable", "TestClient_Enable", MockInjector)
		assert.Equal(t, true, IsNSpan(span))

		span, _ = ApmClient().CreateLocalSpan(ctx)
		assert.Equal(t, true, IsNSpan(span))

		span, _ = ApmClient().CreateEntrySpan(ctx, "TestClient_Enable", MockExtractor)
		assert.Equal(t, true, IsNSpan(span))
	}()

	func() {
		build := CreateBuild("127.0.0.1:11800", "TestClient_Enable", 1, true)
		BuildApmClient(build)
		defer CloseApmClient()

		ctx := context.Background()
		span := ApmClient().CreateExitSpan(ctx, "TestClient_Enable", "TestClient_Enable", MockInjector)
		assert.Equal(t, false, IsNSpan(span))

		span, _ = ApmClient().CreateLocalSpan(ctx)
		assert.Equal(t, false, IsNSpan(span))

		span, _ = ApmClient().CreateEntrySpan(ctx, "TestClient_Enable", MockExtractor)
		assert.Equal(t, false, IsNSpan(span))
	}()

}

func calcStringHashCode(str string) uint32 {
	hash := fnv.New32a()
	_, _ = hash.Write([]byte(str))
	return hash.Sum32()
}

func TestClient_calcStringHashCode(t *testing.T) {
	samples := 100000
	count := 0
	results := make([]uint32, 0, samples)
	start := float64(time.Now().UnixNano()) / 1e9
	for i := 0; i < samples; i++ {
		str := fmt.Sprintf("%d", i)
		res := calcStringHashCode(str) % 10
		if res == 0 {
			count += 1
		}
		// log.Printf("str[%s], res[%d]", str, res)
		results = append(results, res)
	}
	end := float64(time.Now().UnixNano()) / 1e9
	log.Printf("samples[%d], time[%f], rate[%f], count[%d], ratio[%f]",
		samples, end-start, float64(samples)/(end-start), count, float64(count)/float64(samples))
}

func TestClient_sprintf(t *testing.T) {
	samples := 100000
	results := make([]string, 0, samples)
	start := float64(time.Now().UnixNano()) / 1e9
	for i := 0; i < samples; i++ {
		str := fmt.Sprintf("adkfjaldjfalkdjf;lakdjf;lakdjflkadjflkjadfkja[%d], adkfjaldjfalkdjf;lakdjf;lakdjflkadjflkjadfkja[%s]", i, "adkfjaldjfalkdjf;lakdjf;lakdjflkadjflkjadfkja")
		results = append(results, str)
	}
	end := float64(time.Now().UnixNano()) / 1e9
	log.Printf("samples[%d], time[%f], rate[%f]",
		samples, end-start, float64(samples)/(end-start))
}

func TestClient_ParseURL_perform(t *testing.T) {
	samples := 100000
	results := make([]string, 0, samples)
	start := float64(time.Now().UnixNano()) / 1e9
	for i := 0; i < samples; i++ {
		str := "https://www.google.com/search?q=golang+print+string+10+times&oq=golang+print+string+10+times&aqs=chrome..69i57.8786j0j8&sourceid=chrome&ie=UTF-8"
		_, host, _ := ParseURL(str)
		results = append(results, host)
	}
	end := float64(time.Now().UnixNano()) / 1e9
	log.Printf("samples[%d], time[%f], rate[%f]",
		samples, end-start, float64(samples)/(end-start))
}

func TestClient_ParseURL(t *testing.T) {
	var links = []string{"https://analytics.google.com/analytics/web/#embed/report-home/a98705171w145119383p149829595/",
		"jdbc:mysql://test_user:ouupppssss@localhost:3306/sakila?profileSQL=true",
		"https://bob:pass@testing.com/country/state",
		"http://www.golangprograms.com/",
		"mailto:John.Mark@testing.com",
		"https://www.google.com/search?q=golang+print+string+10+times&oq=golang+print+string+10+times&aqs=chrome..69i57.8786j0j8&sourceid=chrome&ie=UTF-8",
		"urn:oasis:names:description:docbook:dtd:xml:4.1.2",
		"https://stackoverflow.com/jobs?med=site-ui&ref=jobs-tab",
		"ssh://mark@testing.com",
		"postgres://user:pass@host.com:5432/path?k=v#f",
	}
	for _, link := range links {
		scheme, host, path := ParseURL(link)
		log.Printf("scheme[%s], host[%s], path[%s]", scheme, host, path)
	}
}

func TestClient_ParseURLWithCache(t *testing.T) {
	var links = []string{"https://analytics.google.com/analytics/web/#embed/report-home/a98705171w145119383p149829595/",
		"jdbc:mysql://test_user:ouupppssss@localhost:3306/sakila?profileSQL=true",
		"https://bob:pass@testing.com/country/state",
		"http://www.golangprograms.com/",
		"mailto:John.Mark@testing.com",
		"https://www.google.com/search?q=golang+print+string+10+times&oq=golang+print+string+10+times&aqs=chrome..69i57.8786j0j8&sourceid=chrome&ie=UTF-8",
		"urn:oasis:names:description:docbook:dtd:xml:4.1.2",
		"https://stackoverflow.com/jobs?med=site-ui&ref=jobs-tab",
		"ssh://mark@testing.com",
		"postgres://user:pass@host.com:5432/path?k=v#f",
		"redis://proxy.redis-cluster:6379/6",
		"http://efk-elasticsearch.efk:9200",
	}
	for _, link := range links {
		scheme, host, path := ParseURLWithCache(link)
		log.Printf("scheme[%s], host[%s], path[%s]", scheme, host, path)
	}
}

func TestClient_ParseMongoURLWithCache(t *testing.T) {
	var links = []string{
		"mongo-cluster.storage.svc.cluster.local",
		"mongodb+srv://admin:password@prefix.mongodb.net:27017/dbname",
		"mongodb://2E45fFg:aPgjb54!cyDOU$y41QYX@mongo-cluster.storage:27017/mop-applet-build/admin?replicaSet=rs0&authSource=admin",
	}
	for _, link := range links {
		scheme, host, path := ParseMongoURLWithCache(link)
		log.Printf("scheme[%s], host[%s], path[%s]", scheme, host, path)
	}
	for _, link := range links {
		scheme, host, path := ParseURLWithCache(link)
		log.Printf("scheme[%s], host[%s], path[%s]", scheme, host, path)
	}
}
