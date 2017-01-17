FROM appscode/base:8.6

# Install strongswan and etcdctl
RUN set -x \
  && apt-get update \
  && apt-get install -y --no-install-recommends \
    ca-certificates \
    wget \
    strongswan \
    strongswan-ikev2 \
  && apt-get purge -y --auto-remove wget \
  && rm -rf /var/lib/apt/lists/* /usr/share/doc /usr/share/man /tmp/*

# Install confd
COPY swanc /usr/bin/swanc

# Setup runit scripts
COPY sv /etc/sv/
RUN ln -s /etc/sv /etc/service

ENTRYPOINT ["/runit.sh"]
EXPOSE 4500/udp 500/udp
