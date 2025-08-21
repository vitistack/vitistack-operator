FROM ncr.sky.nhn.no/gcr/distroless/static:nonroot
WORKDIR /

COPY dist/vitistack-operator /bin/vitistack-operator
USER 10000:10000
EXPOSE 8888

ENTRYPOINT ["/bin/vitistack-operator"]