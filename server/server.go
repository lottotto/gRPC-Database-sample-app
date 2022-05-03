package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	_ "net/http/pprof"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	semconv "go.opentelemetry.io/otel/semconv/v1.7.0"

	pb "github.com/lottotto/stdgrpc/api/proto"
	"github.com/lottotto/stdgrpc/server/dao"
	"github.com/lottotto/stdgrpc/utils"
	"github.com/lottotto/stdgrpc/utils/logger"
	"google.golang.org/grpc"

	"github.com/XSAM/otelsql"
)

// var zaplogger zap.Logger

type server struct {
	pb.UnimplementedExampleServiceServer
	// DBのコネクションプールは,serverに持たせる。globalだと失敗する
	conn *sqlx.DB
}

func (s *server) ExampleGet(ctx context.Context, in *pb.ExampleRequest) (*pb.ExampleResponse, error) {

	log.Printf("ExampleGet Recieved: %v", in.GetName())

	dao := &dao.UserDao{Conn: s.conn}
	users, err := dao.FindByName(ctx, in.GetName())
	if err != nil {
		log.Fatalf("could not get rows: \n%v", err)
	}
	b, err := json.Marshal(users)
	if err != nil {
		log.Fatalf("could not marshal: %v", err)
	}

	return &pb.ExampleResponse{Message: string(b)}, nil
}

func (s *server) ExamplePost(ctx context.Context, in *pb.ExampleRequest) (*pb.ExampleResponse, error) {

	zaplogger, _ := logger.GetZapLogger()
	zaplogger.Info("Example Post Recieved",
		logger.GetOtelLogMetadataFields(ctx)...,
	)

	dao := &dao.UserDao{Conn: s.conn}
	err := dao.Update(ctx, in.GetName())
	if err != nil {
		log.Fatalf("could not insert: %v", err)
		return &pb.ExampleResponse{Message: "grpc_ng"}, nil
	}
	return &pb.ExampleResponse{Message: "grpc_ok"}, nil
}

func main() {
	zaplogger, err := logger.GetZapLogger()
	if err != nil {
		panic(err)
	}
	// Exporter の設定
	exporter, err := utils.GetTraceExporterStdOut()
	if err != nil {
		log.Fatalf("Could not create Exporter: %v \n", err)
	}
	// tracerProviderの設定
	tp, err := utils.InitTraceProvider(exporter, "gRPC", "1.0.0")

	otel.SetTracerProvider(tp)
	// 受け取る側にも必要→超ハマった
	otel.SetTextMapPropagator(
		propagation.NewCompositeTextMapPropagator(
			propagation.TraceContext{},
			propagation.Baggage{},
		),
	)

	driverName, err := otelsql.Register("postgres", semconv.DBSystemPostgreSQL.Value.AsString())
	if err != nil {
		log.Fatalf("something error: %v", err)
	}
	conn, err := sqlx.Open(driverName, utils.GetPostgresConnectionInfo())
	if err != nil {
		log.Fatalf("could not connect db: %v", err)
	}
	defer conn.Close()

	lis, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", 50051))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	defer lis.Close()
	s := grpc.NewServer(
		grpc.UnaryInterceptor(otelgrpc.UnaryServerInterceptor()),
	)

	pb.RegisterExampleServiceServer(s, &server{conn: conn})
	// log.Printf("server listening at %v", lis.Addr())
	// zaplogger.Info("Hello zap")
	// zaplogger.WithOptions(logger.GetOtelLogMetadataFields(context.Background()))
	zaplogger.Info(
		fmt.Sprintf("server listening at %v", lis.Addr()),
		logger.GetOtelLogMetadataFields(context.Background())...,
	)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}

}
