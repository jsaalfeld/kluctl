package e2e

import (
	"fmt"
	"github.com/kluctl/kluctl/v2/e2e/test-utils"
	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
	"sync"
	"testing"

	"github.com/kluctl/kluctl/v2/pkg/utils/uo"
)

type hooksTestContext struct {
	t *testing.T
	k *test_utils.EnvTestCluster

	p *test_utils.TestProject

	m              sync.Mutex
	seenConfigMaps []string
	runCount       int

	whh *test_utils.CallbackHandlerEntry
}

func (s *hooksTestContext) setupWebhook() {
	s.whh = s.k.AddWebhookHandler(schema.GroupVersionResource{
		Version: "v1", Resource: "configmaps",
	}, s.handleConfigmap)
}

func (s *hooksTestContext) removeWebhook() {
	s.k.RemoveWebhookHandler(s.whh)
}

func (s *hooksTestContext) handleConfigmap(request admission.Request) {
	if s.p.TestSlug() != request.Namespace {
		return
	}
	s.m.Lock()
	defer s.m.Unlock()

	x, err := uo.FromString(string(request.Object.Raw))
	if err != nil {
		s.t.Fatal(err)
	}
	generation, _, err := x.GetNestedInt("metadata", "generation")
	if err != nil {
		s.t.Fatal(err)
	}
	uid, _, err := x.GetNestedString("metadata", "uid")
	if err != nil {
		s.t.Fatal(err)
	}
	runCount := x.GetK8sLabel("tests.kluctl.io/runCount")
	if runCount == nil || *runCount != fmt.Sprintf("%d", s.runCount) {
		return
	}

	s.t.Logf("handleConfigmap: op=%s, name=%s/%s, generation=%d, uid=%s", request.Operation, request.Namespace, request.Name, generation, uid)

	s.seenConfigMaps = append(s.seenConfigMaps, request.Name)
}

func (s *hooksTestContext) clearSeenConfigmaps() {
	s.m.Lock()
	defer s.m.Unlock()
	s.t.Logf("clearSeenConfigmaps: %v", s.seenConfigMaps)
	s.seenConfigMaps = nil
}

func (s *hooksTestContext) addHookConfigMap(dir string, opts resourceOpts, isHelm bool, hook string, hookDeletionPolicy string) {
	annotations := make(map[string]string)
	if isHelm {
		annotations["helm.sh/hook"] = hook
		if hookDeletionPolicy != "" {
			annotations["helm.sh/hook-deletion-policy"] = hookDeletionPolicy
		}
	} else {
		annotations["kluctl.io/hook"] = hook
	}
	if hookDeletionPolicy != "" {
		annotations["kluctl.io/hook-deletion-policy"] = hookDeletionPolicy
	}

	opts.annotations = uo.CopyMergeStrMap(opts.annotations, annotations)

	s.addConfigMap(dir, opts)
}

func (s *hooksTestContext) addConfigMap(dir string, opts resourceOpts) {
	o := uo.New()
	o.SetK8sGVKs("", "v1", "ConfigMap")
	mergeMetadata(o, opts)
	o.SetNestedField(map[string]interface{}{}, "data")
	s.p.AddKustomizeResources(dir, []test_utils.KustomizeResource{
		{fmt.Sprintf("%s.yml", opts.name), "", o},
	})
}

func prepareHookTestProject(t *testing.T, hook string, hookDeletionPolicy string) *hooksTestContext {
	s := &hooksTestContext{
		t: t,
		k: defaultCluster2, // use cluster2 as it has webhooks setup
		p: test_utils.NewTestProject(t),
	}
	s.setupWebhook()
	t.Cleanup(func() {
		s.removeWebhook()
	})

	createNamespace(s.t, s.k, s.p.TestSlug())

	s.p.UpdateTarget("test", func(target *uo.UnstructuredObject) {
		_ = target.SetNestedField(s.k.Context, "context")
	})

	s.p.AddKustomizeDeployment("hook", nil, nil)

	s.addConfigMap("hook", resourceOpts{name: "cm1", namespace: s.p.TestSlug()})
	s.addHookConfigMap("hook", resourceOpts{name: "hook1", namespace: s.p.TestSlug()}, false, hook, hookDeletionPolicy)

	return s
}

