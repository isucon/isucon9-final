FROM node:8.17.0

RUN npm install -g npm @vue/cli @vue/cli-service-global


WORKDIR /opt/frontend
ENV HOST=0.0.0.0

CMD ["bash", "-c", "npm install && npm run serve"]
