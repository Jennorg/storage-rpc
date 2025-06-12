package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"time"

	pb "storage-rpc/pkg/api"

	"google.golang.org/grpc"
)

func main() {
	serverAddr := flag.String("server", "localhost:50051", "Servidor gRPC")
	mode := flag.String("mode", "set", "Modo de operacion: set, get, getprefix, stat")
	count := flag.Int("count", 1000, "Cantidad de operaciones a ejecutar")
	prefix := flag.String("prefix", "key", "Prefijo de las claves")
	valueSize := flag.Int("value-size", 100, "Tamaño del valor en bytes")
	flag.Parse()

	conn, err := grpc.Dial(*serverAddr, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("No se pudo conectar: %v", err)
	}
	defer conn.Close()
	client := pb.NewKeyValueStoreClient(conn)

	switch *mode {
	case "set":
		for i := 0; i < *count; i++ {
			key := fmt.Sprintf("%s-%d", *prefix, i)
			value := randomString(*valueSize) // Usa el tamaño del valor
			start := time.Now()
			_, err := client.Set(context.Background(), &pb.SetRequest{Key: key, Value: value})
			duration := time.Since(start)
			if err != nil {
				log.Printf("[SET FAIL] %v", err)
			} else {
				log.Printf("SET OK, LATENCY: %v", duration) // Imprime la latencia
			}
		}

	case "get":
		for i := 0; i < *count; i++ {
			key := fmt.Sprintf("%s-%d", *prefix, i)
			start := time.Now()
			_, err := client.Get(context.Background(), &pb.GetRequest{Key: key})
			duration := time.Since(start)
			if err != nil {
				log.Printf("[GET FAIL] %v", err)
			} else {
				log.Printf("GET OK, LATENCY: %v", duration) // Imprime la latencia
			}
		}

	case "getprefix":
		for i := 0; i < *count; i++ {
			start := time.Now()
			_, err := client.GetPrefix(context.Background(), &pb.GetPrefixRequest{Prefix: *prefix})
			duration := time.Since(start)
			if err != nil {
				log.Printf("[GETPREFIX FAIL] %v", err)
			} else {
				log.Printf("GETPREFIX OK, LATENCY: %v", duration) // Imprime la latencia
			}
		}

	case "stat":
		start := time.Now()
		resp, err := client.Stat(context.Background(), &pb.StatRequest{})
		duration := time.Since(start)
		if err != nil {
			log.Fatalf("[STAT FAIL] %v", err)
		}
		log.Printf("STAT OK, LATENCY: %v", duration)
		fmt.Println("=== Estadísticas del servidor ===")
		fmt.Printf("Total Sets:        %d\n", resp.TotalSets)
		fmt.Printf("Total Gets:        %d\n", resp.TotalGets)
		fmt.Printf("Total GetPrefixes: %d\n", resp.TotalGetprefixes)

	default:
		log.Fatalf("Modo no reconocido: %s", *mode)
	}
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