func (s *hooksTestContext) incRunCount() {
	s.runCount++
	s.t.Logf("incRunCount: %d", s.runCount)
	s.p.UpdateDeploymentYaml(".", func(o *uo.UnstructuredObject) error {
		_ = o.SetNestedField(fmt.Sprintf("%d", s.runCount), "commonLabels", "tests.kluctl.io/runCount")
		return nil
	})
}

func (s *hooksTestContext) ensureHookExecuted(expectedCms ...string) {
	s.clearSeenConfigmaps()
	s.incRunCount()
	s.p.KluctlMust("deploy", "--yes", "-t", "test")
	assert.Equal(s.t, expectedCms, s.seenConfigMaps)
}

func TestHooksPreDeployInitial(t *testing.T) {
	t.Parallel()
	s := prepareHookTestProject(t, "pre-deploy-initial", "")
	s.ensureHookExecuted("hook1", "cm1")
	s.ensureHookExecuted("cm1")
}

func TestHooksPostDeployInitial(t *testing.T) {
	t.Parallel()
	s := prepareHookTestProject(t, "post-deploy-initial", "")
	s.ensureHookExecuted("cm1", "hook1")
	s.ensureHookExecuted("cm1")
}

func TestHooksPreDeployUpgrade(t *testing.T) {
	t.Parallel()
	s := prepareHookTestProject(t, "pre-deploy-upgrade", "")
	s.ensureHookExecuted("cm1")
	s.ensureHookExecuted("hook1", "cm1")
	s.ensureHookExecuted("hook1", "cm1")
}

func TestHooksPostDeployUpgrade(t *testing.T) {
	t.Parallel()
	s := prepareHookTestProject(t, "post-deploy-upgrade", "")
	s.ensureHookExecuted("cm1")
	s.ensureHookExecuted("cm1", "hook1")
	s.ensureHookExecuted("cm1", "hook1")
}

func doTestHooksPreDeploy(t *testing.T, hooks string) {
	s := prepareHookTestProject(t, hooks, "")
	s.ensureHookExecuted("hook1", "cm1")
	s.ensureHookExecuted("hook1", "cm1")
}

func doTestHooksPostDeploy(t *testing.T, hooks string) {
	s := prepareHookTestProject(t, hooks, "")
	s.ensureHookExecuted("cm1", "hook1")
	s.ensureHookExecuted("cm1", "hook1")
}

func doTestHooksPrePostDeploy(t *testing.T, hooks string) {
	s := prepareHookTestProject(t, hooks, "")
	s.ensureHookExecuted("hook1", "cm1", "hook1")
	s.ensureHookExecuted("hook1", "cm1", "hook1")
}

func TestHooksPreDeploy(t *testing.T) {
	t.Parallel()
	doTestHooksPreDeploy(t, "pre-deploy")
}

func TestHooksPreDeploy2(t *testing.T) {
	t.Parallel()
	// same as pre-deploy
	doTestHooksPreDeploy(t, "pre-deploy-initial,pre-deploy-upgrade")
}

func TestHooksPostDeploy(t *testing.T) {
	t.Parallel()
	doTestHooksPostDeploy(t, "post-deploy")
}

func TestHooksPostDeploy2(t *testing.T) {
	t.Parallel()
	// same as post-deploy
	doTestHooksPostDeploy(t, "post-deploy-initial,post-deploy-upgrade")
}

func TestHooksPrePostDeploy(t *testing.T) {
	t.Parallel()
	doTestHooksPrePostDeploy(t, "pre-deploy,post-deploy")
}

func TestHooksPrePostDeploy2(t *testing.T) {
	t.Parallel()
	doTestHooksPrePostDeploy(t, "pre-deploy-initial,pre-deploy-upgrade,post-deploy-initial,post-deploy-upgrade")
}
