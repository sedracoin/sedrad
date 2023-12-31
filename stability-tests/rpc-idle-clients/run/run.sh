#!/bin/bash
rm -rf /tmp/sedrad-temp

NUM_CLIENTS=128
sedrad --devnet --appdir=/tmp/sedrad-temp --profile=6061 --rpcmaxwebsockets=$NUM_CLIENTS &
SEDRAD_PID=$!
SEDRAD_KILLED=0
function killSedradIfNotKilled() {
  if [ $SEDRAD_KILLED -eq 0 ]; then
    kill $SEDRAD_PID
  fi
}
trap "killSedradIfNotKilled" EXIT

sleep 1

rpc-idle-clients --devnet --profile=7000 -n=$NUM_CLIENTS
TEST_EXIT_CODE=$?

kill $SEDRAD_PID

wait $SEDRAD_PID
SEDRAD_EXIT_CODE=$?
SEDRAD_KILLED=1

echo "Exit code: $TEST_EXIT_CODE"
echo "sedrad exit code: $SEDRAD_EXIT_CODE"

if [ $TEST_EXIT_CODE -eq 0 ] && [ $SEDRAD_EXIT_CODE -eq 0 ]; then
  echo "rpc-idle-clients test: PASSED"
  exit 0
fi
echo "rpc-idle-clients test: FAILED"
exit 1
