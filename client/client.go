package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	pb "github.com/lottotto/stdgrpc/api/proto"
	"github.com/lottotto/stdgrpc/utils"
	"google.golang.org/grpc"
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
		resp, err := client.ExampleGet(context.Background(), &pb.ExampleRequest{Name: queryParam})
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
		resp, err := client.ExamplePost(context.Background(), &pb.ExampleRequest{Name: req.GetName()})
		w.Write([]byte(resp.GetMessage()))
		return
	default:
		w.WriteHeader(405)
		return
	}
}

func main() {
	grpcHost := utils.GetEnv("GRPC_HOST", "localhost")
	grpcPort := utils.GetEnv("GRPC_PORT", "50051")

	conn, err := grpc.Dial(grpcHost+":"+grpcPort,
		grpc.WithInsecure(),
		grpc.WithBlock(),
	)
	defer conn.Close()
	if err != nil {
		fmt.Printf("could not connect: %v", err)
	}

	c := Controller{conn: conn}

	http.HandleFunc("/", c.userHandler)
	fmt.Println("start http server")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
