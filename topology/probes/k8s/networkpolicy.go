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

	"github.com/skydive-project/skydive/filters"
	"github.com/skydive-project/skydive/logging"
	"github.com/skydive-project/skydive/probe"
	"github.com/skydive-project/skydive/topology/graph"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/api/extensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/tools/cache"
)

type networkPolicyProbe struct {
	defaultKubeCacheEventHandler
	graph.DefaultGraphListener
	*kubeCache
	graph          *graph.Graph
	podCache       *kubeCache
	namespaceCache *kubeCache
	objIndexer     *graph.MetadataIndexer
}

func newObjectIndexerByNetworkPolicy(g *graph.Graph) *graph.MetadataIndexer {
	ownedByFilter := filters.NewOrFilter(
		filters.NewTermStringFilter("Type", "namespace"),
		filters.NewTermStringFilter("Type", "pod"),
	)

	filter := filters.NewAndFilter(
		filters.NewTermStringFilter("Manager", managerValue),
		ownedByFilter,
	)
	m := graph.NewGraphElementFilter(filter)
	return graph.NewMetadataIndexer(g, m)
}

func (n *networkPolicyProbe) newMetadata(np *v1beta1.NetworkPolicy) graph.Metadata {
	return newMetadata("networkpolicy", np.Namespace, np.Name, np)
}

func networkPolicyUID(np *v1beta1.NetworkPolicy) graph.Identifier {
	return graph.Identifier(np.GetUID())
}

func dumpNetworkPolicy(np *v1beta1.NetworkPolicy) string {
	return fmt.Sprintf("networkPolicy{Namespace: %s, Name: %s}", np.Namespace, np.Name)
}

func (n *networkPolicyProbe) OnAdd(obj interface{}) {
	if np, ok := obj.(*v1beta1.NetworkPolicy); ok {
		logging.GetLogger().Debugf("Adding %s", dumpNetworkPolicy(np))
		n.graph.Lock()
		npNode := newNode(n.graph, networkPolicyUID(np), n.newMetadata(np))
		n.updateLinks(npNode, np, nil)
		n.graph.Unlock()
	}
}

func (n *networkPolicyProbe) OnUpdate(oldObj, newObj interface{}) {
	if np, ok := newObj.(*v1beta1.NetworkPolicy); ok {
		logging.GetLogger().Debugf("Updating %s", dumpNetworkPolicy(np))
		n.graph.Lock()
		if npNode := n.graph.GetNode(networkPolicyUID(np)); npNode != nil {
			addMetadata(n.graph, npNode, np)
			n.updateLinks(npNode, np, nil)
		}
		n.graph.Unlock()
	}
}

func (n *networkPolicyProbe) OnDelete(obj interface{}) {
	if np, ok := obj.(*v1beta1.NetworkPolicy); ok {
		logging.GetLogger().Debugf("Deleting %s", dumpNetworkPolicy(np))
		n.graph.Lock()
		if npNode := n.graph.GetNode(networkPolicyUID((np))); npNode != nil {
			n.graph.DelNode(npNode)
		}
		n.graph.Unlock()
	}
}

func (n *networkPolicyProbe) filterPodByLabels(in []interface{}, podSelector *metav1.LabelSelector, namespace string) (out []interface{}) {
	selector, _ := metav1.LabelSelectorAsSelector(podSelector)
	for _, pod := range in {
		pod := pod.(*corev1.Pod)
		if namespace == pod.Namespace && selector.Matches(labels.Set(pod.Labels)) {
			out = append(out, pod)
		}
	}
	return
}

func (n *networkPolicyProbe) filterNamespaceByLabels(in []interface{}, podSelector *metav1.LabelSelector, namespace string) (out []interface{}) {
	selector, _ := metav1.LabelSelectorAsSelector(podSelector)
	if !selector.Empty() {
		return
	}

	for _, obj := range in {
		ns := obj.(*corev1.Namespace)
		if namespace == ns.Name {
			out = append(out, ns)
		}
	}
	return
}

func (n *networkPolicyProbe) selectedPods(podSelector *metav1.LabelSelector, namespace string) (nodes []*graph.Node) {
	pods := n.podCache.list()
	pods = n.filterPodByLabels(pods, podSelector, namespace)
	podNodes := n.podsToNodes(pods)
	nodes = append(nodes, podNodes...)
	logging.GetLogger().Debugf("found %d pods", len(nodes))
	return
}

