# Setup user (to stay in-sync with the host system's user)
ARG CUSTOM_USER_NAME=dev
ENV CUSTOM_USER_NAME=$CUSTOM_USER_NAME

ARG CUSTOM_UID=1000
ENV CUSTOM_UID=$CUSTOM_UID

ARG CUSTOM_GID=1000
ENV CUSTOM_GID=$CUSTOM_GID

ARG PHPV
ENV PHPV=$PHPV

# RUN groupadd -g ${CUSTOM_GID} ${CUSTOM_USER_NAME} \
  # && useradd -m -u ${CUSTOM_UID} -g ${CUSTOM_USER_NAME} -G sudo -s /usr/bin/zsh ${CUSTOM_USER_NAME} \
  # && passwd -d ${CUSTOM_USER_NAME} \
  # && echo "${CUSTOM_USER_NAME} ALL=(ALL) NOPASSWD:ALL" >> /etc/sudoers

RUN set -x \
	&& addgroup -g ${CUSTOM_UID} -S ${CUSTOM_USER_NAME} \
	&& adduser -u ${CUSTOM_UID} -D -S -G ${CUSTOM_USER_NAME} ${CUSTOM_USER_NAME} \
  && echo "${CUSTOM_GID} ALL=(ALL) NOPASSWD:ALL" >> /etc/sudoers

# Aliases
RUN echo "alias artisan='php artisan'" >> /home/${CUSTOM_USER_NAME}/.zshrc \
  && echo "alias magento='php bin/magento'" >> /home/${CUSTOM_USER_NAME}/.zshrc

RUN apk add zsh curl wget git
# Install ZSH & config
RUN ZSH="/home/${CUSTOM_USER_NAME}/.oh-my-zsh" sh -c "$(curl -fsSL https://raw.github.com/ohmyzsh/ohmyzsh/master/tools/install.sh)" chsh -s $(which zsh) ${CUSTOM_USER_NAME} \
  && mv /root/.zshrc /home/${CUSTOM_USER_NAME}/.zshrc
  # && chown -R ${CUSTOM_USER_NAME}:${CUSTOM_USER_NAME} /home/${CUSTOM_USER_NAME}/.oh-my-zsh \
RUN sed -i -e "s/bin\/ash/bin\/zsh/" /etc/passwd 
RUN sed -i 's^ZSH_THEME="robbyrussell"^ZSH_THEME="bira"^g' /home/${CUSTOM_USER_NAME}/.zshrc
RUN sed -i 's^# DISABLE_UPDATE_PROMPT="true"^DISABLE_UPDATE_PROMPT="true"^g' /home/${CUSTOM_USER_NAME}/.zshrc 
RUN sed -i 's^# DISABLE_AUTO_UPDATE="true"^DISABLE_AUTO_UPDATE="true"^g' /home/${CUSTOM_USER_NAME}/.zshrc 
RUN echo 'export EDITOR="nano"' >> /home/${CUSTOM_USER_NAME}/.zshrc 
RUN git config --global oh-my-zsh.hide-info 1

# Update permissions for supervisor and cron
RUN chown ${CUSTOM_USER_NAME}:${CUSTOM_USER_NAME} /etc/supervisor/supervisord.pid \
  && chmod gu+rw /var/run 

COPY ./conf/docker-entrypoint.d/*.sh /docker-entrypoint.d/
COPY ./conf/docker-entrypoint.sh /docker-entrypoint.sh
COPY ./conf/sudo-dev /etc/sudoers.d/"${CUSTOM_USER_NAME}"

RUN chmod +x /docker-entrypoint.sh

ENTRYPOINT ["/docker-entrypoint.sh"]