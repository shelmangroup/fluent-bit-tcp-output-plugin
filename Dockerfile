FROM golang:1.8 as builder
ADD . /go/src/github.com/shelmangroup/fluent-bit-tcp-output-plugin
WORKDIR /go/src/github.com/shelmangroup/fluent-bit-tcp-output-plugin
RUN make && mv out_tcp.so /

FROM fluent/fluent-bit:0.12-dev
COPY --from=builder /out_tcp.so /
ADD fluent-bit.conf /fluent-bit/etc/fluent-bit.conf
ENV TCP_OUTPUT_HOST localhost:5170
CMD ["/fluent-bit/bin/fluent-bit", "-c", "/fluent-bit/etc/fluent-bit.conf", "-e", "/out_tcp.so"]
