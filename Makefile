default:
	(cd pkg/html/transform && gomake)

test:
	(cd pkg/html/transform && gotest)

benchmark:
	(cd pkg/html/transform && gotest -benchmarks=".*")
