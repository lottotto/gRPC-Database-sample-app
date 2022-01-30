package main

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"

	pb "github.com/lottotto/mygrpc/api/proto"
	"github.com/lottotto/mygrpc/utils"
	"google.golang.org/grpc"
)

// var tracer = otel.Tracer("github.com/lottotto/mygrpc")

var (
	// tp       *sdktrace.TracerProvider
	grpcHost string
	grpcPort string
	conn     *grpc.ClientConn
)

func userHandler(w http.ResponseWriter, r *http.Request) {
	client := pb.NewExampleServiceClient(conn)
	switch {
	case r.Method == http.MethodGet:
		// Todo: 何かに従ってDBから取得する処理を獲得
	case r.Method == http.MethodPost:
		// Bodyに貼り付けられたNameをもとにDBに着込むようにgRPCサーバにリクエストする
		buf := new(bytes.Buffer)
		io.Copy(buf, r.Body)
		var req *pb.ExampleRequest
		err := json.Unmarshal(buf.Bytes(), &req)
		if err != nil {
			log.Fatalf("cannot unmarshal body: %v", err)
		}
		resp, err := client.ExamplePost(r.Context(), req)
		w.Write([]byte(resp.GetMessage()))
	}
}

func main() {

	grpcHost = utils.GetEnv("GRPC_HOST", "localhost")
	grpcHost = utils.GetEnv("GRPC_PORT", "50051")
	conn, err := grpc.Dial(grpcHost+":"+grpcPort,
		grpc.WithInsecure(),
		grpc.WithBlock(),
	// grpc.WithUnaryInterceptor(grpc_middleware.ChainUnaryClient(
	// otelgrpc.UnaryClientInterceptor(),
	// grpc_prometheus.DefaultClientMetrics.UnaryClientInterceptor())),
	)
	defer conn.Close()
	if err != nil {
		log.Fatalf("cannot connect gRPC server: %v", err)
	}

	http.HandleFunc("/", userHandler)
	log.Fatal(http.ListenAndServe(":8080", nil))

	// tp = utils.Init()

	// otel.SetTracerProvider(tp)
	// otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))

	// if err != nil {
	// 	log.Fatalf("could not connect: %v", err)
	// 	os.Exit(1)
	// }
	// defer conn.Close()

	// e := gin.Default()
	// pprof.Register(e)
	// e.Use(otelgin.Middleware("my-server"))
	// e.GET("/grpc/:message", getGRPC)
	// e.POST("/grpc", postGRPC)
	// e.GET("/metrics", prometheusHandler())
	// e.GET("/hello", func(c *gin.Context) {
	// 	// Tracer名は指定を指定することで,指定されたトレーサーを作ることができる
	// 	_, span := tp.Tracer("hello").Start(c.Request.Context(), "Hello")
	// 	defer span.End()
	// 	c.JSON(http.StatusOK, gin.H{
	// 		"message": "aaaa",
	// 	})
	// })

	// e.Run()
}
