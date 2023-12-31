#!/bin/bash
rm -rf /tmp/sedrad-temp

sedrad --devnet --appdir=/tmp/sedrad-temp --profile=6061 --loglevel=debug &
SEDRAD_PID=$!
SEDRAD_KILLED=0
function killSedradIfNotKilled() {
    if [ $SEDRAD_KILLED -eq 0 ]; then
      kill $SEDRAD_PID
    fi
}
trap "killSedradIfNotKilled" EXIT

sleep 1

application-level-garbage --devnet -alocalhost:22611 -b blocks.dat --profile=7000
TEST_EXIT_CODE=$?

kill $SEDRAD_PID

wait $SEDRAD_PID
SEDRAD_KILLED=1
SEDRAD_EXIT_CODE=$?

echo "Exit code: $TEST_EXIT_CODE"
echo "sedrad exit code: $SEDRAD_EXIT_CODE"

if [ $TEST_EXIT_CODE -eq 0 ] && [ $SEDRAD_EXIT_CODE -eq 0 ]; then
  echo "application-level-garbage test: PASSED"
  exit 0
fi
echo "application-level-garbage test: FAILED"
exit 1
