# syntax = edrevo/dockerfile-plus
FROM php:7.4-fpm

ENV PATH /usr/local/go/bin:/root/.composer/vendor/bin:$PATH

ADD https://github.com/mlocati/docker-php-extension-installer/releases/latest/download/install-php-extensions /usr/local/bin/

RUN apt update && apt install -y --no-install-recommends \
  nano \
  net-tools \
  ssh \
  msmtp \
  sudo \
  supervisor \
  unzip \
  zsh \
  fonts-powerline \
  powerline

RUN set -eux; \
	\
  savedAptMark="$(apt-mark showmanual)"; \
  apt update && apt install -y --no-install-recommends \
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
    libzip-dev \
    libz-dev \
    ; \
  yes '' | pecl install -f memcached redis xdebug mcrypt; \
  docker-php-ext-install -j$(nproc) bcmath calendar exif intl mysqli opcache pdo_mysql soap xsl zip; \
  docker-php-ext-enable mcrypt memcached redis xdebug; \
  docker-php-ext-configure gd --with-freetype --with-jpeg --with-webp; \
  docker-php-ext-install -j$(nproc) gd; \
  chmod +x /usr/local/bin/install-php-extensions && sync; \
  install-php-extensions ldap ioncube_loader; \
  \
	cd /; \
	docker-php-source delete; \
  \
  apt clean; \
  apt-mark auto '.*' > /dev/null; \
  [ -z "$savedAptMark" ] || apt-mark manual $savedAptMark; \
	find /usr/local -type f -executable -exec ldd '{}' ';' \
		| awk '/=>/ { print $(NF-1) }' \
		| sort -u \
		| xargs -r dpkg-query --search \
		| cut -d: -f1 \
		| sort -u \
		| xargs -r apt-mark manual \
	; \
  apt-get purge -y --auto-remove -o APT::AutoRemove::RecommendsImportant=false; \
  rm -rf /var/lib/apt/lists/*

# Configure GD Library

# Install Composer
RUN curl --silent --show-error https://getcomposer.org/installer | php \
  && mv composer.phar /usr/local/bin/composer \
  && composer clear-cache

# Install MailHog
RUN curl -L -o /usr/bin/mhsendmail https://github.com/mailhog/mhsendmail/releases/download/v0.2.0/mhsendmail_linux_amd64 \
  && chmod +x /usr/bin/mhsendmail

# Setup Crons
# RUN touch /root/custom_crontab \
  # && /usr/bin/crontab -u www-data /root/custom_crontab

# Update site's directory permissions
RUN chown -R www-data /var/www/

# Custom PHP.ini Config
COPY ./conf/custom.ini /usr/local/etc/php/conf.d/zza-custom.ini

# PHP-FPM and CRON supervisord config
COPY ./conf/supervisord.conf /etc/supervisor/supervisord.conf
RUN touch /etc/supervisor/supervisord.pid

ARG PHPV=74

INCLUDE+ ./common/shell.Dockerfile
INCLUDE+ ./common/xhprof.Dockerfile
COPY ./conf/global.php /home/dev/global.php

CMD ["sudo", "php-fpm", "-g", "/var/run/php-fpm.pid"]
