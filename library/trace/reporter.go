package trace

import (
    "context"
    "hash/fnv"
    "log"
    "os"
    "time"

    "github.com/SkyAPM/go2sky"
    "github.com/SkyAPM/go2sky/reporter/grpc/common"
    agentv3 "github.com/SkyAPM/go2sky/reporter/grpc/language-agent"
    managementv3 "github.com/SkyAPM/go2sky/reporter/grpc/management"
    "google.golang.org/grpc"
    "google.golang.org/grpc/connectivity"
    "google.golang.org/grpc/credentials"
    "google.golang.org/grpc/metadata"
)

const (
    maxSendQueueSize        int32  = 30000
    defaultCheckInterval           = 20 * time.Second
    defaultLogPrefix               = "go2sky-gRPC"
    authKey                        = "Authentication"
    defaultSamplePartitions uint32 = 1
)

// NewGRPCReporter create a new reporter to send data to gRPC oap server. Only one backend address is allowed.
func NewGRPCReporter(serverAddr string, opts ...GRPCReporterOption) (go2sky.Reporter, error) {
    r := &gRPCReporter{
        logger:           log.New(os.Stderr, defaultLogPrefix, log.LstdFlags),
        sendCh:           make(chan *agentv3.SegmentObject, maxSendQueueSize),
        checkInterval:    defaultCheckInterval,
        samplePartitions: defaultSamplePartitions,
    }
    for _, o := range opts {
        o(r)
    }

    var credsDialOption grpc.DialOption
    if r.creds != nil {
        // use tls
        credsDialOption = grpc.WithTransportCredentials(r.creds)
    } else {
        credsDialOption = grpc.WithInsecure()
    }

    conn, err := grpc.Dial(serverAddr, credsDialOption)
    if err != nil {
        return nil, err
    }
    r.conn = conn
    r.traceClient = agentv3.NewTraceSegmentReportServiceClient(r.conn)
    r.managementClient = managementv3.NewManagementServiceClient(r.conn)
    return r, nil
}

// GRPCReporterOption allows for functional options to adjust behaviour
// of a gRPC reporter to be created by NewGRPCReporter
type GRPCReporterOption func(r *gRPCReporter)

// WithLogger setup logger for gRPC reporter
func WithLogger(logger *log.Logger) GRPCReporterOption {
    return func(r *gRPCReporter) {
        r.logger = logger
    }
}

// WithCheckInterval setup service and endpoint registry check interval
func WithCheckInterval(interval time.Duration) GRPCReporterOption {
    return func(r *gRPCReporter) {
        r.checkInterval = interval
    }
}

// WithSamplePartitions setup sample partitions. For example,
// when partitions is 1, then the sample rate is 1;
// when partitions is 10, then the sample rate is 0.1;
// when partitions is 100, then the sample rate is 0.01;
func WithSamplePartitions(partitions uint32) GRPCReporterOption {
    return func(r *gRPCReporter) {
        if partitions <= 0 {
            partitions = 1
        }
        r.samplePartitions = partitions
    }
}

// WithMaxSendQueueSize setup send span queue buffer length
func WithMaxSendQueueSize(maxSendQueueSize int) GRPCReporterOption {
    return func(r *gRPCReporter) {
        r.sendCh = make(chan *agentv3.SegmentObject, maxSendQueueSize)
    }
}

// WithInstanceProps setup service instance properties eg: org=SkyAPM
func WithInstanceProps(props map[string]string) GRPCReporterOption {
    return func(r *gRPCReporter) {
        r.instanceProps = props
    }
}

// WithTransportCredentials setup transport layer security
func WithTransportCredentials(creds credentials.TransportCredentials) GRPCReporterOption {
    return func(r *gRPCReporter) {
        r.creds = creds
    }
}

// WithAuthentication used Authentication for gRPC
func WithAuthentication(auth string) GRPCReporterOption {
    return func(r *gRPCReporter) {
        r.md = metadata.New(map[string]string{authKey: auth})
    }
}

