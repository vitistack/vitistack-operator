FROM ncr.sky.nhn.no/gcr/distroless/static:nonroot
WORKDIR /

COPY dist/dc-operator /bin/dc-operator
USER 10000:10000
EXPOSE 8888

ENTRYPOINT ["/bin/dc-operator"]