func (n *networkPolicyProbe) selectedNamespaces(podSelector *metav1.LabelSelector, namespace string) (nodes []*graph.Node) {
	nss := n.namespaceCache.list()
	nss = n.filterNamespaceByLabels(nss, podSelector, namespace)
	nsNodes := n.namespacesToNodes(nss)
	nodes = append(nodes, nsNodes...)
	logging.GetLogger().Debugf("found %d namespaces", len(nodes))
	return
}

func (n *networkPolicyProbe) listObjectsConnectedToBegin(np *v1beta1.NetworkPolicy) (nodes []*graph.Node) {
	pods := n.selectedPods(&np.Spec.PodSelector, np.Namespace)
	nss := n.selectedNamespaces(&np.Spec.PodSelector, np.Namespace)
	return append(pods, nss...)
}

func (n *networkPolicyProbe) isPodSelected(np *v1beta1.NetworkPolicy, pod *corev1.Pod) bool {
	return len(n.filterPodByLabels([]interface{}{pod}, &np.Spec.PodSelector, np.Namespace)) == 1
}

func (n *networkPolicyProbe) isNamespaceSelected(np *v1beta1.NetworkPolicy, ns *corev1.Namespace) bool {
	return len(n.filterNamespaceByLabels([]interface{}{ns}, &np.Spec.PodSelector, np.Namespace)) == 1
}

func (n *networkPolicyProbe) isSelected(np *v1beta1.NetworkPolicy, obj interface{}) bool {
	switch obj := obj.(type) {
	case *corev1.Pod:
		return n.isPodSelected(np, obj)
	case *corev1.Namespace:
		return n.isNamespaceSelected(np, obj)
	default:
		return false
	}
}

type PolicyType string

const (
	PolicyTypeIngress PolicyType = "ingress"
	PolicyTypeEgress  PolicyType = "egress"
)

func (val PolicyType) String() string {
	return string(val)
}

type PolicyTarget string

const (
	PolicyTargetDeny  PolicyTarget = "deny"
	PolicyTargetAllow PolicyTarget = "allow"
)

func (val PolicyTarget) String() string {
	return string(val)
}

type PolicyPoint string

const (
	PolicyPointBegin PolicyPoint = "begin"
	PolicyPointEnd   PolicyPoint = "end"
)

func (val PolicyPoint) String() string {
	return string(val)
}

func (n *networkPolicyProbe) isIngress(np *v1beta1.NetworkPolicy) bool {
	if len(np.Spec.Ingress) != 0 {
		return true
	}
	if len(np.Spec.PolicyTypes) == 0 {
		return true
	}
	for _, ty := range np.Spec.PolicyTypes {
		if ty == v1beta1.PolicyTypeIngress {
			return true
		}
	}
	return false
}

func (n *networkPolicyProbe) isEgress(np *v1beta1.NetworkPolicy) bool {
	if len(np.Spec.Egress) != 0 {
		return true
	}
	for _, ty := range np.Spec.PolicyTypes {
		if ty == v1beta1.PolicyTypeEgress {
			return true
		}
	}
	return false
}

func (n *networkPolicyProbe) isPolicyType(np *v1beta1.NetworkPolicy, ty PolicyType) bool {
	switch ty {
	case PolicyTypeIngress:
		return n.isIngress(np)
	case PolicyTypeEgress:
		return n.isEgress(np)
	}
	return false
}

func (n *networkPolicyProbe) isIngressDeny(np *v1beta1.NetworkPolicy) bool {
	return len(np.Spec.Ingress) == 0
}

func (n *networkPolicyProbe) isIngressAllow(np *v1beta1.NetworkPolicy) bool {
	return len(np.Spec.Ingress) != 0
}

func (n *networkPolicyProbe) isIngressTarget(np *v1beta1.NetworkPolicy, target PolicyTarget) bool {
	switch target {
	case PolicyTargetAllow:
		return n.isIngressAllow(np)
	case PolicyTargetDeny:
		return n.isIngressDeny(np)
	}
	return false
}

func (n *networkPolicyProbe) getIngressAllow(np *v1beta1.NetworkPolicy) []*graph.Node {
	out := []*graph.Node{}
	for _, rule := range np.Spec.Ingress {
		for _, from := range rule.From {
			pods := n.filterPodByLabels(n.podCache.list(), from.PodSelector, np.Namespace)
			podNodes := n.podsToNodes(pods)
			out = append(out, podNodes...)
		}
		// TODO: add handling of rule.Ports
	}
	// TODO: missing namespace object
	return out
}

