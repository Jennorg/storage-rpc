# Storage-RPC

Servicio de almacenamiento clave-valor persistente basado en RPC con soporte para operaciones concurrentes, durabilidad ante fallos y mediciones de rendimiento.

## Características

* Interfaz gRPC simple con 3 operaciones: `Set`, `Get`, `GetPrefix`.
* Almacenamiento persistente con recuperación desde WAL.
* Soporte para múltiples clientes concurrentes.
* Estadísticas de uso y benchmark automatizado con `test.sh`.

## Instalación

Requisitos:

* Go 1.20+
* `protoc` y plugin Go:

```bash
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
```

Generar los protos:

```bash
make protos
```

Compilar binarios:

```bash
make all
```

## Uso

### Ejecutar servidor

```bash
make server
```

### Ejecutar cliente

```bash
make client
```

Argumentos del cliente:

```bash
--mode       set | get | getprefix | stat
--count      Número de operaciones
--prefix     Prefijo de claves a usar
--server     Dirección del servidor (por defecto localhost:50051)
```

### Benchmark

```bash
make test
```

Muestra estadísticas del cliente y del servidor al finalizar.

## Protocolo de escritura y durabilidad

El sistema escribe cada operación en un archivo WAL (`write-ahead log`) antes de aplicarla en memoria. En caso de fallo, se recarga el estado completo al reiniciar desde ese log, garantizando consistencia y durabilidad.
