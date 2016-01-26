#!/bin/sh
here=$(readlink -e "$(dirname "$0")")
export GOPATH="$here/../../"
echo ""
echo "==> Part I"
go test -run Sequential mapreduce/...
echo ""
echo "==> Part II"
(cd "$here" && ./test-wc.sh > /dev/null)
echo ""
echo "==> Part III"
go test -run TestBasic mapreduce/...
echo ""
echo "==> Part IV"
go test -run Failure mapreduce/...
echo ""
echo "==> Part V (challenge)"
(cd "$here" && ./test-ii.sh > /dev/null)

rm "$here"/mrtmp.* "$here"/diff.out
