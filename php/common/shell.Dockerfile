# Setup user (to stay in-sync with the host system's user)
ARG CUSTOM_USER_NAME=dev
ENV CUSTOM_USER_NAME=$CUSTOM_USER_NAME

ARG CUSTOM_UID=1000
ENV CUSTOM_UID=$CUSTOM_UID

ARG CUSTOM_GID=1000
ENV CUSTOM_GID=$CUSTOM_GID

ARG PHPV
ENV PHPV=$PHPV

RUN groupadd -g ${CUSTOM_GID} ${CUSTOM_USER_NAME} \
  && useradd -m -u ${CUSTOM_UID} -g ${CUSTOM_USER_NAME} -G sudo -s /usr/bin/zsh ${CUSTOM_USER_NAME} \
  && passwd -d ${CUSTOM_USER_NAME} \
  && echo "${CUSTOM_USER_NAME} ALL=(ALL) NOPASSWD:ALL" >> /etc/sudoers

# Aliases
RUN echo "alias artisan='php artisan'" >> /home/${CUSTOM_USER_NAME}/.zshrc \
  && echo "alias magento='php bin/magento'" >> /home/${CUSTOM_USER_NAME}/.zshrc

RUN apt update && apt -y install gpg \
  && printf "%s\n" "deb http://ftp.de.debian.org/debian buster-backports main" | \
  sudo tee /etc/apt/sources.list.d/buster-backports.list \
  && gpg --keyserver keyserver.ubuntu.com --recv-keys 648ACFD622F3D138 || true \
  && gpg -a --export 648ACFD622F3D138 | sudo apt-key add - || true

RUN apt update && apt install -y zsh git monit \
  && ZSH="/home/${CUSTOM_USER_NAME}/.oh-my-zsh" sh -c "$(curl -fsSL https://raw.github.com/ohmyzsh/ohmyzsh/master/tools/install.sh)" chsh -s $(which zsh) ${CUSTOM_USER_NAME} \
  && mv /root/.zshrc /home/${CUSTOM_USER_NAME}/.zshrc \
  && chown -R ${CUSTOM_USER_NAME}:${CUSTOM_USER_NAME} /home/${CUSTOM_USER_NAME}/.oh-my-zsh \
  && sed -i 's^ZSH_THEME="robbyrussell"^ZSH_THEME="bira"^g' /home/${CUSTOM_USER_NAME}/.zshrc \
  && sed -i 's^# DISABLE_UPDATE_PROMPT="true"^DISABLE_UPDATE_PROMPT="true"^g' /home/${CUSTOM_USER_NAME}/.zshrc \
  && sed -i 's^# DISABLE_AUTO_UPDATE="true"^DISABLE_AUTO_UPDATE="true"^g' /home/${CUSTOM_USER_NAME}/.zshrc \
  && echo 'export EDITOR="nano"' >> /home/${CUSTOM_USER_NAME}/.zshrc \
  && su - ${CUSTOM_USER_NAME} -c "git config --global oh-my-zsh.hide-info 1" \
  && rm -rf /var/lib/apt/lists/*

# Update permissions for supervisor and cron
RUN chown ${CUSTOM_USER_NAME}:${CUSTOM_USER_NAME} /etc/supervisor/supervisord.pid \
  && chmod gu+rw /var/run

COPY ./conf/docker-entrypoint.d/*.sh /docker-entrypoint.d/
COPY ./conf/docker-entrypoint.sh /docker-entrypoint.sh
COPY ./conf/sudo-dev /etc/sudoers.d/"${CUSTOM_USER_NAME}"
COPY ./conf/monit /etc/monit.d/monit

RUN chmod +x /docker-entrypoint.sh

ENTRYPOINT ["/docker-entrypoint.sh"]