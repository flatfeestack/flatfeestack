FROM stripe/stripe-cli
COPY ./config.toml /root/.config/stripe/config.toml
ENTRYPOINT ["/bin/stripe", "listen", "--skip-verify", "--forward-to", "http://backend:9082"]