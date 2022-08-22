#!/usr/bin/env zsh

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


APP_DOCROOT=/var/www/html
CACHE_PREFIX=/var/cache
TEMP_PREFIX=/tmp

###
### Source libs
###
init="$(find "${CONFIG_DIR}" -name '*.sh' -type f | sort -u)"
for f in ${init}; do
	# shellcheck disable=SC1090
	. "${f}"
done

function php-conf() {

	# Determine the PHP-FPM runtime environment
	CPU=$(grep -c ^processor /proc/cpuinfo)
	echo "${CPU}"
	TOTALMEM=$(free -m | awk '/^Mem:/{print $2}')
	echo "${TOTALMEM}"

	if [[ "$CPU" -le "2" ]]; then TOTALCPU=2; else TOTALCPU="${CPU}"; fi

	PHP_START_SERVERS="$(env_get "PHP_START_SERVERS" $(($TOTALCPU / 2)))"
	PHP_MIN_SPARE_SERVERS="$(env_get "PHP_MIN_SPARE_SERVERS" $(($TOTALCPU / 2)))"
	PHP_MAX_SPARE_SERVERS="$(env_get "PHP_MAX_SPARE_SERVERS" "${TOTALCPU}")"
	PHP_MEMORY_LIMIT="$(env_get "PHP_MEMORY_LIMIT" $(($TOTALMEM / 2)))"
	PHP_MAX_CHILDREN="$(env_get "PHP_MAX_CHILDREN" $(($TOTALMEM * 2)))"
	PHP_POST_MAX_SIZE="$(env_get "PHP_POST_MAX_SIZE" "50")"
	PHP_UPLOAD_MAX_FILESIZE="$(env_get "PHP_UPLOAD_MAX_FILESIZE" "50")"
	PHP_MAX_INPUT_VARS="$(env_get "PHP_MAX_INPUT_VARS" "1000")"
	PHP_MAX_EXECUTION_TIME="$(env_get "PHP_MAX_EXECUTION_TIME" "300")"

	PHP_OPCACHE_ENABLE="$(env_get "PHP_OPCACHE_ENABLE" "1")"
	PHP_OPCACHE_MEMORY_CONSUMPTION="$(env_get "PHP_OPCACHE_MEMORY_CONSUMPTION" $(($TOTALMEM / 6)))"

	PHP_FPM_PORT="$(env_get "PHP_FPM_PORT" "9000")"

	SKYWALKING_ENABLE="$(env_get "SKYWALKING_ENABLE" "1")"
	SKYWALKING_GRPC="$(env_get "SKYWALKING_GRPC" "grpc")"

	{
		echo '[global]'
		echo 'daemonize = no'
		echo 'log_level = error'
		echo
		echo '[www]'
		echo "listen = ${PHP_FPM_PORT}"
		echo 'pm = dynamic'
		echo "pm.max_children = ${PHP_MAX_CHILDREN}"
		echo 'pm.max_requests = 1000'
		echo "pm.start_servers = ${PHP_START_SERVERS}"
		echo "pm.min_spare_servers = ${PHP_MIN_SPARE_SERVERS}"
		echo "pm.max_spare_servers = ${PHP_MAX_SPARE_SERVERS}"
	} | sudo tee /usr/local/etc/php-fpm.d/zz-docker.conf

	{
		echo "max_execution_time=${PHP_MAX_EXECUTION_TIME}"
		echo "memory_limit=${PHP_MEMORY_LIMIT}M"
		echo 'error_reporting=1'
		echo 'display_errors=0'
		echo 'log_errors=1'
		echo 'user_ini.filename=user.ini'
		echo 'realpath_cache_size=2M'
		echo 'cgi.check_shebang_line=0'
		echo 'date.timezone=UTC'
		echo 'short_open_tag=Off'
		echo 'session.auto_start=Off'
		echo "upload_max_filesize=${PHP_UPLOAD_MAX_FILESIZE}M"
		echo "post_max_size=${PHP_POST_MAX_SIZE}M"
		echo 'file_uploads=On'
		echo 'file_uploads=On'
		echo "max_input_vars=${PHP_MAX_INPUT_VARS}"
		echo "auto_prepend_file=/home/dev/global.php"

		echo
		echo "opcache.enable=${PHP_OPCACHE_ENABLE}"
		echo 'opcache.enable_cli=0'
		echo 'opcache.save_comments=1'
		echo 'opcache.interned_strings_buffer=8'
		echo 'opcache.fast_shutdown=1'
		echo 'opcache.validate_timestamps=2'
		echo 'opcache.revalidate_freq=0'
		echo 'opcache.use_cwd=1'
		echo 'opcache.max_accelerated_files=100000'
		echo 'opcache.max_wasted_percentage=5'
		echo "opcache.memory_consumption=${PHP_OPCACHE_MEMORY_CONSUMPTION}M"
		echo 'opcache.consistency_checks=0'
		echo 'opcache.huge_code_pages=1'
		
		echo
	} | sudo tee /usr/local/etc/php/conf.d/50-setting.ini

	sudo mkdir -p "${CACHE_PREFIX}"/fastcgi/
}

