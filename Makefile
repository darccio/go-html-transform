# defines $GC) (compiler), $(LD) (linker) and $(O) (architecture)
include $(GOROOT)/src/Make.inc

TARG=html/parse
GOFILES=\
	h5.go\

include $(GOROOT)/src/Make.pkg
