SRCDIR := pkg/html/transform
GOSRCS := $(shell ls ${SRCDIR}/*.go)
GOTESTSRCS := $(shell ls ${SRCDIR}/*_test.go)
GOFMTARGS := ${GOSRCS:%.go=%.fmt}

default:
	(cd ${SRCDIR} && gomake)

test:
	(cd ${SRCDIR} && gotest -test.v)

install:
	(cd ${SRCDIR} && make install)

clean:
	(cd ${SRCDIR} && make clean)

format: ${GOFMTARGS}

${SRCDIR}/%.fmt: ${SRCDIR}/%.go
	gofmt -spaces -w $<
