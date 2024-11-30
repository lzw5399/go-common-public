/**
APM is short for Application Performance Manage
this file implements SkyWalking APM
*/
package trace

import (
    "github.com/SkyAPM/go2sky"
)

type noopReporter struct {
}

func (r *noopReporter) Boot(service string, serviceInstance string) {
}

func (r *noopReporter) Send(spans []go2sky.ReportedSpan) {
}

func (r *noopReporter) Close() {
}

func createTracer(skyWalkingUrl string, serverName string, samplePartitions uint32,
    skyWalkingEnable bool, reporter go2sky.Reporter) (*go2sky.Tracer, go2sky.Reporter) {
    if skyWalkingEnable {
        if reporter == nil {
            rp, err := NewGRPCReporter(skyWalkingUrl, WithSamplePartitions(samplePartitions))
            if err != nil {
                panic(err)
            }
            reporter = rp
        }
    } else {
        reporter = &noopReporter{}
    }
    tr, err := go2sky.NewTracer(serverName, go2sky.WithReporter(reporter))
    if err != nil {
        panic(err)
    }
    return tr, reporter
}
