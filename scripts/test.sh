#!/bin/bash

SERVER=./bin/lbserver
CLIENT=./bin/lbclient
LOG_DIR=./logs
RESULTS_DIR=./results
SUMMARY_FILE=$RESULTS_DIR/summary_results.txt
PORT=50051
HOST=127.0.0.1

mkdir -p "$LOG_DIR"
mkdir -p "$RESULTS_DIR"
rm -f "$SUMMARY_FILE"

kill_port_process() {
  pid=$(lsof -t -i:$PORT)
  if [ -n "$pid" ]; then
    echo "Puerto $PORT ocupado por PID $pid, matando proceso..."
    kill -9 $pid
    sleep 0.5
  fi
}

run_server() {
  kill_port_process
  $SERVER > "$LOG_DIR/server.log" 2>&1 &
  SERVER_PID=$!
  sleep 1

  if ! ps -p $SERVER_PID > /dev/null; then
    echo "[ERROR] El servidor falló al iniciar. Revisa $LOG_DIR/server.log"
    cat "$LOG_DIR/server.log" 
    exit 1
  fi
}

kill_server() {
  if ps -p $SERVER_PID > /dev/null 2>&1; then
    kill $SERVER_PID
    wait $SERVER_PID 2>/dev/null
  fi
}

run_stat() {
  echo "[STAT]" | tee -a "$SUMMARY_FILE"
  $CLIENT -mode stat -server $HOST:$PORT >> "$SUMMARY_FILE" 2>&1
}

is_port_open() {
  ss -tuln | grep -q ":$1"
}

wait_for_server() {
  local port=$1
  local retries=30
  local wait_time=0.5
  local count=0

  while ! is_port_open "$port"; do
    echo "Esperando al servidor en puerto $port... intento $count"
    sleep "$wait_time"
    count=$((count + 1))
    if [ "$count" -ge "$retries" ]; then
      echo "[ERROR] Timeout esperando a que el servidor escuche en puerto $port"
      return 1
    fi
  done
  echo "Servidor escuchando en puerto $port"
  return 0
}

run_client() {
  local label=$1
  local nclients=$2
  local valuesize=$3
  local operations=$4
  local mode=$5

  run_server

  if ! wait_for_server $PORT; then
    echo "[ERROR] El servidor no arrancó correctamente."
    kill_server
    exit 1
  fi

  echo "Ejecutando $label modo $mode con $nclients clientes, tamaño $valuesize bytes, $operations ops"
  start=$(date +%s.%N)
  $CLIENT -server $HOST:$PORT -mode $mode -count $operations -clients $nclients -prefix $label -value-size $valuesize > "$LOG_DIR/${label}_${mode}.log" 2>&1
  end=$(date +%s.%N)
  duration=$(echo "$end - $start" | bc)

  echo "Duración total para $label ($mode): $duration segundos" | tee -a "$SUMMARY_FILE"

  if ps -p $SERVER_PID > /dev/null; then
    run_stat
  else
    echo "[STAT FAIL] El servidor murió antes de ejecutar stat" | tee -a "$SUMMARY_FILE"
  fi

  kill_server
  echo "" >> "$SUMMARY_FILE"
}

echo "Iniciando tests..." | tee -a "$SUMMARY_FILE"

##################################
# Experimento 1: Latencia por tamaño de valor
##################################
echo "[Experimento 1] Latencia por tamaño de valor" | tee -a "$SUMMARY_FILE"
for size in 512 4096 524288 1048576 4194304; do
  run_client "exp1_size_${size}B" 1 $size 100 set
  run_client "exp1_size_${size}B" 1 $size 100 get
  run_client "exp1_size_${size}B" 1 $size 100 mix
done

##################################
# Experimento 2: Lecturas frías vs calientes (durabilidad)
##################################
echo "[Experimento 2] Lecturas frías vs calientes" | tee -a "$SUMMARY_FILE"
# run_client "exp2_durability_4KB" 1 4096 100000 set
run_client "exp2_durability_4KB" 1 4096 1000 set

echo "Reiniciando servidor para medir latencias en frío..."
start_ms=$(date +%s%3N)
run_server
end_ms=$(date +%s%3N)
boot_time=$((end_ms - start_ms))
echo "Tiempo de reinicio del servidor: $boot_time ms" | tee -a "$SUMMARY_FILE"

$CLIENT -server $HOST:$PORT -mode get -count 1000 -clients 1 -prefix exp2_durability_4KB -value-size 4096 > "$LOG_DIR/exp2_durability_cold_get.log" 2>&1
kill_server

##################################
# Experimento 3: Escalabilidad por número de clientes
##################################
echo "[Experimento 3] Escalabilidad por número de clientes" | tee -a "$SUMMARY_FILE"
for clients in 1 2 4 8 16 32 64; do
  run_client "exp3_concurrency_${clients}c" $clients 1024 100 mix
done

echo "Todos los experimentos completados. Resultados resumidos en $SUMMARY_FILE"
