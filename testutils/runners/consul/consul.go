package consul

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/onsi/ginkgo/v2"
	"github.com/solo-io/go-utils/log"
	"github.com/solo-io/go-utils/testutils/runners"
)

const defaultConsulDockerImage = "consul:1.5.3"

type ConsulFactory struct {
	consulpath string
	tmpdir     string
	Ports      ConsulPorts
}

func NewConsulFactory() (*ConsulFactory, error) {
	consulpath := os.Getenv("CONSUL_BINARY")

	if consulpath == "" {
		consulPath, err := exec.LookPath("consul")
		if err == nil {
			log.Printf("Using consul from PATH: %s", consulPath)
			consulpath = consulPath
		}
	}

	ports := NewRandomConsulPorts()

	if consulpath != "" {
		return &ConsulFactory{
			consulpath: consulpath,
			Ports:      ports,
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

# just print the image sha for reproducibility
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
		Ports:      ports,
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
	Ports      ConsulPorts
}

func (ef *ConsulFactory) NewConsulInstance() (*ConsulInstance, error) {
	// try to grab one form docker...
	tmpdir, err := ioutil.TempDir(os.Getenv("HELPER_TMP"), "consul")
	if err != nil {
		return nil, err
	}

	cmd := exec.Command(ef.consulpath, append([]string{"agent", "-dev", "--client=0.0.0.0"}, ef.Ports.Flags()...)...)
	cmd.Dir = ef.tmpdir
	cmd.Stdout = ginkgo.GinkgoWriter
	cmd.Stderr = ginkgo.GinkgoWriter
	return &ConsulInstance{
		consulpath: ef.consulpath,
		tmpdir:     tmpdir,
		cmd:        cmd,
		Ports:      ef.Ports,
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

type ConsulPorts struct {
	DnsPort, HttpPort, GrpcPort, ServerPort, SerfLanPort, SerfWanPort int
}

func NewRandomConsulPorts() ConsulPorts {
	return ConsulPorts{
		HttpPort:    runners.AllocateParallelPort(8500),
		GrpcPort:    runners.AllocateParallelPort(8501),
		DnsPort:     runners.AllocateParallelPort(8502),
		ServerPort:  runners.AllocateParallelPort(8503),
		SerfLanPort: runners.AllocateParallelPort(8504),
		SerfWanPort: runners.AllocateParallelPort(8505),
	}
}

// return flags to set each port type as a string
func (p ConsulPorts) Flags() []string {
	return []string{
		fmt.Sprintf("--dns-port=%v", p.DnsPort),
		fmt.Sprintf("--grpc-port=%v", p.GrpcPort),
		fmt.Sprintf("--http-port=%v", p.HttpPort),
		fmt.Sprintf("--server-port=%v", p.ServerPort),
		fmt.Sprintf("--serf-lan-port=%v", p.SerfLanPort),
		fmt.Sprintf("--serf-wan-port=%v", p.SerfWanPort),
	}
}
