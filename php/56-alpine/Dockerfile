# syntax = edrevo/dockerfile-plus
FROM php:5.6-fpm-alpine

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

RUN apk add --no-cache --virtual build-deps \
		$PHPIZE_DEPS \
		coreutils \
		curl-dev \
		libedit-dev \
		libressl-dev \
		libxml2-dev \
		sqlite-dev

RUN apk add --no-cache --virtual build-deps icu-dev libmcrypt libmcrypt-dev libxml2-dev libxslt-dev zip unzip \
    && yes '' | pecl install -f redis-4.3.0 xdebug-2.5.5 \
  && docker-php-ext-install -j$(nproc) bcmath calendar mysql intl mcrypt mysqli opcache pdo_mysql soap xsl zip\
  && docker-php-ext-enable mcrypt xdebug

# Configure GD Library
RUN apk add --no-cache --virtual build-deps jpeg-dev libvpx-dev libpng-dev freetype-dev gd-dev libwebp-dev
RUN docker-php-ext-install gd

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

ADD https://github.com/mlocati/docker-php-extension-installer/releases/latest/download/install-php-extensions /usr/local/bin/
RUN chmod +x /usr/local/bin/install-php-extensions && sync && \
    install-php-extensions ldap ioncube_loader

RUN apk del --no-network build-deps; \
    cd /; \
	docker-php-source delete; 
    
ARG PHPV=56
INCLUDE+ ./common/alpine.Dockerfile

RUN apk add sudo supervisor \
    && touch /etc/supervisor/supervisord.pid
COPY ./conf/supervisord.conf /etc/supervisord.conf

CMD ["/usr/bin/supervisord"]
