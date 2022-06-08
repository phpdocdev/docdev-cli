USER dev

COPY ./conf/xhprof-5.patch /tmp/xhprof.patch

RUN if [ $(php -r 'echo PHP_MAJOR_VERSION;') -gt 5 ]; then \
    cd /tmp \
    && git clone https://github.com/tideways/php-xhprof-extension.git \
    && sudo docker-php-ext-install /tmp/php-xhprof-extension/; \
    else \
    cd /tmp \
    && sudo curl -s -o xhprof-0.9.4.tgz -b a -L http://pecl.php.net/get/xhprof-0.9.4.tgz \
    && tar xvf xhprof-0.9.4.tgz \
    && git apply /tmp/xhprof.patch \
    && sudo docker-php-ext-install /tmp/xhprof-0.9.4/extension/ \
    && sudo rm -rf /tmp/xhprof-0.9.4; \
  fi

RUN sudo install-php-extensions mongodb \
  && composer global require perftools/php-profiler alcaeus/mongo-php-adapter perftools/xhgui-collector  \
  && git clone https://github.com/perftools/xhgui /home/dev/xhprof \
  && cd /home/dev/xhprof \
  && chmod -R 0777 cache \
  && composer install --no-dev \
  && ln -s webroot/ public

COPY ./conf/xhprof.php /home/dev/xhprof/config/config.default.php