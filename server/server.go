package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net"
	_ "net/http/pprof"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	semconv "go.opentelemetry.io/otel/semconv/v1.7.0"

	pb "github.com/lottotto/stdgrpc/api/proto"
	"github.com/lottotto/stdgrpc/utils"
	"google.golang.org/grpc"

	"github.com/XSAM/otelsql"
)

type User struct {
	Name   string `json:"name" db:"name"`
	Number int    `json:"number" db:"number"`
}

type server struct {
	pb.UnimplementedExampleServiceServer
	// DBのコネクションプールは,serverに持たせる。globalだと失敗する
	conn *sqlx.DB
}

func (s *server) ExampleGet(ctx context.Context, in *pb.ExampleRequest) (*pb.ExampleResponse, error) {

	log.Printf("ExampleGet Recieved: %v", in.GetName())

	query := `SELECT * FROM example where NAME=$1`
	// Todo: QueryxContextとQueryContextの違いを調べる
	rows, err := s.conn.QueryxContext(ctx, query, in.GetName())
	if err != nil {
		log.Fatalf("could not get rows: \n%v", err)
	}
	var users []User
	rows.Scan(&users)
	for rows.Next() {
		var u User
		// Todo: Scanの方法について調べておく
		err = rows.StructScan(&u)
		users = append(users, u)
	}
	b, err := json.Marshal(users)
	if err != nil {
		log.Fatalf("could not marshal: %v", err)
	}

	return &pb.ExampleResponse{Message: string(b)}, nil
}

func (s *server) ExamplePost(ctx context.Context, in *pb.ExampleRequest) (*pb.ExampleResponse, error) {

	log.Printf("ExamplePost Recieved: %v", in.GetName())
	// Todo: MustExecとExecの違いを調べる
	query := `INSERT INTO example (NAME, NUMBER) VALUES ($1, $2)`
	_, err := s.conn.ExecContext(ctx, query, in.GetName(), rand.Intn(100))
	if err != nil {
		log.Fatalf("could not insert: %v", err)
		return &pb.ExampleResponse{Message: "grpc_ng"}, nil
	}

	return &pb.ExampleResponse{Message: "grpc_ok"}, nil
}

func main() {
	// tracerの設定
	tp, err := utils.InitTraceProviderStdOut("gRPC", "1.0.0")
	if err != nil {
		log.Fatalf("something error: %v", err)
	}
	otel.SetTracerProvider(tp)
	// 受け取る側にも必要→超ハマった
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))

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
	log.Printf("server listening at %v", lis.Addr())

	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}

}