type gRPCReporter struct {
    service          string
    serviceInstance  string
    instanceProps    map[string]string
    logger           *log.Logger
    sendCh           chan *agentv3.SegmentObject
    conn             *grpc.ClientConn
    traceClient      agentv3.TraceSegmentReportServiceClient
    managementClient managementv3.ManagementServiceClient
    checkInterval    time.Duration
    samplePartitions uint32

    md    metadata.MD
    creds credentials.TransportCredentials
}

func (r *gRPCReporter) Boot(service string, serviceInstance string) {
    r.service = service
    r.serviceInstance = serviceInstance
    r.initSendPipeline()
    r.check()
}

func (r *gRPCReporter) calcStringHashCode(str string) uint32 {
    hash := fnv.New32a()
    _, _ = hash.Write([]byte(str))
    return hash.Sum32()
}

func (r *gRPCReporter) sampling(spans []go2sky.ReportedSpan) []go2sky.ReportedSpan {
    if spans == nil {
        return nil
    }
    samplePartitions := r.samplePartitions
    if samplePartitions <= 0 {
        samplePartitions = 1
    }
    sampled := make([]go2sky.ReportedSpan, 0, len(spans))
    for _, span := range spans {
        idx := r.calcStringHashCode(span.Context().TraceID) % samplePartitions
        if idx == 0 {
            sampled = append(sampled, span)
        }
    }
    return sampled
}

func (r *gRPCReporter) Send(spans []go2sky.ReportedSpan) {
    spans = r.sampling(spans)
    spanSize := len(spans)
    if spanSize < 1 {
        return
    }
    rootSpan := spans[spanSize-1]
    rootCtx := rootSpan.Context()
    segmentObject := &agentv3.SegmentObject{
        TraceId:         rootCtx.TraceID,
        TraceSegmentId:  rootCtx.SegmentID,
        Spans:           make([]*agentv3.SpanObject, spanSize),
        Service:         r.service,
        ServiceInstance: r.serviceInstance,
    }
    for i, s := range spans {
        spanCtx := s.Context()
        segmentObject.Spans[i] = &agentv3.SpanObject{
            SpanId:        spanCtx.SpanID,
            ParentSpanId:  spanCtx.ParentSpanID,
            StartTime:     s.StartTime(),
            EndTime:       s.EndTime(),
            OperationName: s.OperationName(),
            Peer:          s.Peer(),
            SpanType:      s.SpanType(),
            SpanLayer:     s.SpanLayer(),
            ComponentId:   s.ComponentID(),
            IsError:       s.IsError(),
            Tags:          s.Tags(),
            Logs:          s.Logs(),
        }
        srr := make([]*agentv3.SegmentReference, 0)
        if i == (spanSize-1) && spanCtx.ParentSpanID > -1 {
            srr = append(srr, &agentv3.SegmentReference{
                RefType:               agentv3.RefType_CrossThread,
                TraceId:               spanCtx.TraceID,
                ParentTraceSegmentId:  spanCtx.ParentSegmentID,
                ParentSpanId:          spanCtx.ParentSpanID,
                ParentService:         r.service,
                ParentServiceInstance: r.serviceInstance,
            })
        }
        if len(s.Refs()) > 0 {
            for _, tc := range s.Refs() {
                srr = append(srr, &agentv3.SegmentReference{
                    RefType:                  agentv3.RefType_CrossProcess,
                    TraceId:                  spanCtx.TraceID,
                    ParentTraceSegmentId:     tc.ParentSegmentID,
                    ParentSpanId:             tc.ParentSpanID,
                    ParentService:            tc.ParentService,
                    ParentServiceInstance:    tc.ParentServiceInstance,
                    ParentEndpoint:           tc.ParentEndpoint,
                    NetworkAddressUsedAtPeer: tc.AddressUsedAtClient,
                })
            }
        }
        segmentObject.Spans[i].Refs = srr
    }
    defer func() {
        // recover the panic caused by skyClose sendCh
        if err := recover(); err != nil {
            r.logger.Printf("reporter segment err %v", err)
        }
    }()
    select {
    case r.sendCh <- segmentObject:
    default:
        r.logger.Printf("reach max send buffer")
    }
}

