package main

import (
	"flag"
	"fmt"
	"log"
	"math/rand"
	"time"

	"context"

	pb "storage-rpc/pkg/api"

	"google.golang.org/grpc"
)

func main() {

	serverAddr := flag.String("server", "localhost:50051", "Servidor gRPC")
	mode := flag.String("mode", "set", "Modo de operacion: set, get, getprefix")
	count := flag.Int("count", 1000, "Cantidad de operaciones a ejecutar")
	prefix := flag.String("prefix", "key", "Prefijo de las claves")
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
			value := randomString(100)
			_, err := client.Set(context.Background(), &pb.SetRequest{Key: key, Value: value})
			if err != nil {
				log.Printf("[SET FAIL] %v", err)
			} else {
				log.Println("SET OK")
			}
		}

	case "get":
		for i := 0; i < *count; i++ {
			key := fmt.Sprintf("%s-%d", *prefix, i)
			_, err := client.Get(context.Background(), &pb.GetRequest{Key: key})
			if err != nil {
				log.Printf("[GET FAIL] %v", err)
			} else {
				log.Println("GET OK")
			}
		}

	case "getprefix":
		for i := 0; i < *count; i++ {
			_, err := client.GetPrefix(context.Background(), &pb.GetPrefixRequest{Prefix: *prefix})
			if err != nil {
				log.Printf("[GETPREFIX FAIL] %v", err)
			} else {
				log.Println("GETPREFIX OK")
			}
		}

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
