package importcmd_test

import (
	"path/filepath"
	"testing"

	importcmd "github.com/jenkins-x/jx-extsecret/pkg/cmd/import"
	"github.com/jenkins-x/jx-extsecret/pkg/cmdrunner/fakerunner"
	"github.com/jenkins-x/jx-extsecret/pkg/extsecrets"
	"github.com/jenkins-x/jx-extsecret/pkg/extsecrets/testsecrets"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/runtime"
	dynfake "k8s.io/client-go/dynamic/fake"
	"k8s.io/client-go/kubernetes/fake"
)

func TestImport(t *testing.T) {
	var err error
	_, o := importcmd.NewCmdImport()
	scheme := runtime.NewScheme()

	ns := "jx"
	dynObjects := testsecrets.LoadExtSecretFiles(t, ns, "knative-docker-user-pass.yaml", "lighthouse-oauth-token.yaml")

	fakeDynClient := dynfake.NewSimpleDynamicClient(scheme, dynObjects...)
	o.SecretClient, err = extsecrets.NewClient(fakeDynClient)
	o.KubeClient = fake.NewSimpleClientset(testsecrets.AddVaultSecrets()...)

	require.NoError(t, err, "failed to create fake extsecrets Client")

	runner := &fakerunner.FakeRunner{}
	o.CommandRunner = runner.Run

	o.File = filepath.Join("test_data", "import-file.yaml")

	err = o.Run()
	require.NoError(t, err, "failed to run import")

	runner.ExpectResults(t,
		fakerunner.FakeResult{
			CLI: "vault version",
		},
		fakerunner.FakeResult{
			CLI: "vault kv list secret",
		},
		fakerunner.FakeResult{
			CLI: "vault kv put secret/pipelineUser email=jenkins-x@googlegroups.com token=dummyPipelineUser username=jenkins-x-labs-bot",
		},
	)
}
