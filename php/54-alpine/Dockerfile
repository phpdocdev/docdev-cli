# syntax = edrevo/dockerfile-plus
INCLUDE+ ./54-alpine/build/Dockerfile

ENV PATH /usr/local/go/bin:/root/.composer/vendor/bin:$PATH

ADD https://github.com/mlocati/docker-php-extension-installer/releases/latest/download/install-php-extensions /usr/local/bin/

RUN apk add --no-cache --virtual build-deps $PHPIZE_DEPS icu-dev libmcrypt libmcrypt-dev libxml2-dev libxslt-dev zip unzip \
  && yes '' | pecl install -f memcached-2.2.0 redis-2.2.8 xdebug-2.4.1 \
  && docker-php-ext-install bcmath calendar mysql intl mcrypt mysqli mysql pdo_mysql soap xsl zip \
  && apk add --no-cache --virtual build-deps jpeg-dev libvpx-dev libpng-dev freetype-dev gd-dev libwebp-dev \
  && docker-php-ext-install gd \
  && docker-php-ext-enable xdebug \
  && apk add --no-cache --virtual build-deps ldb-dev libldap openldap-dev \
  && docker-php-ext-install ldap \
  && chmod +x /usr/local/bin/install-php-extensions && sync \
  && apk del build-deps; \
    cd /; \
    docker-php-source delete; 

# Install MailHog
RUN curl -L -o /usr/bin/mhsendmail https://github.com/mailhog/mhsendmail/releases/download/v0.2.0/mhsendmail_linux_amd64 \
  && chmod +x /usr/bin/mhsendmail \
  && apk add --no-cache --virtual build-deps sudo \
  && chown -R www-data /var/www/

# Custom PHP.ini Config
COPY ./conf/custom.ini /usr/local/etc/php/conf.d/zza-custom.ini
COPY ./conf/supervisord.conf /etc/supervisor/supervisord.conf
COPY ./54-alpine/build/docker-php-ext-* ./54-alpine/build/docker-php-entrypoint /usr/local/bin/
COPY ./conf/supervisord.conf /etc/supervisord.conf
RUN apk add --no-cache supervisor \
  && touch /etc/supervisor/supervisord.pid

ARG PHPV=54
INCLUDE+ ./common/alpine.Dockerfile

CMD ["/usr/bin/supervisord"]
