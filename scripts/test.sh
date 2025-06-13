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
  $CLIENT -mode stat -server localhost:$PORT >> "$SUMMARY_FILE" 2>&1
}

run_client() {
  local label=$1
  local nclients=$2
  local valuesize=$3
  local operations=$4
  local mode=$5

  run_server

  echo "Ejecutando $label modo $mode con $nclients clientes concurrentes, value size $valuesize, operaciones $operations"
  start=$(date +%s.%N)
  $CLIENT -server localhost:$PORT -mode $mode -count $operations -clients $nclients -prefix $label -value-size $valuesize > "$LOG_DIR/${label}_${mode}.log" 2>&1
  end=$(date +%s.%N)
  duration=$(echo "$end - $start" | bc)

  echo "Duración total para $label ($mode): $duration segundos" | tee -a "$SUMMARY_FILE"
  run_stat
  kill_server
  echo "" >> "$SUMMARY_FILE"
}

echo "Iniciando tests..."

# Experimento 1: Latencia por tamaño de valor
echo "[Experimento 1] Latencia por tamaño de valor" | tee -a "$SUMMARY_FILE"
for size in 128 512 1024 4096 8192; do
  run_client "exp1_size_${size}B" 1 $size 100 set
  run_client "exp1_size_${size}B" 1 $size 100 get
  run_client "exp1_size_${size}B" 1 $size 10 getprefix
done

# Experimento 2: Lecturas frías vs calientes (durabilidad)
echo "[Experimento 2] Lecturas frías vs calientes (durabilidad)" | tee -a "$SUMMARY_FILE"
run_client "exp2_durability_512B_pre" 1 512 50 set
run_client "exp2_durability_512B_post" 1 512 50 get

# Experimento 3: Escalabilidad por número de clientes concurrentes
echo "[Experimento 3] Escalabilidad por número de clientes" | tee -a "$SUMMARY_FILE"
for clients in 1 4 8 16; do
  run_client "exp3_concurrency_${clients}c" $clients 1024 100 set
  run_client "exp3_concurrency_${clients}c" $clients 1024 100 get
done

echo "Todos los experimentos completados. Resultados resumidos en $SUMMARY_FILE"
