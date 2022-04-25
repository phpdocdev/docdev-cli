# syntax = edrevo/dockerfile-plus
#FROM php:5.4-fpm
INCLUDE+ ./54-alpine/build/Dockerfile

ENV PATH /usr/local/go/bin:/root/.composer/vendor/bin:$PATH

# RUN apt update && apt install -y --no-install-recommends \
#     cron \
#     git \
#     gnupg \
#     iputils-ping \
#     less \
#     libfreetype6-dev \
#     libicu-dev \
#     libjpeg62-turbo-dev \
#     libmcrypt-dev \
#     libmemcached-dev \
#     libpng-dev \
#     libvpx-dev \
#     libxml2-dev \
#     libxslt-dev \
#     nano \
#     net-tools \
#     ssh \
#     ssmtp \
#     sudo \
#     supervisor \
#     unzip \
#     zip

RUN yes '' | pecl install -f memcached-2.2.0 redis-2.2.8


RUN apk add --no-cache --virtual .build-deps icu-dev libmcrypt libmcrypt-dev libxml2-dev libxslt-dev zip unzip
RUN docker-php-ext-install bcmath calendar mysql intl mcrypt mysqli mysql pdo_mysql soap xsl zip

RUN apk add --no-cache --virtual .build-deps jpeg-dev libvpx-dev libpng-dev freetype-dev gd-dev libwebp-dev
# RUN pecl install memcached-2.2.0
RUN docker-php-ext-install gd

# Configure GD Library
# RUN docker-php-ext-configure gd --with-freetype-dir=/usr/include/ --with-jpeg-dir=/usr/include/ --with-vpx-dir=/usr/include/ \
  # && docker-php-ext-install gd

# Install MailHog
RUN curl -L -o /usr/bin/mhsendmail https://github.com/mailhog/mhsendmail/releases/download/v0.2.0/mhsendmail_linux_amd64 \
  && chmod +x /usr/bin/mhsendmail

RUN apk add --no-cache --virtual .build-deps sudo
# Setup Crons

# Update site's directory permissions
RUN chown -R www-data /var/www/

# Custom PHP.ini Config
COPY ./conf/custom.ini /usr/local/etc/php/conf.d/zza-custom.ini

# PHP-FPM and CRON supervisord config
COPY ./conf/supervisord.conf /etc/supervisor/supervisord.conf
RUN touch /etc/supervisor/supervisord.pid

# RUN pecl install xdebug-2.4.1
# RUN docker-php-ext-install xdebug
RUN apk add --no-cache $PHPIZE_DEPS \
    && pecl install xdebug-2.4.1 \
    && docker-php-ext-enable xdebug


RUN apk add --no-cache --virtual .build-deps ldb-dev libldap openldap-dev \
  && docker-php-ext-install ldap

# RUN apt install libldb-dev libldap2-dev -y \
    # && ln -fs /usr/lib/x86_64-linux-gnu/libldap.so /usr/lib/ \
    # && docker-php-ext-install ldap \
    # && docker-php-ext-enable ldap

# RUN curl -o ioncube.tar.gz http://downloads3.ioncube.com/loader_downloads/ioncube_loaders_lin_x86-64.tar.gz \
#     && tar -xvvzf ioncube.tar.gz \
#     && mv ioncube/ioncube_loader_lin_$(php -r 'echo PHP_MAJOR_VERSION . "." . PHP_MINOR_VERSION;').so `php-config --extension-dir` \
#     && rm -Rf ioncube.tar.gz ioncube \
#     && docker-php-ext-enable ioncube_loader_lin_$(php -r 'echo PHP_MAJOR_VERSION . "." . PHP_MINOR_VERSION;')

COPY ./54-alpine/build/docker-php-ext-* ./54-alpine/build/docker-php-entrypoint /usr/local/bin/

ADD https://github.com/mlocati/docker-php-extension-installer/releases/latest/download/install-php-extensions /usr/local/bin/
RUN chmod +x /usr/local/bin/install-php-extensions && sync 

ARG PHPV=54
INCLUDE+ ./common/alpine.Dockerfile

RUN rm -rf /var/lib/apt/lists/*

RUN apk add --no-cache --virtual .build-deps supervisor
COPY ./conf/supervisord.conf /etc/supervisord.conf
RUN touch /etc/supervisor/supervisord.pid

CMD ["/usr/bin/supervisord"]
