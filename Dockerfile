FROM scratch

ARG TARGETARCH

COPY dist/fyaml_linux_${TARGETARCH}_*/fyaml /fyaml

ENTRYPOINT ["/fyaml"]
