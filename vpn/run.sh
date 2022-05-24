#!/bin/sh

dest=${dest:-docker.ovpn}

if [ ! -f "/host/$dest" ]; then
    echo "*** REGENERATING ALL CONFIGS ***"
    set -ex
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
    ' > "/host/$dest"
    echo -e "comp-lzo yes\ncomp-lzo\ncompress lzo" >> "/host/$dest"
fi

# Workaround for https://github.com/wojas/docker-mac-network/issues/6
/sbin/iptables -I FORWARD 1 -i tun+ -j ACCEPT

exec ovpn_run
