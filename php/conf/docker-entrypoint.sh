#!/usr/bin/env bash

set -e
set -u
set -o pipefail


###
### Globals
###

# The following global variables are available by our Dockerfile itself:
#   CUSTOM_USER_NAME
#   CUSTOM_GROUP
#   MY_UID
#   MY_GID

# Path to scripts to source
CONFIG_DIR="/docker-entrypoint.d"
CUSTOM_USER_NAME="dev"
CUSTOM_GROUP="dev"


###
### Source libs
###
init="$( find "${CONFIG_DIR}" -name '*.sh' -type f | sort -u )"
for f in ${init}; do
	# shellcheck disable=SC1090
	. "${f}"
done



#############################################################
## Entry Point
#############################################################

###
### Set Debug level
###
DEBUG_LEVEL="$( env_get "DEBUG_ENTRYPOINT" "0" )"
log "info" "Debug level: ${DEBUG_LEVEL}" "${DEBUG_LEVEL}"

# run "sudo echo 'Defaults env_keep += \"PHP_INI_DIR\"' >> /etc/sudoers.d/env_keep" "${DEBUG_LEVEL}"

###
### Install extra modules
###
EXT_MODULES="$( env_get "ENABLE_MODULES" "" )"
log "info" "Installing extra PHP modules" "${DEBUG_LEVEL}"
run "sudo /usr/local/bin/install-php-extensions ${EXT_MODULES}" "${DEBUG_LEVEL}"

###
### Startup
###
log "info" "Starting $( sudo php-fpm -v 2>&1 | head -1 )" "${DEBUG_LEVEL}"
exec "${@}"
