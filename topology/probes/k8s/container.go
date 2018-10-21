/*
 * Copyright (C) 2017 Red Hat, Inc.
 *
 * Licensed to the Apache Software Foundation (ASF) under one
 * or more contributor license agreements.  See the NOTICE file
 * distributed with this work for additional information
 * regarding copyright ownership.  The ASF licenses this file
 * to you under the Apache License, Version 2.0 (the
 * "License"); you may not use this file except in compliance
 * with the License.  You may obtain a copy of the License at
 *
 *  http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an
 * "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 * KIND, either express or implied.  See the License for the
 * specific language governing permissions and limitations
 * under the License.
 *
 */

package k8s

import (
	"fmt"

	"github.com/skydive-project/skydive/topology/graph"

	"k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
)

type containerHandler struct {
	DefaultResourceHandler
	graph.DefaultGraphListener
	graph *graph.Graph
	cache *ResourceCache
}

// OnAdd is called when a new Kubernetes resource has been created
func (c *containerCache) OnAdd(obj interface{}) {
	c.graph.Lock()
	defer c.graph.Unlock()

	id, metadata := c.handler.Map(obj)
	node := c.graph.NewNode(id, metadata, "")
	c.NotifyEvent(graph.NodeAdded, node)
	logging.GetLogger().Debugf("Added %s", c.handler.Dump(obj))
}

// OnUpdate is called when a Kubernetes resource has been updated
func (c *containerCache) OnUpdate(oldObj, newObj interface{}) {
	c.graph.Lock()
	defer c.graph.Unlock()

	id, metadata := c.handler.Map(newObj)
	if node := c.graph.GetNode(id); node != nil {
		c.graph.SetMetadata(node, metadata)
		c.NotifyEvent(graph.NodeUpdated, node)
		logging.GetLogger().Debugf("Updated %s", c.handler.Dump(newObj))
	}
}

// OnDelete is called when a Kubernetes resource has been deleted
func (c *containerCache) OnDelete(obj interface{}) {
	c.graph.Lock()
	defer c.graph.Unlock()

	id, _ := c.handler.Map(obj)
	if node := c.graph.GetNode(id); node != nil {
		c.graph.DelNode(node)
		c.NotifyEvent(graph.NodeDeleted, node)
		logging.GetLogger().Debugf("Deleted %s", c.handler.Dump(obj))
	}
}

func newContainerProbe(clientset *kubernetes.Clientset, g *graph.Graph) Subprobe {
	return NewResourceCache(clientset.CoreV1().RESTClient(), &v1.Pod{}, "pods", g, &containerHandler{graph: g})
}
