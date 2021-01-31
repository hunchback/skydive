.PHONY: install uinstall status logs
install uninstall status logs:
	./skydive.sh $@

.PHONY: broker east west
broker east west:
	kubectl --context k3d-$@ port-forward service/skydive-analyzer 8082:8082
