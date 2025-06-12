package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	pb "storage-rpc/pkg/api"

	"google.golang.org/grpc"
)

func main() {
	reader := bufio.NewReader(os.Stdin)
	conn, err := grpc.Dial("localhost:50051", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("No se pudo conectar al servidor: %v", err)
	}
	defer conn.Close()

	client := pb.NewKeyValueStoreClient(conn)

	fmt.Println("=== Interfaz Cliente RPC ===")
	for {
		fmt.Print("\nComando (set/get/getprefix/stat/exit): ")
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)

		switch input {
		case "set":
			fmt.Print("Clave: ")
			key, _ := reader.ReadString('\n')
			fmt.Print("Valor: ")
			val, _ := reader.ReadString('\n')
			start := time.Now()
			_, err := client.Set(context.Background(), &pb.SetRequest{
				Key:   strings.TrimSpace(key),
				Value: strings.TrimSpace(val),
			})
			if err != nil {
				fmt.Println("Error en set:", err)
			} else {
				fmt.Println("SET OK. Latencia:", time.Since(start))
			}

		case "get":
			fmt.Print("Clave: ")
			key, _ := reader.ReadString('\n')
			start := time.Now()
			resp, err := client.Get(context.Background(), &pb.GetRequest{
				Key: strings.TrimSpace(key),
			})
			if err != nil {
				fmt.Println("Error en get:", err)
			} else {
				fmt.Printf("GET OK. Valor: %s (latencia: %v)\n", resp.Value, time.Since(start))
			}

		case "getprefix":
			fmt.Print("Prefijo: ")
			prefix, _ := reader.ReadString('\n')
			start := time.Now()
			resp, err := client.GetPrefix(context.Background(), &pb.GetPrefixRequest{
				Prefix: strings.TrimSpace(prefix),
			})
			if err != nil {
				fmt.Println("Error en getprefix:", err)
			} else {
				fmt.Printf("GETPREFIX OK (%v). Claves:\n", time.Since(start))
				for _, pair := range resp.Pairs {
					fmt.Printf("- %s: %s\n", pair.Key, pair.Value)
				}
			}

		case "stat":
			start := time.Now()
			resp, err := client.Stat(context.Background(), &pb.StatRequest{})
			if err != nil {
				fmt.Println("Error en stat:", err)
			} else {
				fmt.Printf("STAT OK (%v)\n", time.Since(start))
				fmt.Printf("Total Sets:        %d\n", resp.TotalSets)
				fmt.Printf("Total Gets:        %d\n", resp.TotalGets)
				fmt.Printf("Total GetPrefixes: %d\n", resp.TotalGetprefixes)
			}

		case "exit":
			fmt.Println("Saliendo.")
			return

		default:
			fmt.Println("Comando no reconocido.")
		}
	}
}
