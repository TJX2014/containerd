/*
Copyright 2017 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package server

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"k8s.io/kubernetes/pkg/kubelet/apis/cri/v1alpha1/runtime"

	sandboxstore "github.com/containerd/cri-containerd/pkg/store/sandbox"
)

func TestPodSandboxStatus(t *testing.T) {
	const (
		id = "test-id"
		ip = "10.10.10.10"
	)
	createdAt := time.Now()
	config := &runtime.PodSandboxConfig{
		Metadata: &runtime.PodSandboxMetadata{
			Name:      "test-name",
			Uid:       "test-uid",
			Namespace: "test-ns",
			Attempt:   1,
		},
		Linux: &runtime.LinuxPodSandboxConfig{
			SecurityContext: &runtime.LinuxSandboxSecurityContext{
				NamespaceOptions: &runtime.NamespaceOption{
					HostNetwork: true,
					HostPid:     false,
					HostIpc:     true,
				},
			},
		},
		Labels:      map[string]string{"a": "b"},
		Annotations: map[string]string{"c": "d"},
	}
	metadata := sandboxstore.Metadata{
		ID:     id,
		Name:   "test-name",
		Config: config,
	}

	expected := &runtime.PodSandboxStatus{
		Id:        id,
		Metadata:  config.GetMetadata(),
		CreatedAt: createdAt.UnixNano(),
		Network:   &runtime.PodSandboxNetworkStatus{Ip: ip},
		Linux: &runtime.LinuxPodSandboxStatus{
			Namespaces: &runtime.Namespace{
				Options: &runtime.NamespaceOption{
					HostNetwork: true,
					HostPid:     false,
					HostIpc:     true,
				},
			},
		},
		Labels:      config.GetLabels(),
		Annotations: config.GetAnnotations(),
	}
	for desc, test := range map[string]struct {
		state         sandboxstore.State
		expectedState runtime.PodSandboxState
	}{
		"sandbox state ready": {
			state:         sandboxstore.StateReady,
			expectedState: runtime.PodSandboxState_SANDBOX_READY,
		},
		"sandbox state not ready": {
			state:         sandboxstore.StateNotReady,
			expectedState: runtime.PodSandboxState_SANDBOX_NOTREADY,
		},
	} {
		t.Logf("TestCase: %s", desc)
		status := sandboxstore.Status{
			CreatedAt: createdAt,
			State:     test.state,
		}
		expected.State = test.expectedState
		got := toCRISandboxStatus(metadata, status, ip)
		assert.Equal(t, expected, got)
	}
}
