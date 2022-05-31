USER dev
RUN cd /tmp \
  && git clone https://github.com/tideways/php-xhprof-extension.git \
  && sudo docker-php-ext-install /tmp/php-xhprof-extension/ \
  && sudo install-php-extensions mongodb \
  && composer global require perftools/php-profiler alcaeus/mongo-php-adapter perftools/xhgui-collector  \
  && git clone https://github.com/perftools/xhgui /home/dev/xhprof \
  && cd /home/dev/xhprof \
  && chmod -R 0777 cache \
  && composer install --no-dev \
  && ln -s webroot/ public

COPY ./conf/xhprof.php /home/dev/xhprof/config/config.default.php