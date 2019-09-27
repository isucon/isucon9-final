FROM php:7-apache

RUN apt-get update && apt-get install -y git unzip && apt-get clean


ENV COMPOSER_ALLOW_SUPERUSER 1
ENV COMPOSER_NO_INTERACTION 1

ADD php.ini /usr/local/etc/php/

RUN docker-php-ext-install pdo_mysql mysqli mbstring iconv

# install composer
RUN curl -sS https://getcomposer.org/installer | php -- --install-dir=/usr/local/bin --filename=composer

RUN a2enmod rewrite

RUN sed -i 's/Listen 80/Listen 8000/' /etc/apache2/ports.conf
RUN sed -i 's/:80/:8000/' /etc/apache2/sites-available/000-default.conf
RUN sed -i 's_/var/www/html_/var/www/html/public_' /etc/apache2/sites-available/000-default.conf

ADD entrypoint.sh /var/www/html

CMD ["bash", "-xe", "entrypoint.sh"]