func (n *networkPolicyProbe) isEgressDeny(np *v1beta1.NetworkPolicy) bool {
	return len(np.Spec.Egress) == 0
}

func (n *networkPolicyProbe) isEgressAllow(np *v1beta1.NetworkPolicy) bool {
	return len(np.Spec.Egress) != 0
}

func (n *networkPolicyProbe) isEgressTarget(np *v1beta1.NetworkPolicy, target PolicyTarget) bool {
	switch target {
	case PolicyTargetAllow:
		return n.isEgressAllow(np)
	case PolicyTargetDeny:
		return n.isEgressDeny(np)
	}
	return false
}

func (n *networkPolicyProbe) podsToNodes(pods []interface{}) []*graph.Node {
	nodes := []*graph.Node{}
	for _, pod := range pods {
		pod := pod.(*corev1.Pod)
		if podNode := n.graph.GetNode(podUID(pod)); podNode != nil {
			nodes = append(nodes, podNode)
		}
	}
	return nodes
}

func (n *networkPolicyProbe) namespacesToNodes(namespaces []interface{}) []*graph.Node {
	nodes := []*graph.Node{}
	for _, ns := range namespaces {
		ns := ns.(*corev1.Namespace)
		if nsNode := n.graph.GetNode(namespaceUID(ns)); nsNode != nil {
			nodes = append(nodes, nsNode)
		}
	}
	return nodes
}

func (n *networkPolicyProbe) getEgressAllow(np *v1beta1.NetworkPolicy) []*graph.Node {
	out := []*graph.Node{}
	for _, rule := range np.Spec.Egress {
		for _, to := range rule.To {
			pods := n.filterPodByLabels(n.podCache.list(), to.PodSelector, np.Namespace)
			podNodes := n.podsToNodes(pods)
			out = append(out, podNodes...)
		}
		// TODO: add handling of rule.Ports
	}
	// TODO: missing namespace object
	return out
}

func (n *networkPolicyProbe) isPolicyTarget(np *v1beta1.NetworkPolicy, ty PolicyType, target PolicyTarget) bool {
	switch ty {
	case PolicyTypeIngress:
		return n.isIngressTarget(np, target)
	case PolicyTypeEgress:
		return n.isEgressTarget(np, target)
	}
	return false
}

func (n *networkPolicyProbe) listObjectsConnectedToEnd(np *v1beta1.NetworkPolicy, ty PolicyType) []*graph.Node {
	switch ty {
	case PolicyTypeIngress:
		return n.getIngressAllow(np)
	case PolicyTypeEgress:
		return n.getEgressAllow(np)
	}
	return []*graph.Node{}
}

func (n *networkPolicyProbe) newEdgeMetadata(ty PolicyType, target PolicyTarget, point PolicyPoint) graph.Metadata {
	m := newEdgeMetadata()
	m.SetField("RelationType", "networkpolicy")
	m.SetFieldAndNormalize("PolicyType", ty)
	m.SetFieldAndNormalize("PolicyTarget", target)
	m.SetFieldAndNormalize("PolicyPoint", point)
	return m
}

func (n *networkPolicyProbe) lookupChildren(npNode *graph.Node, m graph.Metadata) []*graph.Node {
	childFilter := graph.Metadata{
		"Manager": managerValue,
	}
	return n.graph.LookupChildren(npNode, childFilter, m)
}

func (n *networkPolicyProbe) updateLinksForTypeTargetPoint(np *v1beta1.NetworkPolicy, npNode, filterNode *graph.Node, ty PolicyType, target PolicyTarget, point PolicyPoint, selected []*graph.Node) {

	if len(selected) == 0 {
		return
	}

	logging.GetLogger().Debugf("prcessing %d children", len(selected))

	m := n.newEdgeMetadata(ty, target, point)

	staleChilderen := make(map[graph.Identifier]*graph.Node)
	for _, child := range n.lookupChildren(npNode, m) {
		staleChilderen[child.ID] = child
	}

	for _, objNode := range selected {
		if _, found := staleChilderen[objNode.ID]; found {
			delete(staleChilderen, objNode.ID)
		}
	}

	for _, childNode := range staleChilderen {
		if filterNode == nil || filterNode.ID == childNode.ID {
			delLink(n.graph, npNode, childNode, m)
		}
	}

	for _, objNode := range selected {
		if filterNode == nil || filterNode.ID == objNode.ID {
			addLink(n.graph, npNode, objNode, m)
		}
	}
}

