#!/bin/bash

composer install
docker-php-entrypoint apache2-foreground
