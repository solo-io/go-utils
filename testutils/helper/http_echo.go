package helper

const (
	defaultHttpEchoImage = "kennship/http-echo:latest"
	HttpEchoName         = "http-echo"
	HttpEchoPort         = 3000

)


func NewEchoHttp(namespace string) (*TestContainer, error) {
	return newTestContainer(namespace, defaultHttpEchoImage, HttpEchoName, HttpEchoPort)
}