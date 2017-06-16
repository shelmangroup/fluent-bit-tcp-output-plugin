all:
	go get . && \
		go build -buildmode=c-shared -ldflags="-s -w" -o out_tcp.so out_tcp.go

clean:
	rm -rf *.so *.h *~

docker:
	docker build -t shelmangroup/fluent-bit-tcp:$(IMAGE_TAG) .

push:
	docker push shelmangroup/fluent-bit-tcp:$(IMAGE_TAG)
