/*
** Copyright [2013-2017] [Megam Systems]
**
** Licensed under the Apache License, Version 2.0 (the "License");
** you may not use this file except in compliance with the License.
** You may obtain a copy of the License at
**
** http://www.apache.org/licenses/LICENSE-2.0
**
** Unless required by applicable law or agreed to in writing, software
** distributed under the License is distributed on an "AS IS" BASIS,
** WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
** See the License for the specific language governing permissions and
** limitations under the License.
 */
package carton

import (
	"encoding/json"
	"reflect"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/virtengine/libgo/api"
	"github.com/virtengine/libgo/pairs"
	"gopkg.in/yaml.v2"
)

const (
	ASSEMBLIESBUCKET = "assemblies"
)

//bunch Assemblys
type Cartons []*Carton

type ApiAssemblies struct {
	JsonClaz string       `json:"json_claz"`
	Results  []Assemblies `json:"results"`
}

//The grand elephant for megam cloud platform.
type Assemblies struct {
	Id          string              `json:"id" cql:"id"`
	OrgId       string              `json:"org_id" cql:"org_id"`
	JsonClaz    string              `json:"json_claz" cql:"json_claz"`
	Name        string              `json:"name" cql:"name"`
	AssemblysId []string            `json:"assemblies" cql:"assemblies"`
	Inputs      pairs.JsonPairs     `json:"inputs" cql:"inputs"`
	CreatedAt   string              `json:"created_at" cql:"created_at"`
	Assemblys   map[string]Assembly `json:"-"`
}

func (a *Assemblies) String() string {
	if d, err := yaml.Marshal(a); err != nil {
		return err.Error()
	} else {
		return string(d)
	}
}

/** A public function which pulls the assemblies for deployment.
and any others we do. **/

func Get(id, email string) (*Assemblies, error) {
	a := new(Assemblies)
	a.Id = id
	return a.get(newArgs(email, ""))
}

// get all assemblies under a user
func Gets(email, org string) ([]Assemblies, error) {
	return new(Assemblies).gets(newArgs(email, org))
}

func (a *Assemblies) get(args api.ApiArgs) (*Assemblies, error) {
	cl := api.NewClient(args, "/assemblies/"+a.Id)
	response, err := cl.Get()
	if err != nil {
		return nil, err
	}

	ac := &ApiAssemblies{}
	err = json.Unmarshal(response, ac)
	if err != nil {
		return nil, err
	}
	a = &ac.Results[0]
	return a, nil
}

func (a *Assemblies) gets(args api.ApiArgs) ([]Assemblies, error) {
	cl := api.NewClient(args, "/assemblies")
	response, err := cl.Get()
	if err != nil {
		return nil, err
	}

	ac := &ApiAssemblies{}
	err = json.Unmarshal(response, ac)
	if err != nil {
		return nil, err
	}
	return ac.Results, nil
}

//make cartons from assemblies.
func (a *Assemblies) MkCartons(email string) (Cartons, error) {
	newCs := make(Cartons, 0, len(a.AssemblysId))
	for _, ay := range a.AssemblysId {
		if len(strings.TrimSpace(ay)) > 1 {
			if ca, err := mkCarton(a.Id, ay, email); err != nil {
				return nil, err
			} else {
				ca.toBox()                //on success, make a carton2box if BoxLevel is BoxZero
				newCs = append(newCs, ca) //on success append carton
			}
		}
	}
	log.Debugf("Cartons %v", newCs)
	return newCs, nil
}

func (a *Assemblies) Delete(asmid, email string, removedAssemblys []string) {
	existingAssemblys := make([]string, len(a.AssemblysId))
	for i := 0; i < len(a.AssemblysId); i++ {
		if len(strings.TrimSpace(a.AssemblysId[i])) > 1 {
			existingAssemblys[i] = a.AssemblysId[i]
		}
	}
	args := newArgs(email, a.OrgId)
	if reflect.DeepEqual(existingAssemblys, removedAssemblys) {
		cl := api.NewClient(args, "/assemblies/"+asmid)
		_, _ = cl.Delete()
	}
}
