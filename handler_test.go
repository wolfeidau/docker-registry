package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"regexp"
	"testing"

	"github.com/Sirupsen/logrus"
	"github.com/remogatto/prettytest"
)

type testSuite struct {
	prettytest.Suite
}

func TestRunner(t *testing.T) {
	logger.Level = logrus.WarnLevel
	prettytest.RunWithFormatter(
		t,
		new(prettytest.TDDFormatter),
		new(testSuite),
	)
}

func (t *testSuite) TestRegexp() {
	re := regexp.MustCompile("^/v(\\d+)/ping")
	str := re.FindAllStringSubmatch("/v1/ping", -1)
	t.True(len(str) > 0)
}

func (t *testSuite) TestRepositories() {
	dir, _ := os.Getwd()
	root := dir + "/fixtures/index"
	repo := &Repository{root + "/repositories/dynport/redis/"}
	tags := repo.Tags()
	t.Equal("e0acc43660ac918e0cd7f21f1020ee3078fec7b2c14006603bbc21499799e7d5", tags["latest"])
}

func (t *testSuite) TestImage() {
	dir, _ := os.Getwd()
	root := dir + "/fixtures/index"
	image := &Image{root + "/images/e0acc43660ac918e0cd7f21f1020ee3078fec7b2c14006603bbc21499799e7d5"}
	atts, err := image.Attributes()
	if err != nil {
		t.Failed()
	}
	t.Equal("0e03f25112cd513ade7c194109217b9381835ac2298bd0ffb61d28fbe47081a8", atts.Parent)
	ancestry := image.Ancestry()
	t.Equal(3, len(ancestry))
	t.Equal("e0acc43660ac918e0cd7f21f1020ee3078fec7b2c14006603bbc21499799e7d5", ancestry[0])
	t.Equal("0e03f25112cd513ade7c194109217b9381835ac2298bd0ffb61d28fbe47081a8", ancestry[1])
	t.Equal("8dbd9e392a964056420e5d58ca5cc376ef18e2de93b5cc90e868a1bbc8318c1c", ancestry[2])
}

func resetTmpDataDir() string {
	dir, _ := os.Getwd()
	dataDir := dir + "/tmp/data"
	os.RemoveAll(dataDir)
	os.MkdirAll(dataDir, 0755)
	return dataDir
}

func (t *testSuite) TestWriteImageResource() {
	h := NewHandler(resetTmpDataDir(), nil)
	ser := httptest.NewServer(h)
	defer ser.Close()

	reader := bytes.NewReader([]byte("content"))
	req, _ := http.NewRequest("PUT", ser.URL+"/v1/images/1234/json", reader)
	client := http.Client{}
	rsp, _ := client.Do(req)

	t.Equal(rsp.StatusCode, 200)

	data, err := ioutil.ReadFile(h.DataDir + "/images/1234/json")
	if err != nil {
		logger.Error(err.Error())
	}
	t.Equal("content", string(data))
}

func (t *testSuite) TestPutRepositoryTag() {
	h := NewHandler(resetTmpDataDir(), nil)
	ser := httptest.NewServer(h)
	defer ser.Close()

	reader := bytes.NewReader([]byte("thetag"))
	req, _ := http.NewRequest("PUT", ser.URL+"/v1/repositories/dynport/test/tags/latest", reader)
	client := http.Client{}
	rsp, _ := client.Do(req)

	t.Equal(200, rsp.StatusCode)

	data, err := ioutil.ReadFile(h.DataDir + "/repositories/dynport/test/tags/latest")
	if err != nil {
		logger.Error(err.Error())
	}
	t.Equal("thetag", string(data))
}

