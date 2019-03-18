FROM golang:1.12 AS build
ENV GOPATH=""
ARG LDFLAGS
COPY go.sum .
COPY go.mod .
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o helm-convert -ldflags "$LDFLAGS" main.go

FROM alpine AS helm
ENV HELM_VERSION=v2.13.0
ENV HELM_TMP_FILE=helm-${HELM_VERSION}-linux-amd64.tar.gz
RUN wget https://storage.googleapis.com/kubernetes-helm/${HELM_TMP_FILE} && \
  wget https://storage.googleapis.com/kubernetes-helm/${HELM_TMP_FILE}.sha256
RUN apk --no-cache add openssl
RUN if [ "$(openssl sha1 -sha256 ${HELM_TMP_FILE} | awk '{print $2}')" != "$(cat helm-${HELM_VERSION}-linux-amd64.tar.gz.sha256)" ]; \
  then \
    echo "SHA sum of ${HELM_TMP_FILE} does not match. Aborting."; \
    exit 1; \
  fi
RUN tar -xvf helm-${HELM_VERSION}-linux-amd64.tar.gz

FROM alpine:3.8
COPY --from=helm /linux-amd64/helm /usr/local/bin/helm
RUN helm init --skip-refresh --client-only && \
  mkdir -p /root/.helm/plugins/helm-convert
COPY plugin.yaml /root/.helm/plugins/helm-convert/plugin.yaml
COPY --from=build /go/helm-convert /root/.helm/plugins/helm-convert/helm-convert
ENTRYPOINT ["helm"]
