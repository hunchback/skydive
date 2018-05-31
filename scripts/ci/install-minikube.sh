#!/bin/bash

DIR="$(dirname "$0")"

OS=linux
ARCH=amd64
TARGET_DIR=/usr/bin

MINIKUBE_VERSION="v0.25.2"
MINIKUBE_URL="https://github.com/kubernetes/minikube/releases/download/$MINIKUBE_VERSION/minikube-$OS-$ARCH"

KUBECTL_VERSION="v1.9.4"
KUBECTL_URL="https://storage.googleapis.com/kubernetes-release/release/$KUBECTL_VERSION/bin/$OS/$ARCH/kubectl"

export MINIKUBE_WANTUPDATENOTIFICATION=false
export MINIKUBE_WANTREPORTERRORPROMPT=false
export MINIKUBE_HOME=$HOME
export CHANGE_MINIKUBE_NONE_USER=true
export KUBECONFIG=$HOME/.kube/config

uninstall_binary() {
        local prog=$1
        sudo rm -f $TARGET_DIR/$prog
}

install_binary() {
        local prog=$1
        local url=$2

        wget --no-check-certificate -O $prog $url
        if [ $? != 0 ]; then
                echo "failed to download $url"
                exit 1
        fi

        chmod a+x $prog
        sudo mv $prog $TARGET_DIR/$prog
}

check_minikube() {
        which minikube 2>/dev/null
        if [ $? != 0 ]; then
                echo "minikube is not installed. Please run install-minikube.sh install"
                exit 1
        fi

        sudo systemctl start docker
}

patch_iptables() {
        # HACK HACK HACK: kube-proxy passes iptables-restore flag -w5 which is
	# ignored in v1.6.1 (FC27) but fails v1.6.2 (FC28). Issue can be found
	# https://github.com/NixOS/nixpkgs/issues/35544, it's patch has not yet
	# been released in default kubernetes server used by minikube.
        . /etc/os-release
        if [ "$ID" == "fedora" -a "$VERSION_ID" == "28" ]; then
		sudo rm -rf /usr/sbin/iptable-restore
		sudo cp -f -a -L $DIR/iptables-restore /usr/sbin/.
        fi
}

install() {
        patch_iptables
        install_binary minikube $MINIKUBE_URL
        install_binary kubectl $KUBECTL_URL
}

uninstall() {
        uninstall_binary minikube
        uninstall_binary kubectl
}

stop() {
        check_minikube

        sudo -E minikube delete
        sudo rm -rf /etc/kubernetes
        sudo rm -rf /var/lib/localkube
        sudo rm -rf $HOME/.minikube $HOME/.kube
        sudo rm -rf /root/.minikube /root/.kube

        sudo docker system prune -af
        for i in $(sudo docker ps -aq --filter name=k8s); do
                sudo docker stop $i
                sudo docker rm $i
        done

        sudo systemctl stop localkube
        sudo systemctl disable localkube
}

start() {
        check_minikube

        local args="--vm-driver=none"
        local driver=$(sudo docker info --format '{{print .CgroupDriver}}')
        if [ -n "$driver" ]; then
                args="$args --extra-config=kubelet.CgroupDriver=$driver"
        fi

        sudo -E minikube start $args
        minikube status
        kubectl config use-context minikube

        # enable accesss from root account
        for i in .kube .minikube; do
                sudo rm -rf /root/$i
                sudo cp -ar $HOME/$i /root/$i
        done

        # check kube-proxy
        curl -I -k https://10.96.0.1
        kubectl get services kubernetes

        # check kube-dns
        kubectl delete pod busybox --grace-period=0 --force
        kubectl run -i --rm busybox --image=busybox --restart=Never -- \
                nslookup kubernetes.default
        kubectl get pods -n kube-system
}

status() {
        kubectl version
        kubectl config get-contexts
        minikube status
}

case "$1" in
        install)
                install
                ;;
        uninstall)
                uninstall
                ;;
        start)
                start
                ;;
        stop)
                stop
                ;;
        status)
                status
                ;;
        *)
                echo "$0 [install|uninstall|start|stop|status]"
                exit 1
                ;;
esac