func (n *networkPolicyProbe) updateLinksForTypeTarget(np *v1beta1.NetworkPolicy, npNode, filterNode *graph.Node, ty PolicyType, target PolicyTarget) {
	if !n.isPolicyType(np, ty) {
		return
	}

	if !n.isPolicyTarget(np, ty, target) {
		return
	}

	logging.GetLogger().Debugf("Refreshing: %s --(%s,%s)--> <object>", dumpNetworkPolicy(np), ty, target)

	n.updateLinksForTypeTargetPoint(np, npNode, filterNode, ty, target, PolicyPointBegin, n.listObjectsConnectedToBegin(np))
	n.updateLinksForTypeTargetPoint(np, npNode, filterNode, ty, target, PolicyPointEnd, n.listObjectsConnectedToEnd(np, ty))
}

func (n *networkPolicyProbe) updateLinks(npNode *graph.Node, np *v1beta1.NetworkPolicy, filterNode *graph.Node) {
	n.updateLinksForTypeTarget(np, npNode, filterNode, PolicyTypeIngress, PolicyTargetDeny)
	n.updateLinksForTypeTarget(np, npNode, filterNode, PolicyTypeIngress, PolicyTargetAllow)
	n.updateLinksForTypeTarget(np, npNode, filterNode, PolicyTypeEgress, PolicyTargetDeny)
	n.updateLinksForTypeTarget(np, npNode, filterNode, PolicyTypeEgress, PolicyTargetAllow)
}

func (n *networkPolicyProbe) getObjByNode(node *graph.Node) interface{} {
	var cache *kubeCache

	ty, _ := node.GetFieldString("Type")
	switch ty {
	case "pod":
		cache = n.podCache
	case "namespace":
		cache = n.namespaceCache
	default:
		return nil
	}

	namespace, _ := node.GetFieldString("Namespace")
	name, _ := node.GetFieldString("Name")
	obj := cache.getByKey(namespace, name)
	return obj
}

func (n *networkPolicyProbe) onNodeUpdated(objNode *graph.Node) {
	logging.GetLogger().Debugf("refreshing: %s", dumpGraphNode(objNode))
	obj := n.getObjByNode(objNode)
	if obj == nil {
		logging.GetLogger().Debugf("can't find %s", dumpGraphNode(objNode))
		return
	}

	for _, np := range n.kubeCache.list() {
		np := np.(*v1beta1.NetworkPolicy)
		logging.GetLogger().Debugf("refreshing: %s --> %s", dumpNetworkPolicy(np), dumpGraphNode(objNode))
		npNode := n.graph.GetNode(networkPolicyUID(np))
		if npNode == nil {
			logging.GetLogger().Debugf("can't find %s", dumpNetworkPolicy(np))
			continue
		}
		n.updateLinks(npNode, np, objNode)
	}
}

func (n *networkPolicyProbe) OnNodeAdded(node *graph.Node) {
	n.onNodeUpdated(node)
}

func (n *networkPolicyProbe) OnNodeUpdated(node *graph.Node) {
	n.onNodeUpdated(node)
}

func (n *networkPolicyProbe) Start() {
	n.objIndexer.AddEventListener(n)
	n.objIndexer.Start()
	n.kubeCache.Start()
	n.podCache.Start()
	n.namespaceCache.Start()
}

func (n *networkPolicyProbe) Stop() {
	n.objIndexer.RemoveEventListener(n)
	n.objIndexer.Stop()
	n.kubeCache.Stop()
	n.podCache.Stop()
	n.namespaceCache.Stop()
}

func newNetworkPolicyKubeCache(handler cache.ResourceEventHandler) *kubeCache {
	return newKubeCache(getClientset().ExtensionsV1beta1().RESTClient(), &v1beta1.NetworkPolicy{}, "networkpolicies", handler)
}

func newNetworkPolicyProbe(g *graph.Graph) probe.Probe {
	n := &networkPolicyProbe{
		graph:          g,
		podCache:       newPodKubeCache(nil),
		namespaceCache: newNamespaceKubeCache(nil),
		objIndexer:     newObjectIndexerByNetworkPolicy(g),
	}
	n.kubeCache = newNetworkPolicyKubeCache(n)
	return n
}