func (t *testSuite) TestPutRepositoryImages() {
	h := NewHandler(resetTmpDataDir(), nil)
	ser := httptest.NewServer(h)
	defer ser.Close()

	reader := bytes.NewReader([]byte("imagesdata"))
	req, _ := http.NewRequest("PUT", ser.URL+"/v1/repositories/dynport/test/images", reader)
	client := http.Client{}
	rsp, _ := client.Do(req)

	t.Equal(204, rsp.StatusCode)

	data, err := ioutil.ReadFile(h.DataDir + "/repositories/dynport/test/images")
	if err != nil {
		logger.Error(err.Error())
	}
	t.Equal("imagesdata", string(data))
}

func (t *testSuite) TestGetImageJson() {
	h := NewHandler(resetTmpDataDir(), nil)
	ser := httptest.NewServer(h)
	defer ser.Close()

	reader := bytes.NewReader([]byte("just a test"))
	req, _ := http.NewRequest("GET", ser.URL+"/v1/images/123/json", reader)
	client := http.Client{}
	rsp, _ := client.Do(req)

	t.Equal(404, rsp.StatusCode)
}

func (t *testSuite) TestPutRepository() {
	h := NewHandler(resetTmpDataDir(), nil)
	ser := httptest.NewServer(h)
	defer ser.Close()

	reader := bytes.NewReader([]byte("just a test"))
	req, _ := http.NewRequest("PUT", ser.URL+"/v1/repositories/dynport/test/", reader)
	client := http.Client{}
	rsp, _ := client.Do(req)

	t.Equal(200, rsp.StatusCode)

	data, err := ioutil.ReadFile(h.DataDir + "/repositories/dynport/test/_index")
	if err != nil {
		logger.Error(err.Error())
	}

	t.Equal("just a test", string(data))
}

func (t *testSuite) TestReadFromServer() {
	dir, _ := os.Getwd()
	dataDir := dir + "/fixtures/index"
	ser := httptest.NewServer(NewHandler(dataDir, nil))
	defer ser.Close()

	r, _ := http.Get(ser.URL + "/v1/_ping")
	t.Equal(200, r.StatusCode)
	body, _ := ioutil.ReadAll(r.Body)
	r.Body.Close()
	t.Equal("pong", string(body))
	t.Equal(r.Header.Get("X-Docker-Registry-Version"), "0.6.0")

	r, _ = http.Get(ser.URL + "/v1/images/e0acc43660ac918e0cd7f21f1020ee3078fec7b2c14006603bbc21499799e7d5/json")
	t.Equal(200, r.StatusCode)
	t.Equal(r.Header.Get("X-Docker-Size"), "93")

	r, _ = http.Get(ser.URL + "/v1/images/e0acc43660ac918e0cd7f21f1020ee3078fec7b2c14006603bbc21499799e7d5/ancestry")
	t.Equal(200, r.StatusCode)
}

func (t *testSuite) TestBasicServer() {
	users := NewSingleUserStore("test1234asdfg")

	auth := NewBasicAuth(users, "testing")

	r := &http.Request{}

	r.Header = map[string][]string{
		"Authorization": []string{"Basic dGVzdHRlc3Q6dGVzdDEyMzRhc2RmZw=="},
	}

	if session, err := auth.CheckAuth(r); err != nil {
		t.Error(err)
	} else {
		t.Equal("testtest", session.Login)
		t.Equal(SessionNew, session.Status)
	}

}

func (t *testSuite) TestToken() {
	users := NewSingleUserStore("test1234asdfg")

	auth := NewBasicAuth(users, "testing")

	r := &http.Request{}

	r.Header = map[string][]string{
		"Authorization": []string{"Basic dGVzdHRlc3Q6dGVzdDEyMzRhc2RmZw=="},
	}

	session, err := auth.CheckAuth(r)

	if err != nil {
		t.Error(err)
	}

	t.Equal("testtest", session.Login)

	r.Header = map[string][]string{
		"Authorization": []string{fmt.Sprintf("Token %s", session.Token)},
	}

	if tsession, err := auth.CheckAuth(r); err != nil {
		t.Error(err)
	} else {
		t.Equal(session.Token, tsession.Token)
		t.Equal(SessionExisting, tsession.Status)
	}

}
