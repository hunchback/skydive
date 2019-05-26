// +build sa

/*
 * Copyright (C) 2019 IBM, Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy ofthe License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specificlanguage governing permissions and
 * limitations under the License.
 *
 */

package tests

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	"github.com/spf13/viper"

	"github.com/skydive-project/skydive/config"
	"github.com/skydive-project/skydive/contrib/objectstore/subscriber"
	g "github.com/skydive-project/skydive/gremlin"
)

const securityAdvisorConfigTemplate = `---
objectstore:
  endpoint: http://127.0.0.1:9000
  region: local
  bucket: bucket
  access_key: user
  secret_key: password
`

func setupSecurityAdvisor(c *TestContext) error {
	cfg := viper.New()

	analyzers := config.GetStringSlice("analyzers")
	if len(analyzers) == 0 {
		return errors.New("Missing analyzers config")
	}

	f, err := ioutil.TempFile("", "skydive_agent")
	if err != nil {
		return fmt.Errorf("failed to create configuration file: %s", err)
	}
	if _, err = f.Write([]byte(fmt.Sprintf(securityAdvisorConfigTemplate, analyzers[0]))); err != nil {
		return fmt.Errorf("failed to write configuration file: %s", err)
	}
	if err = f.Close(); err != nil {
		return fmt.Errorf("failed to close configuration file: %s", err)
	}

	configFile, err := os.Open(f.Name())
	if err != nil {
		return fmt.Errorf("failed to open configuration file: %s", err)
	}
	cfg.SetConfigType("yaml")
	if err := cfg.MergeConfig(configFile); err != nil {
		return fmt.Errorf("failed to update configuration: %s", err)
	}

	s, err := subscriber.NewSubscriberFromConfig(cfg)
	if err != nil {
		return err
	}

	s.Start()
	c.data["subscriber"] = s
	return err
}

func tearDownsecurityAdvisor(c *TestContext) error {
	rawS, ok := c.data["subscriber"]
	if ok {
		s := rawS.(*subscriber.Subscriber)
		s.Stop()

		storage := s.GetStorage()
		objectKeys, err := storage.ListObjects()
		if err != nil {
			return err
		}

		for _, objectKey := range objectKeys {
			if err = storage.DeleteObject(objectKey); err != nil {
				return err
			}
		}
	}
	return nil
}

const (
	container = "sa"
	// ifname value is the 1st interface created with template cni<index>
	ifname = "cni0"
	host   = "myhost"
	// networkName using the template of 0_0_<host>_0
	networkName = "0_0_myhost_0"
	// ip address must be within the default CNI network's pool (default 10.88.0.0/16)
	ip = "10.88.64.128"
)

func checkSecurityAdvisor(c *CheckContext) error {
	storage := c.data["subscriber"].(*subscriber.Subscriber).GetStorage()

	objectKeys, err := storage.ListObjects()
	if err != nil {
		return fmt.Errorf("Failed to list objects: %s", err)
	}

	flows := make([]*subscriber.SecurityAdvisorFlow, 0)
	for _, objectKey := range objectKeys {
		var objectFlows []*subscriber.SecurityAdvisorFlow
		if err := storage.ReadObjectFlows(objectKey, &objectFlows); err != nil {
			return fmt.Errorf("Failed to read object flows: %s", err)
		}

		flows = append(flows, objectFlows...)
	}

	found := false
	for _, fl := range flows {
		if fl.Network != nil && fl.Network.B == ip {
			if fl.NodeType != "bridge" {
				return fmt.Errorf("Expected 'bridge' NodeType, but got: %s", fl.NodeType)
			}

			if fl.Network.BName != networkName {
				return fmt.Errorf("Expected '"+networkName+"' B_Name, but got: %s", fl.Network.BName)
			}

			found = true
		}
	}

	if !found {
		return errors.New("No flows found with destination " + ip)
	}

	return nil
}

func TestSecurityAdvisor(t *testing.T) {
	test := &Test{
		setupFunction: setupSecurityAdvisor,

		preCleanup: true,

		setupCmds: []Cmd{
			{"podman run -d --network=host --name=" + container + " --hostname=" + host + " --ip=" + ip + " nginx", false},
		},

		injections: []TestInjection{{
			from:  g.G.V().Has("Name", ifname),
			toIP:  ip,
			count: 1,
		}},

		tearDownCmds: []Cmd{
			{"podman kill " + container, false},
			{"podman rm -f " + container, false},
		},

		captures: []TestCapture{
			{gremlin: g.G.V().Has("Name", ifname)},
		},

		mode: OneShot,

		checks: []CheckFunction{checkSecurityAdvisor},

		tearDownFunction: tearDownsecurityAdvisor,
	}

	RunTest(t, test)
}
