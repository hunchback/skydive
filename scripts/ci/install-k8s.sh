#!/bin/bash

set -v

OS=linux
ARCH=amd64
TARGET_DIR="/usr/bin"

MINIKUBE_VERSION="v0.24.1"
MINIKUBE_URL="https://github.com/kubernetes/minikube/releases/download/$MINIKUBE_VERSION/minikube-${OS}-${ARCH}"

KUBECTL_VERSION="v1.9.0"
KUBECTL_URL="https://storage.googleapis.com/kubernetes-release/release/$KUBECTL_VERSION/bin/$OS/$ARCH/kubectl"

HELM_VERSION="v2.8.0"
HELM_URL="https://storage.googleapis.com/kubernetes-helm/helm-${HELM_VERSION}-${OS}-${ARCH}.tar.gz"

wget_file() {
	local file=$1
	local url=$2
	wget --no-check-certificate -O $file $url
}

uninstall_binary() {
	local prog=$1
	sudo rm -f ${TARGET_DIR}/$prog
}

install_binary() {
	local prog=$1
	local url=$2
	wget_file $prog $url
	chmod a+x $prog
	sudo mv -f $prog ${TARGET_DIR}/.
}

install_tgz() {
	local prog=$1
	local url=$2
	local tgz=$prog.tar.gz
	local tmpfile=$(mktemp /tmp/skydive.XXXXXXXXXX)
	mkdir -p $tmpfile
	cd $tmpfile
	wget_file $tgz $url
	tar xvfz $tgz
	sudo mv -f */$prog ${TARGET_DIR}/.
	rm -rf $tmpfile
}

install() {
	install_binary minikube $MINIKUBE_URL
	install_binary kubectl $KUBECTL_URL
	install_tgz helm $HELM_URL
}

uninstall() {
	uninstall_binary minikube
	uninstall_binary kubectl
	uninstall_binary helm
}

stop() {
	sudo minikube delete || true
	sudo rm -rf ~/.minikube
}

start() {
	sudo CHANGE_MINIKUBE_NONE_USER=true minikube --vm-driver=none start
	sudo minikube status
	kubectl config use-context minikube
	helm init
	helm repo update
}

status() {
	kubectl version
	kubectl config get-contexts
	export HELM_HOME=$(kubectl get svc -n=kube-system tiller-deploy --output=jsonpath='{.spec.clusterIP}:{.spec.ports[0].port}')
	helm version
	helm ls
}

case "$1" in
	start)
		stop
		uninstall
		install
		start
		;;
	stop)
		stop
		uninstall
		;;
	status)
		status
		;;
	*)
		echo "$0 [start|stop|status]"
		exit 1
		;;
esac

exit 0
