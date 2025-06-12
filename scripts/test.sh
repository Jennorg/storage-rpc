#!/bin/bash

SERVER=./bin/lbserver
CLIENT=./bin/lbclient
LOG_DIR=./logs
RESULTS_DIR=./results
SUMMARY_FILE=$RESULTS_DIR/summary_results.txt
PORT=50051

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
  $CLIENT -mode stat >> "$SUMMARY_FILE" 2>&1
}

run_client() {
  local label=$1
  local nclients=$2
  local valuesize=$3
  local operations=$4
  local phase=$5

  local client_output="$LOG_DIR/${label}.log"

  echo "Ejecutando $label"
  run_server

  start=$(date +%s.%N)
  $CLIENT -server localhost:$PORT -mode set -count $operations -prefix $label -value-size $valuesize > "$client_output" 2>&1
  end=$(date +%s.%N)
  duration=$(echo "$end - $start" | bc)

  echo "Duración total para $label ($nclients clientes): $duration segundos" | tee -a "$SUMMARY_FILE"

  run_stat
  kill_server
}

echo "Running tests..."

# Experimento 1: Latencia por tamaño de valor
echo "[Experimento 1] Latencia por tamaño de valor" | tee -a "$SUMMARY_FILE"
for size in 128 512 1024 4096 8192; do
  run_client "exp1_size_${size}B" 1 $size 100
done

# Experimento 2: Lecturas frías vs calientes (durabilidad)
echo "[Experimento 2] Lecturas frías vs calientes (durabilidad)" | tee -a "$SUMMARY_FILE"
run_client "exp2_durability_512B_pre" 1 512 50 "-putget"
run_client "exp2_durability_512B_post" 1 512 50 "-getonly"

# Experimento 3: Escalabilidad por número de clientes
echo "[Experimento 3] Escalabilidad por número de clientes (valor fijo 1024B)" | tee -a "$SUMMARY_FILE"
for clients in 1 4 8 16; do
  run_client "exp3_concurrency_${clients}c" $clients 1024 100
done

echo "Todos los experimentos completados. Resultados resumidos en $SUMMARY_FILE"
