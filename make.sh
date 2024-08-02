#!/bin/sh

PROJECT=masters

VERBOSITY=0
VERBOSITYFLAGS=""
while test "$1" = "-v"; do
	VERBOSITY=$((VERBOSITY+1))
	VERBOSITYFLAGS="$VERBOSITYFLAGS -v"
	shift
done

run()
{
	if test $VERBOSITY -gt 1; then echo "$@"; fi
	"$@" || exit 1
}

printv()
{
	if test $VERBOSITY -gt 0; then echo "$@"; fi
}

# NOTE(anton2920): disable Go 1.11+ package management.
GO111MODULE=off; export GO111MODULE
GOPATH=`go env GOPATH`; export GOPATH

CGO_ENABLED=0; export CGO_ENABLED

STARTTIME=`date +%s`

case $1 in
	'' | debug)
		CGO_ENABLED=1; export CGO_ENABLED
		run go build -o $PROJECT -race -pgo off -gcflags='all=-N -l -d=checkptr=0' -ldflags='-X main.BuildMode=Debug' -tags gofadebug
		;;
	clean)
		run rm -f $PROJECT $PROJECT.s $PROJECT.esc $PROJECT.test c.out cpu.pprof cpu.png mem.pprof mem.png
		run go clean -cache -modcache -testcache
		run rm -rf `go env GOCACHE`
		run rm -rf /tmp/cover*
		;;
	check)
		run $0 $VERBOSITYFLAGS test-race-cover
		run ./$PROJECT.test
		;;
	check-bench)
		run $0 $VERBOSITYFLAGS test
		run ./$PROJECT.test -test.bench=. -test.benchmem -test.run=^Benchmark
		;;
	check-bench-cpu)
		run $0 $VERBOSITYFLAGS test
		run ./$PROJECT.test -test.bench=. -test.benchmem -test.run=^Benchmark -test.cpuprofile=cpu.pprof
		;;
	check-bench-mem)
		run $0 $VERBOSITYFLAGS test
		run ./$PROJECT.test -test.bench=. -test.benchmem -test.run=^Benchmark -test.memprofile=mem.pprof
		;;
	check-cover)
		run $0 $VERBOSITYFLAGS test-race-cover
		run ./$PROJECT.test -test.coverprofile=c.out
		run go tool cover -html=c.out
		run rm -f c.out
		;;
	gofa/prof)
		run go build -o $PROJECT -ldflags="-s -w -X main.BuildMode=gofa/prof" -tags gofaprof
		;;
	disas | disasm | disassembly)
		printv go build -gcflags="-S"
		go build -o $PROJECT -gcflags="-S" >$PROJECT.s 2>&1
		;;
	esc | escape | escape-analysis)
		printv go build -gcflags="-m -m"
		go build -o $PROJECT -gcflags="-m -m" >$PROJECT.m 2>&1
		;;
	fmt)
		if which goimports >/dev/null; then
			run goimports -l -w *.go
		else
			run gofmt -l -s -w *.go
		fi
		;;
	objdump)
		go build -o $PROJECT
		printv go tool objdump -S -s ^main\. $PROJECT
		go tool objdump -S $PROJECT >$PROJECT.s
		;;
	png)
		printv go tool pprof -png masters-cpu.pprof
		go tool pprof -png masters-cpu.pprof >cpu.png
		;;
	profiling)
		run go build -o $PROJECT -ldflags="-s -w -X main.BuildMode=Profiling"
		;;
	release)
		run go build -o $PROJECT -gcflags="-d=checkptr=0" -ldflags="-s -w"
		;;
	test)
		run $0 $VERBOSITYFLAGS vet
		run go test $VERBOSITYFLAGS -c -o $PROJECT.test -vet=off
		;;
	test-race-cover)
		CGO_ENABLED=1; export CGO_ENABLED
		run $0 $VERBOSITYFLAGS vet
		run go test $VERBOSITYFLAGS -c -o $PROJECT.test -vet=off -race -cover -gcflags='all=-N -l'
		;;
	tracing)
		run go build -o $PROJECT -ldflags="-s -w -X main.BuildMode=Tracing"
		;;
	vet)
		run go vet $VERBOSITYFLAGS
		;;
	*)
		echo "Target $1 is not supported"
		;;
esac

ENDTIME=`date +%s`

echo Done $1 in $((ENDTIME-STARTTIME))s