func (r *gRPCReporter) Close() {
    if r.sendCh != nil {
        close(r.sendCh)
    }
    r.closeGRPCConn()
}

func (r *gRPCReporter) closeGRPCConn() {
    if r.conn != nil {
        if err := r.conn.Close(); err != nil {
            r.logger.Print(err)
        }
    }
}

func (r *gRPCReporter) initSendPipeline() {
    if r.traceClient == nil {
        return
    }
    go func() {
    StreamLoop:
        for {
            stream, err := r.traceClient.Collect(metadata.NewOutgoingContext(context.Background(), r.md))
            if err != nil {
                r.logger.Printf("open stream error %v", err)
                time.Sleep(5 * time.Second)
                continue StreamLoop
            }
            for s := range r.sendCh {
                err = stream.Send(s)
                if err != nil {
                    r.logger.Printf("send segment error %v", err)
                    r.closeStream(stream)
                    continue StreamLoop
                }
            }
            r.closeStream(stream)
            r.closeGRPCConn()
            break
        }
    }()
}

func (r *gRPCReporter) closeStream(stream agentv3.TraceSegmentReportService_CollectClient) {
    err := stream.CloseSend()
    if err != nil {
        r.logger.Printf("send closing error %v", err)
    }
}

func (r *gRPCReporter) reportInstanceProperties() (err error) {
    props := buildOSInfo()
    if r.instanceProps != nil {
        for k, v := range r.instanceProps {
            props = append(props, &common.KeyStringValuePair{
                Key:   k,
                Value: v,
            })
        }
    }
    _, err = r.managementClient.ReportInstanceProperties(metadata.NewOutgoingContext(context.Background(), r.md), &managementv3.InstanceProperties{
        Service:         r.service,
        ServiceInstance: r.serviceInstance,
        Properties:      props,
    })
    return err
}

func (r *gRPCReporter) check() {
    if r.checkInterval < 0 || r.conn == nil || r.managementClient == nil {
        return
    }
    go func() {
        instancePropertiesSubmitted := false
        for {
            if r.conn.GetState() == connectivity.Shutdown {
                break
            }

            if !instancePropertiesSubmitted {
                err := r.reportInstanceProperties()
                if err != nil {
                    r.logger.Printf("report serviceInstance properties error %v", err)
                    time.Sleep(r.checkInterval)
                    continue
                }
                instancePropertiesSubmitted = true
            }

            _, err := r.managementClient.KeepAlive(metadata.NewOutgoingContext(context.Background(), r.md), &managementv3.InstancePingPkg{
                Service:         r.service,
                ServiceInstance: r.serviceInstance,
            })

            if err != nil {
                r.logger.Printf("send keep alive signal error %v", err)
            }
            time.Sleep(r.checkInterval)
        }
    }()
}

func buildOSInfo() (props []*common.KeyStringValuePair) {
    processNo := ProcessNo()
    if processNo != "" {
        kv := &common.KeyStringValuePair{
            Key:   "Process No.",
            Value: processNo,
        }
        props = append(props, kv)
    }

    hostname := &common.KeyStringValuePair{
        Key:   "hostname",
        Value: HostName(),
    }
    props = append(props, hostname)

    language := &common.KeyStringValuePair{
        Key:   "language",
        Value: "go",
    }
    props = append(props, language)

    osName := &common.KeyStringValuePair{
        Key:   "OS Name",
        Value: OSName(),
    }
    props = append(props, osName)

    ipv4s := AllIPV4()
    if len(ipv4s) > 0 {
        for _, ipv4 := range ipv4s {
            kv := &common.KeyStringValuePair{
                Key:   "ipv4",
                Value: ipv4,
            }
            props = append(props, kv)
        }
    }
    return
}
