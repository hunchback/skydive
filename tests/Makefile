GOFLAGS = -gcflags='-N -l' -race
SRCS := tests.go k8s_test.go
PROG := ${PWD}/a.out
ARGS := \
        "-standalone" \
        "-analyzer.flow.backend=elasticsearch" \
        "-analyzer.topology.backend=elasticsearch" \
        "-analyzer.topology.probes=k8s" \

all: $(PROG)
	sudo $$(which dlv) exec $(PROG) -- $(ARGS)

run: $(PROG)
	sudo $(PROG) $(ARGS)

strace: $(PROG)
	sudo strace -o /tmp/strace.log -s 10000 -f $(PROG) $(ARGS)

build: $(PROG)

$(PROG): $(SRCS)
	go test $(GOFLAGS) $(SRCS) -c -o $@

clean:
	rm -rf $(PROG)
