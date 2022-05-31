USER dev

RUN if [ $(php -r 'echo PHP_MAJOR_VERSION;') -gt 5 ]; then \
    cd /tmp \
    && git clone https://github.com/tideways/php-xhprof-extension.git \
    && sudo docker-php-ext-install /tmp/php-xhprof-extension/; \
    else \
    sudo pecl install xhprof-0.9.4; \
    fi

RUN sudo install-php-extensions mongodb \
  && composer global require perftools/php-profiler alcaeus/mongo-php-adapter perftools/xhgui-collector  \
  && git clone https://github.com/perftools/xhgui /home/dev/xhprof \
  && cd /home/dev/xhprof \
  && chmod -R 0777 cache \
  && composer install --no-dev \
  && ln -s webroot/ public

COPY ./conf/xhprof.php /home/dev/xhprof/config/config.default.php