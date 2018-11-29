/*

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

package composition

import (
	"log"
	"path/filepath"
	"testing"
	"time"

	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	context "github.com/ibm/cloud-operators/pkg/context"
	resv1 "github.com/ibm/cloud-operators/pkg/types/apis/resource/v1"

	"github.com/ibm/openwhisk-operator/pkg/apis"
	ow "github.com/ibm/openwhisk-operator/pkg/controller/common"
	owtest "github.com/ibm/openwhisk-operator/test"
)

var (
	c         client.Client
	cfg       *rest.Config
	namespace string
	scontext  context.Context
	cclient   *ow.CompositionClient
	t         *envtest.Environment
	stop      chan struct{}
)

func TestComposition(t *testing.T) {
	RegisterFailHandler(Fail)
	SetDefaultEventuallyPollingInterval(1 * time.Second)
	SetDefaultEventuallyTimeout(30 * time.Second)

	RunSpecs(t, "Composition Suite")
}

var _ = BeforeSuite(func() {
	logf.SetLogger(logf.ZapLoggerTo(GinkgoWriter, true))

	// Start kube apiserver
	t = &envtest.Environment{
		CRDDirectoryPaths:        []string{filepath.Join("..", "..", "..", "config", "crds")},
		ControlPlaneStartTimeout: 2 * time.Minute,
	}
	apis.AddToScheme(scheme.Scheme)

	var err error
	if cfg, err = t.Start(); err != nil {
		log.Fatal(err)
	}

	// Setup the Manager and Controller.  Wrap the Controller Reconcile function so it writes each request to a
	// channel when it is finished.
	mgr, err := manager.New(cfg, manager.Options{})
	Expect(err).NotTo(HaveOccurred())
	c = mgr.GetClient()

	recFn := newReconciler(mgr)
	Expect(add(mgr, recFn)).NotTo(HaveOccurred())
	stop = owtest.StartTestManager(mgr)

	// Initialize objects
	namespace = owtest.SetupKubeOrDie(cfg, "openwhisk-composition-")
	scontext = context.New(c, reconcile.Request{NamespacedName: types.NamespacedName{Name: "", Namespace: namespace}})

	clientset := owtest.GetClientsetOrDie(cfg)
	owtest.ConfigureOwprops("seed-defaults-owprops", clientset.CoreV1().Secrets(namespace))

	cclient, err = ow.NewCompositionClient(scontext, nil)
	Expect(err).NotTo(HaveOccurred())

})

var _ = AfterSuite(func() {
	close(stop)
	t.Stop()
})

var _ = Describe("composition", func() {

	DescribeTable("should be ready",
		func(specfile string) {
			composition := owtest.LoadComposition("testdata/" + specfile)
			obj := owtest.PostInNs(scontext, &composition, true, 0)
			Eventually(owtest.GetState(scontext, obj)).Should(Equal(resv1.ResourceStateOnline))

			params := make(map[string]string)
			params["msg"] = "Hello"

			Expect(owtest.CompositionInvocation(cclient, composition.Name, params)).Should(WithTransform(owtest.Result, HaveKeyWithValue("msg", "Hello")))

			err := scontext.Client().Delete(scontext, obj)
			Expect(err).NotTo(HaveOccurred())
		},
		Entry("with inline composition", "owc-inline.yaml"),
		Entry("with external composition", "owc-external.yaml"),
	)
})
