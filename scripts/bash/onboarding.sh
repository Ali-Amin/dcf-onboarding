#!/bin/bash

ONBOARDING_URL=$2

if [[ -z $1 ]];then
	echo "Usage:"
	echo "./preflight.sh <COMMAND> <ONBOARDING_URL>"
	echo ""
	echo "COMMAND:                swtp:        Installs SWTPM and runs it in the background"
	echo ""
	echo "                        upload_keys: Uses SWTPM to generate an RSA key pair and uploads"
	echo "                                     it to the onboarding service at the ONBOARDING_URL"
	echo ""
	echo "ONBOARDING_URL:         Required in case of upload_keys only"\n
	echo "                        This is the URL of the onboarding service for which the"
	echo "                        public key will be uploaded to"
fi

if [[ "$1" == *"swtpm"* ]];then
	echo "Installing required packages..."
	sudo apt-get update;
	sudo apt -y install dpkg-dev debhelper libssl-dev libtool net-tools libfuse-dev libglib2.0-dev \
		libgmp-dev expect libtasn1-dev socat python3-twisted gnutls-dev gnutls-bin  \
		libjson-glib-dev gawk git python3-setuptools softhsm2 libseccomp-dev automake autoconf \
		libtool gcc build-essential libssl-dev dh-exec pkg-config dh-autoreconf libtool-bin \
		tpm2-tools libtss0 libtss2-dev dh-apparmor unzip > /dev/null;

	cd /usr/lib;
	wget https://github.com/stefanberger/libtpms/archive/refs/tags/v0.10.0.zip;
	unzip v0.10.0.zip;
	rm v0.10.0.zip
	cd libtpms-0.10.0;
	./autogen.sh --with-openssl;
	make dist;
	dpkg-buildpackage -us -uc -j4;

	libtool --finish /usr/lib/x86_64-linux-gnu;

	sudo apt install ../libtpms*.deb;
	wget https://github.com/stefanberger/swtpm/archive/refs/tags/v0.10.0.zip;
	unzip v0.10.0.zip;
	rm v0.10.0.zip
	cd swtpm-0.10.0;
	dpkg-buildpackage -us -uc -j4;

	libtool --finish /usr/lib/x86_64-linux-gnu/swtpm;

	sudo apt install ../swtpm*.deb;

	echo "Running swtpm on port 2321 as a background process"
	swtpm socket --tpmstate dir=/tmp/ --tpm2 --server type=tcp,port=2321 \
		--ctrl type=tcp,port=2322 --flags not-need-init,startup-clear 

elif [[ $1 == *"upload_key"* ]];then
	if [[ -z $ONBOARDING_URL  ]];then
		echo "Onboarding service URL is required"
		exit 1
	fi

	echo "Generating keys using tpm..."
	export TPM2TOOLS_TCTI="swtpm:port=2321"
	tpm2_flushcontext -t;
	tpm2_createprimary -C o -c primary.ctx -Q;

	tpm2_flushcontext -t;
	tpm2_create -G rsa2048:rsassa:null -g sha256 -u key.pub -r key.priv -C primary.ctx -Q;

	tpm2_flushcontext -t;
	tpm2_load -C primary.ctx -u key.pub -r key.priv -c key.ctx;

	tpm2_flushcontext -t;
	tpm2_readpublic -c key.ctx  -o key.pem -f PEM -Q;

	tpm2_flushcontext -t;

	PUB_KEY=$(base64 key.pem | tr -d " \t\n\r");
	echo "Authentication to onboarding service is required to upload device keys"
	read -p "Username: " username;

	read -sp "Password: " password;

	
	echo "Uploading public key.."
	DEVICE_ID=$(hostname);
	curl $ONBOARDING_URL/device/$DEVICE_ID/key -d "{\"publicKey\": \"$PUB_KEY\"}" --user $username:$password -v;


	echo "Creating DCF user for ansible playbooks"
	useradd dcf -p $(perl -e 'print crypt($ARGV[0], "password")' "password");
	usermod -G root dcf;
	usermod -G sudo dcf

	echo "Node ready for onboarding"
elif [[ $1 == *"join_cluster"*]];then
	CONTROL_PLANE=$3
	BOOTSTRAP_TOKEN=$4
	CA_CERT_HASH=$5
	time_initiated=$(date +"%Y-%M-%dT%T.%8NZ")
	kubeadm join $CONTROL_PLANE --token $BOOTSTRAP_TOKEN --discovery-token-ca-cert-hash $CA_CERT_HASH
	echo $time_initiated >> time_file
fi
