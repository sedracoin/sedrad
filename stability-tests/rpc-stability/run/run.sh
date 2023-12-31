#!/bin/bash
rm -rf /tmp/sedrad-temp

sedrad --devnet --appdir=/tmp/sedrad-temp --profile=6061 --loglevel=debug &
SEDRAD_PID=$!

sleep 1

rpc-stability --devnet -p commands.json --profile=7000
TEST_EXIT_CODE=$?

kill $SEDRAD_PID

wait $SEDRAD_PID
SEDRAD_EXIT_CODE=$?

echo "Exit code: $TEST_EXIT_CODE"
echo "sedrad exit code: $SEDRAD_EXIT_CODE"

if [ $TEST_EXIT_CODE -eq 0 ] && [ $SEDRAD_EXIT_CODE -eq 0 ]; then
  echo "rpc-stability test: PASSED"
  exit 0
fi
echo "rpc-stability test: FAILED"
exit 1
