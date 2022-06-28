#!/usr/bin/env bash

set -e
set -u
set -o pipefail

log() {
	local type="${1}"     # ok, warn or err
	local message="${2}"  # msg to print
	local debug="${3}"    # 0: only warn and error, >0: ok and info

	local clr_ok="\033[0;32m"
	local clr_info="\033[0;34m"
	local clr_warn="\033[0;33m"
	local clr_err="\033[0;31m"
	local clr_rst="\033[0m"

	if [ "${type}" = "ok" ]; then
		if [ "${debug}" -gt "0" ]; then
			printf "${clr_ok}[OK]   %s${clr_rst}\n" "${message}"
		fi
	elif [ "${type}" = "info" ]; then
		if [ "${debug}" -gt "0" ]; then
			printf "${clr_info}[INFO] %s${clr_rst}\n" "${message}"
		fi
	elif [ "${type}" = "warn" ]; then
		printf "${clr_warn}[WARN] %s${clr_rst}\n" "${message}" 1>&2	# stdout -> stderr
	elif [ "${type}" = "err" ]; then
		printf "${clr_err}[ERR]  %s${clr_rst}\n" "${message}" 1>&2	# stdout -> stderr
	else
		printf "${clr_err}[???]  %s${clr_rst}\n" "${message}" 1>&2	# stdout -> stderr
	fi
}


###
### Wrapper for run_run command
###
run() {
	local cmd="${1}"      # command to execute
	local debug="${2}"    # show commands if debug level > 1

	local clr_red="\033[0;31m"
	local clr_green="\033[0;32m"
	local clr_reset="\033[0m"

	if [ "${debug}" -gt "1" ]; then
		printf "${clr_red}%s \$ ${clr_green}${cmd}${clr_reset}\n" "$( whoami )"
	fi
	
	/bin/sh -c "LANG=C LC_ALL=C ${cmd} &"
}


###
### Is argument a positive integer?
###
isint() {
	test -n "${1##*[!0-9]*}"
}


###
### Is env variable set?
###
env_set() {
	printenv "${1}" >/dev/null 2>&1
}


###
### Get env variable by name
###
env_get() {
	local env_name="${1}"

	# Did we have a default value specified?
	if [ "${#}" -gt "1" ]; then
		if ! env_set "${env_name}"; then
			echo "${2}"
			return 0
		fi
	fi
	# Just output the env value
	printenv "${1}"
}

start() {

	DEST_FILE="$(env_get "DEST" "docker.ovpn")"
	bindip=$(getent hosts "bind" | awk '{ print $1 }')


	if [ ! -f "/host/$DEST_FILE" ]; then
		log "info" "*** REGENERATING ALL CONFIGS ***" "${DEBUG_LEVEL}";
	
		rm -rf /etc/openvpn/*
		ovpn_genconfig -u tcp://localhost
		sed -i 's|^push|#push|' /etc/openvpn/openvpn.conf
		sed -i 's|^comp-lzo no|comp-lzo yes|' /etc/openvpn/openvpn.conf
		echo -e "compress lzo\npush \"compress lzo\"\ncomp-lzo\npush \"dhcp-option DOMAIN bind\"\npush \"dhcp-option DOMAIN loc\"\npush \"dhcp-option DNS ${bindip}\"" >> /etc/openvpn/openvpn.conf
		echo localhost | ovpn_initpki nopass
		easyrsa build-client-full host nopass
		ovpn_getclient host | sed '
			s|localhost 1194|localhost 13194|;
			s|redirect-gateway.*|route 192.168.0.0 255.252.0.0|;
		' > "/host/$DEST_FILE"
		echo -e "comp-lzo yes\ncomp-lzo\ncompress lzo" >> "/host/$DEST_FILE"
	fi

	# Workaround for https://github.com/wojas/docker-mac-network/issues/6
	/sbin/iptables -I FORWARD 1 -i tun+ -j ACCEPT

}

###
### Set Debug level
###
DEBUG_LEVEL="$(env_get "DEBUG_ENTRYPOINT" "0")"
log "info" "Debug level: ${DEBUG_LEVEL}" "${DEBUG_LEVEL}"

#############################################################
## Entry Point
#############################################################


###
### OVPN config
###
log "info" "Building ovpn config" "${DEBUG_LEVEL}"
start


###
### Start socat
###
log "info" "Starting socat" "${DEBUG_LEVEL}"
run "socat TCP-LISTEN:13194,fork,reuseaddr TCP-CONNECT:127.0.0.1:1194" "${DEBUG_LEVEL}"

###
### Start OVPN
###
log "info" "Starting ovpn..." "${DEBUG_LEVEL}"
exec "${@}"
