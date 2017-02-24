package marketplaces

import (
	"bytes"
	"encoding/json"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/megamsys/libgo/api"
	"github.com/megamsys/libgo/cmd"
	"github.com/megamsys/libgo/pairs"
	"github.com/megamsys/libgo/utils"
	"github.com/megamsys/vertice/marketplaces/provision"
	"gopkg.in/yaml.v2"
	"io"
	"time"
)

const (
	APIRAWIMAGES        = "/rawimages"
	APIRAWIMAGES_UPDATE = "/rawimages/update"
)

type apiRawImages struct {
	JsonClaz string      `json:"json_claz"`
	Results  []RawImages `json:"results"`
}

type RawImages struct {
	Id         string          `json:"id"`
	Name       string          `json:"name"`
	AccountId  string          `json:"account_id"`
	OrgId      string          `json:"org_id"`
	Inputs     pairs.JsonPairs `json:"inputs"`
	Outputs    pairs.JsonPairs `json:"outputs"`
	Repository string          `json:"repos"`
	Repos      Repos           `json:"-"`
	Status     string          `json:"status"`
}

type Repos struct {
	Source     string          `json:"source"`
	PublicUrl  string          `json:"public_url"`
	Properties pairs.JsonPairs `json:"properties"`
}

func (r Repos) toMap() map[string]string {
	m := r.Properties.ToMap()
	m["source"] = r.Source
	m["public_url"] = r.PublicUrl
	return m
}

// marketplaces json string
func (s *RawImages) String() string {
	if d, err := yaml.Marshal(s); err != nil {
		return err.Error()
	} else {
		return string(d)
	}
}

func (r *RawImages) Get() (*RawImages, error) {
	return r.get()
}

/** A public function which pulls the snapshot for disk save as image.
and any others we do. **/
func (r *RawImages) get() (*RawImages, error) {
	cl := api.NewClient(newArgs(r.AccountId, ""), APIRAWIMAGES+"/"+r.Id)
	response, err := cl.Get()
	if err != nil {
		return nil, err
	}

	res := &apiRawImages{}
	err = json.Unmarshal(response, res)
	if err != nil {
		return nil, err
	}
	a := &res.Results[0]
	log.Debugf("rawimages  %v", a)
	return a, nil
}

func (r *RawImages) Update() error {
	return r.update()
}

func (r *RawImages) update() error {
	cl := api.NewClient(newArgs(r.AccountId, ""), APIRAWIMAGES_UPDATE)
	_, err := cl.Post(r)
	if err != nil {
		return err
	}
	return nil
}

// Deploy runs a deployment of an application. It will first try to run an
// image based deploy, and then fallback to the Git based deployment.
func (r *RawImages) create() error {
	var outBuffer bytes.Buffer
	start := time.Now()
	logWriter := LogWriter{}
	// logWriter.Async()
	// defer logWriter.Close()
	writer := io.MultiWriter(&outBuffer, &logWriter)
	err := r.deployToProvisioner(writer)
	elapsed := time.Since(start)
	if err != nil {
		return err
	}
	log.Debugf("%s in (%s)\n%s", cmd.Colorfy(r.Name, "cyan", "", "bold"),
		cmd.Colorfy(elapsed.String(), "green", "", "bold"),
		cmd.Colorfy(outBuffer.String(), "yellow", "", ""))
	return nil
}

func (r *RawImages) deployToProvisioner(writer io.Writer) error {
	box := r.mkBox()
	if deployer, ok := ProvisionerMap[utils.PROVIDER_ONE].(provision.RawImageAccess); ok {
		return deployer.ISODeploy(box, writer)
	}
	return fmt.Errorf("cannot provision rawimages")
}

func (r *RawImages) mkBox() *provision.Box {
	box := &provision.Box{
		CartonId:  r.Id,
		AccountId: r.AccountId,
		Name:      r.Name,
		Region:    r.Region(),
		Provider:  r.provider(),
	}
	box.Repos = r.Repos.toMap()
	return box
}

func (s *RawImages) Region() string {
	return s.Inputs.Match(utils.REGION)
}

func (a *RawImages) provider() string {
	return a.Inputs.Match(utils.PROVIDER)
}

func (a *RawImages) ImageId() string {
	return a.Outputs.Match(utils.RAW_IMAGE_ID)
}
