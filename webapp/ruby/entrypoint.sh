#!/bin/bash

bundle install --path vendor
bundle exec rackup config.ru -o 0.0.0.0 -p 8000
