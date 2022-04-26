# syntax = edrevo/dockerfile-plus
FROM php:5.6-fpm-alpine

LABEL org.opencontainers.image.source "https://github.com/phpdocdev/docdev-cli"
LABEL org.opencontainers.image.description "Alpine PHP 5.6 FPM"

ENV PATH /usr/local/go/bin:/root/.composer/vendor/bin:$PATH

ENV PHPIZE_DEPS \
		autoconf \
		dpkg-dev dpkg \
		file \
		g++ \
		gcc \
		libc-dev \
		make \
		pkgconf \
		re2c

ADD https://github.com/mlocati/docker-php-extension-installer/releases/latest/download/install-php-extensions /usr/local/bin/

RUN apk add --no-cache --virtual build-deps \
		$PHPIZE_DEPS \
		coreutils \
		curl-dev \
		libedit-dev \
		libressl-dev \
		libxml2-dev \
		sqlite-dev \
        jpeg-dev \
        libvpx-dev \
        libpng-dev \
        freetype-dev \
        gd-dev \
        libwebp-dev \
        && apk add --no-cache --virtual build-deps icu-dev libmcrypt libmcrypt-dev libxml2-dev libxslt-dev zip unzip \
        && yes '' | pecl install -f redis-4.3.0 xdebug-2.5.5 \
        && docker-php-ext-install -j$(nproc) bcmath calendar mysql intl mcrypt mysqli opcache pdo_mysql soap xsl zip gd \
        && docker-php-ext-enable mcrypt xdebug \
        && chmod +x /usr/local/bin/install-php-extensions && sync && \
        install-php-extensions ldap ioncube_loader \
        && apk del --no-network build-deps; \
            cd /; \
            docker-php-source delete;

# Install Composer
RUN curl --silent --show-error https://getcomposer.org/installer | php \
    && mv composer.phar /usr/local/bin/composer \
    && composer clear-cache \
    && curl -L -o /usr/bin/mhsendmail https://github.com/mailhog/mhsendmail/releases/download/v0.2.0/mhsendmail_linux_amd64 \
    && chmod +x /usr/bin/mhsendmail \
    && chown -R www-data /var/www/

# Custom PHP.ini Config
COPY ./conf/custom.ini /usr/local/etc/php/conf.d/zza-custom.ini

# PHP-FPM and CRON supervisord config
COPY ./conf/supervisord.conf /etc/supervisor/supervisord.conf
RUN touch /etc/supervisor/supervisord.pid

COPY ./conf/supervisord.conf /etc/supervisord.conf
RUN apk add --no-cache supervisor \
  && touch /etc/supervisor/supervisord.pid
    
ARG PHPV=56
INCLUDE+ ./common/alpine.Dockerfile

CMD ["/usr/bin/supervisord"]
