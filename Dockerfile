FROM gcr.io/distroless/static:nonroot
WORKDIR /

ARG TARGETARCH
COPY dist/vitistack-operator-${TARGETARCH} /bin/vitistack-operator
USER 10000:10000
EXPOSE 8888

ENTRYPOINT ["/bin/vitistack-operator"]