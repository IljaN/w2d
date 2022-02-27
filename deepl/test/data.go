package test

import (
	"encoding/json"
	"github.com/IljaN/w2d/deepl"
	"io/ioutil"
	"testing"
)

const (
	DataSetShort = "short"
	DataSetLong  = "long"
)

type DataMap map[string]DataSet

var DataSets = DataMap{
	DataSetShort: {"short_src.txt", "short_tgt.json"},
	DataSetLong:  {"long_src.txt", "long_tgt.json"},
}

func (ds *DataMap) Get(id string) DataSet {
	return DataSets[id]
}

type DataSet struct {
	inTxt, outJson string
}

func (ds *DataSet) GetInTxt(t *testing.T) string {
	return string(readOrFail(t, ds.inTxt))
}
func (ds *DataSet) GetOutJson(t *testing.T) string {
	return string(readOrFail(t, ds.outJson))
}
func (ds *DataSet) GetOutTxt(t *testing.T) string {
	r := deepl.Response{}
	if err := json.Unmarshal(readOrFail(t, ds.outJson), &r); err != nil {
		t.Fatalf("Failed to unmarshall json: %s", ds.outJson)
	}

	return r.Translations[0].Text
}

func readOrFail(t *testing.T, filename string) []byte {
	f, err := ioutil.ReadFile(filename)
	if err != nil {
		t.Fatalf("Reading dataset from file %s failed", filename)
	}

	return f
}
