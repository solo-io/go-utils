package fileutils

import (
	"io/ioutil"

	"github.com/ghodss/yaml"
	"github.com/gogo/protobuf/proto"
	"github.com/rotisserie/eris"
	"github.com/solo-io/go-utils/protoutils"
)

func WriteToFile(filename string, pb proto.Message) error {
	jsn, err := protoutils.MarshalBytes(pb)
	if err != nil {
		return err
	}
	data, err := yaml.JSONToYAML(jsn)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(filename, data, 0644)
}

func ReadFileInto(filename string, v proto.Message) error {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return eris.Errorf("error reading file: %v", err)
	}
	jsn, err := yaml.YAMLToJSON(data)
	if err != nil {
		return err
	}
	return protoutils.UnmarshalBytes(jsn, v)
}
