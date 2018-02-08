FROM golang:1.9
WORKDIR /go/src/app
COPY getSecretFromVault .
RUN chmod +x getSecretFromVault
CMD ["./getSecretFromVault"]