FROM ruby:2.7.1

ENV TZ "Asia/Tokyo"
ENV BUNDLER_VERSION 1.17.2

RUN mkdir -p /opt/webapp
WORKDIR /opt/webapp

RUN gem install bundler --version "$BUNDLER_VERSION"

CMD ["bash", "-x", "entrypoint.sh"]
