#!/bin/bash

BOOKINFO_URL=https://raw.githubusercontent.com/istio/istio/release-1.0/samples/bookinfo

if [ -z "$NAMESPACE" ]; then
	NAMESPACE=bookinfo
fi

if [ -z "$HAS_ISTIO" ]; then
	HAS_ISTIO=false
fi

apply() {
	kubectl apply -n $NAMESPACE -f $1
}

delete() {
	kubectl delete --grace-period=0 --force -n $NAMESPACE -f $1
}

stop() {
	$HAS_ISTIO && delete $BOOKINFO_URL/networking/destination-rule-all.yaml
	$HAS_ISTIO && delete $BOOKINFO_URL/networking/bookinfo-gateway.yaml
	delete $BOOKINFO_URL/platform/kube/bookinfo.yaml
	delete $BOOKINFO_URL/platform/kube/bookinfo-ingress.yaml
	kubectl delete --grace-period=0 --force $NAMESPACE
}

start() {
	kubectl create namespace $NAMESPACE
	$HAS_ISTIO && apply $BOOKINFO_URL/networking/destination-rule-all.yaml
	$HAS_ISTIO && apply $BOOKINFO_URL/networking/bookinfo-gateway.yaml
	apply $BOOKINFO_URL/platform/kube/bookinfo.yaml
	apply $BOOKINFO_URL/platform/kube/bookinfo-ingress.yaml
}

case "$1" in
	stop)
		stop
		;;
	start)
		start
		;;
	*)
		echo "$0 [stop|start|help]"
		exit 1
		;;
esac
exit 0
