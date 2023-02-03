package healthchecker_test

import (
    "context"
    "fmt"
    "net"
    "strconv"
    "time"

    . "github.com/onsi/ginkgo/v2"
    . "github.com/onsi/gomega"
    "github.com/solo-io/go-utils/healthchecker"
    "google.golang.org/grpc"
    "google.golang.org/grpc/health"
    healthpb "google.golang.org/grpc/health/grpc_health_v1"
    "google.golang.org/grpc/reflection"
)

type testGRPCServer struct {
    Port          uint32
    HealthChecker healthchecker.HealthChecker
}

var (
    serviceName = "TestService"
)

func RunServer(ctx context.Context) *testGRPCServer {
    lis, err := net.Listen("tcp", ":0")
    if err != nil {
        panic(err)
    }
    grpcServer := grpc.NewServer()
    reflection.Register(grpcServer)
    hc := healthchecker.NewGrpc(serviceName, health.NewServer(), false, healthpb.HealthCheckResponse_SERVING)
    healthpb.RegisterHealthServer(grpcServer, hc.GetServer())
    go grpcServer.Serve(lis)
    time.Sleep(time.Millisecond)

    addr := lis.Addr().String()
    _, portstr, err := net.SplitHostPort(addr)
    if err != nil {
        panic(err)
    }

    port, err := strconv.Atoi(portstr)
    if err != nil {
        panic(err)
    }

    srv := &testGRPCServer{
        Port:          uint32(port),
        HealthChecker: hc,
    }

    return srv
}

var _ = Describe("grpc healthchecker", func() {

    var (
        ctx    context.Context
        cancel context.CancelFunc

        conn   *grpc.ClientConn
        client healthpb.HealthClient
        srv    *testGRPCServer
    )

    BeforeEach(func() {
        ctx, cancel = context.WithCancel(context.Background())
        srv = RunServer(ctx)
        var err error
        conn, err = grpc.DialContext(ctx, fmt.Sprintf("localhost:%d", srv.Port), grpc.WithInsecure())
        Expect(err).NotTo(HaveOccurred())
        client = healthpb.NewHealthClient(conn)
    })

    AfterEach(func() {
        cancel()
    })

    Context("without service name", func() {
        It("can recieve serving from a healthy server", func() {
            resp, err := client.Check(ctx, &healthpb.HealthCheckRequest{})
            Expect(err).NotTo(HaveOccurred())
            Expect(resp.Status).To(Equal(healthpb.HealthCheckResponse_SERVING))
        })
    })

    Context("with service name", func() {

        ExpectStatus := func(s healthpb.HealthCheckResponse_ServingStatus) {
            resp, err := client.Check(ctx, &healthpb.HealthCheckRequest{
                Service: serviceName,
            })
            Expect(err).NotTo(HaveOccurred())
            Expect(resp.Status).To(Equal(s))
        }

        It("can recieve serving from a healthy server", func() {
            ExpectStatus(healthpb.HealthCheckResponse_SERVING)
        })

        It("can receive not serving from an unhealthy server", func() {
            srv.HealthChecker.Fail()
            ExpectStatus(healthpb.HealthCheckResponse_NOT_SERVING)
        })

        It("can receive not serving from an unhealthy server, and then serving when it becomes health", func() {
            srv.HealthChecker.Fail()
            ExpectStatus(healthpb.HealthCheckResponse_NOT_SERVING)

            srv.HealthChecker.Ok()
            ExpectStatus(healthpb.HealthCheckResponse_SERVING)
        })
    })
})
