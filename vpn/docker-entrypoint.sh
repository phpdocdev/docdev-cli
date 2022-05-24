#!/usr/bin/env bash

set -e
set -u
set -o pipefail

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

#############################################################
## Entry Point
#############################################################

###
### Start socat
###
run "socat TCP-LISTEN:13194,fork,reuseaddr TCP-CONNECT:127.0.0.1:1194" "1"

###
### Config and start openvpn
###
exec /run.sh
