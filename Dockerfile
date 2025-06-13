# --- Stage 1: Build the Go application ---
# Usamos una imagen base de Go para compilar. Alpine es ligera.
FROM golang:1.22-alpine AS builder

# Establece el directorio de trabajo dentro del contenedor
WORKDIR /app

# Copia los archivos go.mod y go.sum para descargar dependencias primero.
# Esto permite que Docker cachee las dependencias si no cambian.
COPY go.mod go.sum ./

# Descarga las dependencias
RUN go mod download

# Copia todo el código fuente del proyecto al directorio de trabajo en el contenedor
# Incluye cmd/, pkg/, etc.
COPY . .

# Compila los binarios del servidor y cliente
# `-o` especifica el nombre del archivo de salida
# `./cmd/lbserver` y `./cmd/lbclient` son las rutas a tus paquetes main
RUN go build -o ./bin/lbserver ./cmd/lbserver
RUN go build -o ./bin/lbclient ./cmd/lbclient

# --- Stage 2: Create the final, smaller image ---
# Usamos una imagen base mucho más pequeña para el contenedor final.
# Alpine es una buena opción para aplicaciones Go.
FROM alpine:latest

# Establece el directorio de trabajo
WORKDIR /app

# Copia los binarios compilados desde la etapa 'builder' a la imagen final
COPY --from=builder /app/bin/lbserver ./bin/lbserver
COPY --from=builder /app/bin/lbclient ./bin/lbclient

EXPOSE 50051
