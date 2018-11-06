#!/bin/sh
#
# This file is used for running Proof-of-Work benchmarks, generating
# histograms and calculating average time to find the proof.
#
# Copyright (c) 2018 Vitaly Minko <vitaly.minko@gmail.com>

COUNT=10
BIT_NUM=16
TEST_TYPE=BenchmarkPoW
DATAFILE=histogram.dat
IMGFILE=pow_histogram.png
DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

if [ $# -gt 2 ] || [ "$1" == "-h" ]
then
    cat >&2 << EOF
Usage:
    $0 [bit_num] [count]
where,
    [bit_num] is either 8 or 16. Default value is ${BIT_NUM}.
    [count] specifies how many times to run the test. Default value is ${COUNT}.
EOF
    exit 1
fi

if [ -n "$1" ]; then
	BIT_NUM=$1
fi
TEST_NAME="${TEST_TYPE}${BIT_NUM}"

if [ -n "$2" ]; then
	COUNT=$2
fi

echo "Starting benchmark. Use 'tail -f benchmark.log' to monitor the progress."
go test -bench=$TEST_NAME -benchtime=1ns -timeout=24h -count=$COUNT > benchmark.log
echo "The benchmark is over."
cat benchmark.log | grep $TEST_NAME | awk '{print $3/1E9}' > $DATAFILE

echo -n "The average time to find proof is: "
awk '{ total += $1; count++ } END { print total/count }' $DATAFILE

gnuplot -e "infile='$DATAFILE'" \
        -e "bitnum=$BIT_NUM" \
	-e "count=$COUNT" \
	-e "outfile='$IMGFILE'" \
	$DIR/pow_histogram.gp
echo "The histogram is ready: ${IMGFILE}."
