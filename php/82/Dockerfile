# syntax = edrevo/dockerfile-plus
FROM php:8.2-fpm

ENV PATH /usr/local/go/bin:/root/.composer/vendor/bin:$PATH

RUN apt update && apt install -y --no-install-recommends \
    cron \
    git \
    gnupg \
    iputils-ping \
    less \
    libfreetype6-dev \
    libicu-dev \
    libjpeg62-turbo-dev \
    libmcrypt-dev \
    libmemcached-dev \
    libpng-dev \
    libwebp-dev \
    libxml2-dev \
    libxslt-dev \
    nano \
    net-tools \
    ssh \
    msmtp \
    sudo \
    supervisor \
    unzip \
    libzip-dev \
    libz-dev \
    zsh \
    fonts-powerline \
    powerline \
  && yes '' | pecl install -f redis xdebug mcrypt \
  && docker-php-ext-install -j$(nproc) bcmath calendar exif intl mysqli opcache pdo_mysql soap xsl zip\
  && docker-php-ext-enable mcrypt redis xdebug\
  && apt clean \
  && rm -rf /var/lib/apt/lists/*

ADD https://github.com/mlocati/docker-php-extension-installer/releases/latest/download/install-php-extensions /usr/local/bin/
RUN chmod +x /usr/local/bin/install-php-extensions && sync; \
  install-php-extensions ldap; 

RUN apt update && apt install -y zlib1g-dev libssl-dev memcached && install-php-extensions php-memcached-dev/php-memcached@master

# Configure GD Library
RUN docker-php-ext-configure gd --with-freetype --with-jpeg --with-webp \
  && docker-php-ext-install -j$(nproc) gd

# Install Composer
RUN curl --silent --show-error https://getcomposer.org/installer | php \
  && mv composer.phar /usr/local/bin/composer \
  && composer clear-cache

# Install MailHog
RUN curl -L -o /usr/bin/mhsendmail https://github.com/mailhog/mhsendmail/releases/download/v0.2.0/mhsendmail_linux_amd64 \
  && chmod +x /usr/bin/mhsendmail

# Setup Crons
RUN touch /root/custom_crontab \
  && /usr/bin/crontab -u www-data /root/custom_crontab

# Update site's directory permissions
RUN chown -R www-data /var/www/

# Custom PHP.ini Config
COPY ./conf/custom.ini /usr/local/etc/php/conf.d/zza-custom.ini

# PHP-FPM and CRON supervisord config
COPY ./conf/supervisord.conf /etc/supervisor/supervisord.conf
RUN touch /etc/supervisor/supervisord.pid

ARG PHPV=82

INCLUDE+ ./common/shell.Dockerfile
INCLUDE+ ./common/xhprof.Dockerfile
COPY ./conf/global.php /home/dev/global.php

CMD ["sudo", "php-fpm", "-g", "/var/run/php-fpm.pid"]
