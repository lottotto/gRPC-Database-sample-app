package main

import (
	"context"
	"fmt"
	"log"
	"net"
	_ "net/http/pprof"

	_ "github.com/lib/pq"

	pb "github.com/lottotto/mygrpc/api/proto"
	"google.golang.org/grpc"
)

var (
// tp   *sdktrace.TracerProvider
// conn *sqlx.DB
)

type server struct {
	pb.UnimplementedExampleServiceServer
}

func (s *server) ExampleGet(ctx context.Context, in *pb.ExampleRequest) (*pb.ExampleResponse, error) {
	log.Printf("Recieved: %v", in.GetName())
	return &pb.ExampleResponse{Message: "Hello grpc: " + in.GetName()}, nil
}

func (s *server) ExamplePost(ctx context.Context, in *pb.ExampleRequest) (*pb.ExampleResponse, error) {
	log.Printf("Recieved: %v", in.GetName())
	return &pb.ExampleResponse{Message: "ok"}, nil
}

func main() {
	// tp = utils.Init()
	// conn, err := utils.GetPostgresConnection()
	// if err != nil {
	// 	log.Fatalf("could not connect db: %v", err)
	// }
	// defer conn.Close()

	lis, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", 50051))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	defer lis.Close()
	s := grpc.NewServer(
	// grpc.UnaryInterceptor(
	// 	grpc_middleware.ChainUnaryServer(
	// 		otelgrpc.UnaryServerInterceptor(),
	// 		grpc_prometheus.UnaryServerInterceptor,
	// 	),
	// ),
	)

	pb.RegisterExampleServiceServer(s, &server{})
	// grpc_prometheus.Register(s)
	// http.Handle("/metrics", promhttp.Handler())
	// httpServer := &http.Server{Handler: promhttp.Handler(), Addr: fmt.Sprintf("0.0.0.0:%d", 9092)}

	// go func() {
	// 	if err := httpServer.ListenAndServe(); err != nil {
	// 		log.Fatal("Unable to start a http server.")
	// 	}
	// }()
	// log.Printf("server listening at %v", lis.Addr())

	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}

}
