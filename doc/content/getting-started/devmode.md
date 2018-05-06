---
date: 2016-11-28T15:21:44+01:00
title: Development
---

A Vagrant box with all the required dependencies to compile Skydive and run its
testsuite is [available](https://app.vagrantup.com/skydive/boxes/skydive-dev).

In devmode your host skydive source tree is synchronized with the guest skydive
source tree, if your host $GOPATH is not set - then this will default to the
same source tree from which you have invoked the devmode VM.

If you want to explicitly specify the host skydive source tree, please use the
following commands:

```console
export GOPATH=$HOME/go
echo "export GOPATH=$GOPATH" >> ~/.bashrc
export SKYDIVE=$GOPATH/src/github.com/skydive-project
echo "export SKYDIVE=$SKYDIVE" >> ~/.bashrc
mkdir -p $SKYDIVE
cd $SKYDIVE
```

In case you don't already have a skydive source tree then run:

```console
git clone https://github.com/skydive-project/skydive.git
```

So as to create the devmode VM run, resulting on creating a box on either
`VirtualBox` and `libvirt`.

```console
cd skydive/contrib/dev
vagrant up
vagrant ssh
```

Next run full build and test cycle on the VM:

```console
vagrant$ cd $SKYDIVE 
vagrant$ make
vagrant$ make fmt
vagrant$ make test
vagrant$ make test.functionals
```

