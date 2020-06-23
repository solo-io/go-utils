package internal

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/rotisserie/eris"
	"github.com/solo-io/go-utils/pkgmgmtutils/brew/formula_updater_types"
)

var ErrNoShaDataFound = eris.New("pkgmgmtutils: no data in SHA256 file")

func NewRemoteShaGetter() formula_updater_types.RemoteShaGetter {
	return &remoteShaGetter{}
}

type remoteShaGetter struct{}

func (*remoteShaGetter) GetShaFromUrl(url string) (sha string, err error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if !(len(b) > 0) {
		return "", ErrNoShaDataFound
	}

	s := strings.Fields(string(b))
	if len(s) != 2 {
		return "", fmt.Errorf("pkgmgmtutils: Sha256 file %s is not in expected format", url)
	}

	return s[0], nil
}
