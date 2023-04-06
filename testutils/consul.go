package testutils

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"io/ioutil"

	"time"

	"github.com/onsi/ginkgo/v2"
)

const defaultConsulDockerImage = "consul@sha256:6ffe55dcc1000126a6e874b298fe1f1b87f556fb344781af60681932e408ec6a"

type ConsulFactory struct {
	consulpath string
	tmpdir     string
}

func NewConsulFactory() (*ConsulFactory, error) {
	consulpath := os.Getenv("CONSUL_BINARY")

	if consulpath != "" {
		return &ConsulFactory{
			consulpath: consulpath,
		}, nil
	}

	// try to grab one form docker...
	tmpdir, err := ioutil.TempDir(os.Getenv("HELPER_TMP"), "consul")
	if err != nil {
		return nil, err
	}

	bash := fmt.Sprintf(`
set -ex
CID=$(docker run -d  %s /bin/sh -c exit)

# just print the image sha for repoducibility
echo "Using Consul Image:"
docker inspect %s -f "{{.RepoDigests}}"

docker cp $CID:/bin/consul .
docker rm -f $CID
    `, defaultConsulDockerImage, defaultConsulDockerImage)
	scriptfile := filepath.Join(tmpdir, "getconsul.sh")

	ioutil.WriteFile(scriptfile, []byte(bash), 0755)

	cmd := exec.Command("bash", scriptfile)
	cmd.Dir = tmpdir
	cmd.Stdout = ginkgo.GinkgoWriter
	cmd.Stderr = ginkgo.GinkgoWriter
	if err := cmd.Run(); err != nil {
		return nil, err
	}

	return &ConsulFactory{
		consulpath: filepath.Join(tmpdir, "consul"),
		tmpdir:     tmpdir,
	}, nil
}

func (ef *ConsulFactory) Clean() error {
	if ef == nil {
		return nil
	}
	if ef.tmpdir != "" {
		os.RemoveAll(ef.tmpdir)

	}
	return nil
}

type ConsulInstance struct {
	consulpath string
	tmpdir     string
	cmd        *exec.Cmd
}

func (ef *ConsulFactory) NewConsulInstance() (*ConsulInstance, error) {
	// try to grab one form docker...
	tmpdir, err := ioutil.TempDir(os.Getenv("HELPER_TMP"), "consul")
	if err != nil {
		return nil, err
	}

	cmd := exec.Command(ef.consulpath, "agent", "-dev", "--client=0.0.0.0")
	cmd.Dir = ef.tmpdir
	cmd.Stdout = ginkgo.GinkgoWriter
	cmd.Stderr = ginkgo.GinkgoWriter
	return &ConsulInstance{
		consulpath: ef.consulpath,
		tmpdir:     tmpdir,
		cmd:        cmd,
	}, nil

}

func (i *ConsulInstance) Silence() {
	i.cmd.Stdout = nil
	i.cmd.Stderr = nil
}

func (i *ConsulInstance) Run() error {
	return i.RunWithPort()
}

func (i *ConsulInstance) RunWithPort() error {
	err := i.cmd.Start()
	if err != nil {
		return err
	}
	time.Sleep(time.Millisecond * 1500)
	return nil
}

func (i *ConsulInstance) Binary() string {
	return i.consulpath
}

func (i *ConsulInstance) Clean() error {
	if i.cmd != nil {
		i.cmd.Process.Kill()
		i.cmd.Wait()
	}
	if i.tmpdir != "" {
		os.RemoveAll(i.tmpdir)
	}
	return nil
}
