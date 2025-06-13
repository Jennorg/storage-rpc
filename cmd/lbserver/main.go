package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"sync/atomic"
	"syscall"

	"storage-rpc/internal/kvstore"
	pb "storage-rpc/pkg/api"

	"google.golang.org/grpc"
)

type server struct {
	pb.UnimplementedKeyValueStoreServer
	store            *kvstore.KVStore
	totalSets        uint64
	totalGets        uint64
	totalGetPrefixes uint64
}

func (s *server) Set(ctx context.Context, req *pb.SetRequest) (*pb.SetResponse, error) {
	err := s.store.Set(req.Key, req.Value)
	if err != nil {
		return nil, fmt.Errorf("failed to set key %s: %w", req.Key, err)
	}
	atomic.AddUint64(&s.totalSets, 1)
	log.Println("SET OK")
	return &pb.SetResponse{}, nil
}

func (s *server) Get(ctx context.Context, req *pb.GetRequest) (*pb.GetResponse, error) {
	value, ok := s.store.Get(req.Key)
	if !ok {
		return nil, fmt.Errorf("key %s not found", req.Key)
	}
	atomic.AddUint64(&s.totalGets, 1)
	log.Println("GET OK")
	return &pb.GetResponse{Value: value}, nil
}

func (s *server) GetPrefix(ctx context.Context, req *pb.GetPrefixRequest) (*pb.GetPrefixResponse, error) {
	kvMap := s.store.GetPrefix(req.Prefix)
	atomic.AddUint64(&s.totalGetPrefixes, 1)
	log.Println("GETPREFIX OK")

	pairs := make([]*pb.KeyValuePair, 0, len(kvMap))
	for k, v := range kvMap {
		pairs = append(pairs, &pb.KeyValuePair{
			Key:   k,
			Value: v,
		})
	}

	return &pb.GetPrefixResponse{
		Pairs: pairs,
	}, nil
}

func (s *server) Stat(ctx context.Context, req *pb.StatRequest) (*pb.StatResponse, error) {
	return &pb.StatResponse{
		TotalSets:        atomic.LoadUint64(&s.totalSets),
		TotalGets:        atomic.LoadUint64(&s.totalGets),
		TotalGetprefixes: atomic.LoadUint64(&s.totalGetPrefixes),
	}, nil
}

func main() {
	log.Println("Iniciando servidor...")

	store, err := kvstore.NewKVStore("logs/wal.log")
	if err != nil {
		log.Fatalf("Error iniciando KVStore: %v", err)
	}

	lis, err := net.Listen("tcp", "127.0.0.1:50051")
	if err != nil {
		log.Fatalf("Error escuchando: %v", err)
	}

	s := grpc.NewServer(
		grpc.MaxRecvMsgSize(10*1024*1024), // 10 MB LIMITE
		grpc.MaxSendMsgSize(10*1024*1024),
	)

	kvServer := &server{store: store}
	pb.RegisterKeyValueStoreServer(s, kvServer)

	go func() {
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
		<-sig
		fmt.Println("\nEstadísticas finales:")
		fmt.Printf("#total_sets: %d\n", atomic.LoadUint64(&kvServer.totalSets))
		fmt.Printf("#total_gets: %d\n", atomic.LoadUint64(&kvServer.totalGets))
		fmt.Printf("#total_getprefixes: %d\n", atomic.LoadUint64(&kvServer.totalGetPrefixes))
		s.GracefulStop()
	}()

	if err := s.Serve(lis); err != nil {
		log.Fatalf("Fallo al servir: %v", err)
	}
}
