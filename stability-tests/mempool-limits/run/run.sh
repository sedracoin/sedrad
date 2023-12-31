#!/bin/bash

APPDIR=/tmp/sedrad-temp
SEDRAD_RPC_PORT=29587

rm -rf "${APPDIR}"

sedrad --simnet --appdir="${APPDIR}" --rpclisten=0.0.0.0:"${SEDRAD_RPC_PORT}" --profile=6061 &
SEDRAD_PID=$!

sleep 1

RUN_STABILITY_TESTS=true go test ../ -v -timeout 86400s -- --rpc-address=127.0.0.1:"${SEDRAD_RPC_PORT}" --profile=7000
TEST_EXIT_CODE=$?

kill $SEDRAD_PID

wait $SEDRAD_PID
SEDRAD_EXIT_CODE=$?

echo "Exit code: $TEST_EXIT_CODE"
echo "sedrad exit code: $SEDRAD_EXIT_CODE"

if [ $TEST_EXIT_CODE -eq 0 ] && [ $SEDRAD_EXIT_CODE -eq 0 ]; then
  echo "mempool-limits test: PASSED"
  exit 0
fi
echo "mempool-limits test: FAILED"
exit 1
