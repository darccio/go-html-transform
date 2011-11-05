# html/Transform
TRANS_SRCDIR := html/transform
TRANS_GOSRCS := $(shell ls ${TRANS_SRCDIR}/*.go)
TRANS_GOTESTSRCS := $(shell ls ${TRANS_SRCDIR}/*_test.go)

#h5
H5_SRCDIR := h5
H5_GOSRCS := $(shell ls ${H5_SRCDIR}/*.go)
H5_GOTESTSRCS := $(shell ls ${H5_SRCDIR}/*_test.go)

#format for both
GOFMTARGS := ${TRANS_GOSRCS:%.go=%.fmt}
GOFMTARGS += ${H5_GOSRCS:%.go=%.fmt}

default: trans

test: h5test transtest

install: transinstall

clean: h5clean transclean

# h5
h5:
	(cd ${H5_SRCDIR} && gomake)

h5test:
	(cd ${H5_SRCDIR} && gotest -test.v)

h5install: h5
	(cd ${H5_SRCDIR} && make install)

h5clean:
	(cd ${H5_SRCDIR} && make clean)

# html/trans
trans: h5install
	(cd ${TRANS_SRCDIR} && gomake)

transtest: h5install
	(cd ${TRANS_SRCDIR} && gotest -test.v)

transinstall: h5install
	(cd ${TRANS_SRCDIR} && make install)

transclean:
	(cd ${TRANS_SRCDIR} && make clean)

${TRANS_SRCDIR}/%.fmt: ${SRCDIR}/%.go
	gofmt -spaces -w $<

#common to both
format: ${GOFMTARGS}
