FROM scratch
COPY vault-utils /
ENTRYPOINT ["/vault-utils"]
