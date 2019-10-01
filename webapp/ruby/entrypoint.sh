#!/bin/bash

bundle install --path vendor/bundle
bundle exec puma -b tcp://0.0.0.0:8000