#---------------------------------------------------------------------
# configure monit
#---------------------------------------------------------------------

function monit() {

	{
		echo 'set daemon 10'
		echo '    with START DELAY 10'
		echo 'set pidfile /var/run/monit.pid'
		echo 'set statefile /var/run/monit.state'
		echo 'set httpd port 2849 and'
		echo '    use address 0.0.0.0'
		echo '    allow 0.0.0.0/0.0.0.0'
		echo '    allow localhost'
		echo '    allow admin:monit'
		echo 'set logfile syslog'
		echo 'set eventqueue'
		echo '    basedir /var/run'
		echo '    slots 100'
		echo 'include /etc/monit.d/*'
	} | sudo tee /etc/monitrc

	# Start monit
	sudo find "/etc/monit.d" -maxdepth 4 -type f -exec sed -i -e 's|{{APP_DOCROOT}}|'"${APP_DOCROOT}"'|g' {} \;
	sudo find "/etc/monit.d" -maxdepth 4 -type f -exec sed -i -e 's|{{CACHE_PREFIX}}|'"${CACHE_PREFIX}"'|g' {} \;
	sudo find "/etc/monit.d" -maxdepth 4 -type f -exec sed -i -e 's|{{PHP_FPM_PORT}}|'"${PHP_FPM_PORT}"'|g' {} \;

	sudo chmod 700 /etc/monitrc
	run="sudo monit -c /etc/monitrc" && sudo bash -c "${run}"

}

#############################################################
## Entry Point
#############################################################

###
### Set Debug level
###
DEBUG_LEVEL="$(env_get "DEBUG_ENTRYPOINT" "0")"
log "info" "Debug level: ${DEBUG_LEVEL}" "${DEBUG_LEVEL}"

###
### Install extra modules
###
EXT_MODULES="$(env_get "ENABLE_MODULES" "")"
if [ "${EXT_MODULES}" != "" ]; then
	log "info" "Installing extra PHP modules" "${DEBUG_LEVEL}"
	run "sudo /usr/local/bin/install-php-extensions ${EXT_MODULES}" "${DEBUG_LEVEL}"
fi

###
### Setup xhprof viewer
###
log "info" "Setting up xhprof" "${DEBUG_LEVEL}"
if [ -d "/home/dev/xhprof" ] 
then
	run "sudo rm -rf /var/www/html/xhprof && sudo cp -r /home/dev/xhprof /var/www/html/xhprof" "${DEBUG_LEVEL}"
fi

###
### Startup
###

log "info" "Performing PHP-FPM configuration changes" "${DEBUG_LEVEL}"
php-conf

# log "info" "Setting up monit" "${DEBUG_LEVEL}"
# monit

log "info" "Starting $(sudo php-fpm -v 2>&1 | head -1)" "${DEBUG_LEVEL}"
exec "${@}"
