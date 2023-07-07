#!/usr/bin/env bash

# exected after build

# startup pgw for tests
pkill pgw
./out/bin/pgw 2>/dev/null > /dev/null &

function cleanup() {
    echo [INFO] pkill pgw
    pkill pgw
}
trap cleanup EXIT

function run_xk6gtp() {
    local jsfile=$1
    ./out/bin/xk6gtp run $jsfile 2> /dev/null
}

# execute test scenarios
for jsfile in example/*.js; do
    echo "run $jsfile"
    res=$(run_xk6gtp $jsfile |grep 'checks'|awk '{print $2}')
    echo "result: $res"
    if [ "$res" != "100.00%" ]; then
        failed=1
    fi
done

# reporting termination
if [ -v failed ]; then
    exit 1
fi
echo OK