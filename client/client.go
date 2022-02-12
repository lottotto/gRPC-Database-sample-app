package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	pb "github.com/lottotto/stdgrpc/api/proto"
	"github.com/lottotto/stdgrpc/utils"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
)

type Controller struct {
	conn *grpc.ClientConn
}

// Todo: メソッドにするときに親の構造体にアドレスをつける理由を調べる
func (c *Controller) userHandler(w http.ResponseWriter, r *http.Request) {
	client := pb.NewExampleServiceClient(c.conn)
	fmt.Printf("method: %v", r.Method)
	switch {
	case r.Method == http.MethodGet:
		queryParam := r.URL.Query().Get("name")
		log.Printf("param recieved: %v\n", queryParam)
		resp, err := client.ExampleGet(r.Context(), &pb.ExampleRequest{Name: queryParam})
		if err != nil {
			log.Fatalf("could not connect gRPC server: %v", err)
		}
		w.Write([]byte(resp.GetMessage()))

	case r.Method == http.MethodPost:
		// Bodyに貼り付けられたNameをもとにDBに着込むようにgRPCサーバにリクエストする
		log.Println("gRPC client made")
		buf := new(bytes.Buffer)
		io.Copy(buf, r.Body)
		var req *pb.ExampleRequest
		err := json.Unmarshal(buf.Bytes(), &req)
		if err != nil {
			log.Fatalf("cannot unmarshal body: %v", err)
		}
		resp, err := client.ExamplePost(r.Context(), &pb.ExampleRequest{Name: req.GetName()})
		w.Write([]byte(resp.GetMessage()))
		return
	default:
		w.WriteHeader(405)
		return
	}
}

func main() {
	// トレースの追加
	tp, err := utils.InitTraceProviderStdOut("gRPC", "1.0.0")
	otel.SetTracerProvider(tp)
	// 後続のサービスにつなげるためにpropagaterを追加
	otel.SetTextMapPropagator(
		propagation.NewCompositeTextMapPropagator(
			propagation.TraceContext{},
			propagation.Baggage{},
		),
	)

	grpcHost := utils.GetEnv("GRPC_HOST", "localhost")
	grpcPort := utils.GetEnv("GRPC_PORT", "50051")

	conn, err := grpc.Dial(grpcHost+":"+grpcPort,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithUnaryInterceptor(otelgrpc.UnaryClientInterceptor()),
	)
	defer conn.Close()
	if err != nil {
		fmt.Printf("could not connect: %v", err)
	}

	c := Controller{conn: conn}
	// otelhttp用のオプションが必要？？？
	otelOptions := []otelhttp.Option{
		otelhttp.WithTracerProvider(otel.GetTracerProvider()),
		otelhttp.WithPropagators(otel.GetTextMapPropagator()),
	}
	// Todo: 3点リーダつけると何が起きるのか調べる
	otelUserHandler := otelhttp.NewHandler(
		http.HandlerFunc(c.userHandler),
		"UserHandler",
		otelOptions...,
	)

	// http.HandleFunc("/", c.userHandler)
	http.HandleFunc("/", otelUserHandler.ServeHTTP)
	fmt.Println("start http server")
	log.Fatal(http.ListenAndServe(":8080", nil))

	// 直接gRPCを実行
	// client := pb.NewExampleServiceClient(conn)
	// _, err = client.ExamplePost(context.Background(), &pb.ExampleRequest{Name: "saito"})

	// time.Sleep(10 * time.Second)
}
