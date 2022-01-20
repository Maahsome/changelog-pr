#####################
#   Download stage  #
#####################
FROM alpine:3 as download

ARG VERSION="v0.0.1"
ENV DL_URL="https://github.com/Maahsome/changelog-pr/releases/download/${VERSION}/changelog-pr_linux_amd64.tar.gz"

RUN apk add --no-cache \
      curl \
      bash

SHELL ["/bin/bash", "-o", "pipefail", "-c"]
RUN curl -Ls "${DL_URL}" | tar -zxv --strip-components=1 -C /usr/local/bin/ changelog-pr_linux_amd64/changelog-pr && \
    chmod +x /usr/local/bin/changelog-pr

#####################
#  Finalized image  #
#####################
FROM alpine:3

RUN apk add --no-cache \
      curl \
      git \
      git-lfs \
      jq

COPY --from=download /usr/local/bin/* /usr/local/bin/
