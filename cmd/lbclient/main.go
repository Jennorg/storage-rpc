package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"sync"
	"time"

	pb "storage-rpc/pkg/api"

	"google.golang.org/grpc"
	insecure "google.golang.org/grpc/credentials/insecure"
)

func main() {
	serverAddr := flag.String("server", "localhost:50051", "Servidor gRPC")
	mode := flag.String("mode", "set", "Modo de operacion: set, get, getprefix, stat")
	count := flag.Int("count", 1000, "Cantidad de operaciones a ejecutar")
	prefix := flag.String("prefix", "key", "Prefijo de las claves")
	valueSize := flag.Int("value-size", 100, "Tamaño del valor en bytes")
	clients := flag.Int("clients", 1, "Número de clientes concurrentes (goroutines)")

	flag.Parse()

	conn, err := grpc.Dial(*serverAddr,
		grpc.WithDefaultCallOptions(
			grpc.MaxCallRecvMsgSize(10*1024*1024),
			grpc.MaxCallSendMsgSize(10*1024*1024),
		),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)

	if err != nil {
		log.Fatalf("No se pudo conectar: %v", err)
	}
	defer conn.Close()
	client := pb.NewKeyValueStoreClient(conn)

	opsPerClient := *count / *clients
	if opsPerClient == 0 {
		opsPerClient = *count
	}

	var wg sync.WaitGroup
	wg.Add(*clients)

	for c := 0; c < *clients; c++ {
		go func(clientID int) {
			defer wg.Done()
			startIndex := clientID * opsPerClient
			endIndex := startIndex + opsPerClient
			if clientID == *clients-1 {
				endIndex = *count // El último puede hacer más si no es exacto
			}

			switch *mode {
			case "set":
				for i := startIndex; i < endIndex; i++ {
					key := fmt.Sprintf("%s-%d", *prefix, i)
					value := randomString(*valueSize)
					start := time.Now()
					_, err := client.Set(context.Background(), &pb.SetRequest{Key: key, Value: value})
					duration := time.Since(start)
					if err != nil {
						log.Printf("[SET FAIL] %v", err)
					} else {
						log.Printf("SET OK %s LATENCY: %v", key, duration)
					}
				}
			case "get":
				for i := startIndex; i < endIndex; i++ {
					key := fmt.Sprintf("%s-%d", *prefix, i)
					start := time.Now()
					_, err := client.Get(context.Background(), &pb.GetRequest{Key: key})
					duration := time.Since(start)
					if err != nil {
						log.Printf("[GET FAIL] %v", err)
					} else {
						log.Printf("GET OK %s LATENCY: %v", key, duration)
					}
				}
			case "getprefix":
				for i := 0; i < opsPerClient; i++ {
					start := time.Now()
					_, err := client.GetPrefix(context.Background(), &pb.GetPrefixRequest{Prefix: *prefix})
					duration := time.Since(start)
					if err != nil {
						log.Printf("[GETPREFIX FAIL] %v", err)
					} else {
						log.Printf("GETPREFIX OK LATENCY: %v", duration)
					}
				}

			case "mix":
				rng := rand.New(rand.NewSource(time.Now().UnixNano()))
				for i := startIndex; i < endIndex; i++ {
					key := fmt.Sprintf("%s-%d", *prefix, i)
					if rng.Intn(2) == 0 {
						value := randomString(*valueSize)
						start := time.Now()
						_, err := client.Set(context.Background(), &pb.SetRequest{Key: key, Value: value})
						duration := time.Since(start)
						if err != nil {
							log.Printf("[MIX-SET FAIL] %v", err)
						} else {
							log.Printf("MIX SET OK %s LATENCY: %v", key, duration)
						}
					} else {
						start := time.Now()
						_, err := client.Get(context.Background(), &pb.GetRequest{Key: key})
						duration := time.Since(start)
						if err != nil {
							log.Printf("[MIX-GET FAIL] %v", err)
						} else {
							log.Printf("MIX GET OK %s LATENCY: %v", key, duration)
						}
					}
				}

			case "stat":
				start := time.Now()
				resp, err := client.Stat(context.Background(), &pb.StatRequest{})
				duration := time.Since(start)
				if err != nil {
					log.Fatalf("[STAT FAIL] %v", err)
				}
				log.Printf("STAT OK LATENCY: %v", duration)
				fmt.Println("=== Estadísticas del servidor ===")
				fmt.Printf("Total Sets:        %d\n", resp.TotalSets)
				fmt.Printf("Total Gets:        %d\n", resp.TotalGets)
				fmt.Printf("Total GetPrefixes: %d\n", resp.TotalGetprefixes)
			default:
				log.Fatalf("Modo no reconocido: %s", *mode)
			}
		}(c)
	}

	wg.Wait()
}

func randomString(n int) string {
	letters := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
	s := make([]rune, n)
	rand.Seed(time.Now().UnixNano())
	for i := range s {
		s[i] = letters[rand.Intn(len(letters))]
	}
	return string(s)
}
