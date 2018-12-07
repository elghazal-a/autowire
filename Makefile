

dep-import:
	docker build -t geniousphp/autowire .
	docker run --rm -i -v $(pwd):/go/src/github.com/geniousphp/autowire geniousphp/autowire dep ensure -add $(m)