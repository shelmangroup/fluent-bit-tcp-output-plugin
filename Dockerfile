FROM golang:1.8 as builder
ADD . /go/src/github.com/shelmangroup/fluent-bit-tcp-output-plugin
WORKDIR /go/src/github.com/shelmangroup/fluent-bit-tcp-output-plugin
RUN make && mv out_tcp.so /

FROM fluent/fluent-bit:0.12-dev
COPY --from=builder /out_tcp.so /
ENV TCP_OUTPUT_HOST localhost:5170
CMD ["/fluent-bit/bin/fluent-bit", "-e", "/out_tcp.so", "-c", "/fluent-bit/etc/fluent-bit.conf", "-o", "out_tcp"]
