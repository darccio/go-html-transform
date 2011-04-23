default:
	(cd pkg/html/transform && gomake)

test:
	(cd pkg/html/transform && gotest -test.v)

benchmark:
	(cd pkg/html/transform && gotest -benchmarks=".*")

install:
	(cd pkg/html/transform && make install)
