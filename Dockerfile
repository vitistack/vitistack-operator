FROM ncr.sky.nhn.no/gcr/distroless/static:nonroot
WORKDIR /

COPY dist/datacenter-operator /bin/datacenter-operator
USER 10000:10000
EXPOSE 8888

ENTRYPOINT ["/bin/datacenter-operator"]