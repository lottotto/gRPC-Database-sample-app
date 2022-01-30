FROM golang:1.17.1 as builder
COPY . /work
WORKDIR /work
# RUN  GOOS=linux GOARCH=amd64 go build server/main.go && mv main /bin/main
RUN  GOOS=linux GOARCH=amd64 go build -o dist/client client/client.go 

FROM gcr.io/distroless/base:debug
COPY --from=builder /work/dist/client /bin/client
EXPOSE 8080
EXPOSE 50051
ENTRYPOINT ["/bin/client"]