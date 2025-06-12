#!/bin/bash

set -e

PORT=50051
SERVER_LOG="server_output.log"
CLIENT_LOG="client_output.log"

echo "==== [0] Verificando si el puerto $PORT está libre ===="
if lsof -i :$PORT -sTCP:LISTEN -t >/dev/null; then
    echo "Error: El puerto $PORT ya está en uso. Aborta el script."
    lsof -i :$PORT
    exit 1
fi

echo "==== [1] Limpiando logs anteriores ===="
rm -f $SERVER_LOG $CLIENT_LOG

echo "==== [2] Iniciando el servidor ===="
./bin/lbserver 2>&1 | tee $SERVER_LOG &
SERVER_PID=$!
sleep 2 

echo "Servidor iniciado (PID $SERVER_PID)"

if ps -p $SERVER_PID > /dev/null; then
    echo "Servidor está corriendo"
else
    echo "Servidor no está corriendo, revisa $SERVER_LOG"
    exit 1
fi

echo "==== [3] Ejecutando clientes concurrentes ===="
START_TIME=$(date)
echo "Hora de inicio del servidor: $START_TIME"

NUM_CLIENTS=3
OPS_PER_CLIENT=100
PREFIX="key"

for i in $(seq 1 $NUM_CLIENTS); do
    echo "-> Lanzando cliente set #$i"
    ./bin/lbclient --mode=set --count=$OPS_PER_CLIENT --prefix=$PREFIX >> $CLIENT_LOG 2>&1 &
    
    echo "-> Lanzando cliente get #$i"
    ./bin/lbclient --mode=get --count=$OPS_PER_CLIENT --prefix=$PREFIX >> $CLIENT_LOG 2>&1 &
    
    echo "-> Lanzando cliente getprefix #$i"
    ./bin/lbclient --mode=getprefix --count=$OPS_PER_CLIENT --prefix=$PREFIX >> $CLIENT_LOG 2>&1 &
done

echo "==== [4] Esperando a que terminen los clientes ===="
wait
echo "Todos los clientes han terminado."

echo "==== [5] Deteniendo servidor (PID $SERVER_PID) ===="
kill -INT $SERVER_PID
wait $SERVER_PID || true

echo "==== [6] Mostrando estadísticas ===="

echo -e "\n========== ESTADÍSTICAS DEL CLIENTE =========="
echo "#total_sets completados:      $(grep -c 'SET OK' $CLIENT_LOG)"
echo "#total_gets completados:      $(grep -c 'GET OK' $CLIENT_LOG)"
echo "#total_getprefixes completados: $(grep -c 'GETPREFIX OK' $CLIENT_LOG)"

echo -e "\n========== ESTADÍSTICAS DEL SERVIDOR =========="
grep "#total_" $SERVER_LOG || echo "(no se encontraron estadísticas en el log del servidor)"

END_TIME=$(date)
echo "Hora de finalización: $END_TIME"
