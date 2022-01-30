FROM golang:1.17.1 as builder
COPY . /work
WORKDIR /work
# RUN  GOOS=linux GOARCH=amd64 go build server/main.go && mv main /bin/main
RUN  GOOS=linux GOARCH=amd64 go build -o dist/server server/server.go

FROM gcr.io/distroless/base:debug
COPY --from=builder /work/dist/server /bin/server
EXPOSE 50051
ENTRYPOINT ["/bin/server"]