#!/bin/bash
rm -rf /tmp/sedrad-temp

sedrad --simnet --appdir=/tmp/sedrad-temp --profile=6061 &
SEDRAD_PID=$!

sleep 1

orphans --simnet -alocalhost:22511 -n20 --profile=7000
TEST_EXIT_CODE=$?

kill $SEDRAD_PID

wait $SEDRAD_PID
SEDRAD_EXIT_CODE=$?

echo "Exit code: $TEST_EXIT_CODE"
echo "sedrad exit code: $SEDRAD_EXIT_CODE"

if [ $TEST_EXIT_CODE -eq 0 ] && [ $SEDRAD_EXIT_CODE -eq 0 ]; then
  echo "orphans test: PASSED"
  exit 0
fi
echo "orphans test: FAILED"
exit 1
