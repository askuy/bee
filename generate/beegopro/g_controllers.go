// Copyright 2013 bee authors
//
// Licensed under the Apache License, Version 2.0 (the "License"): you may
// not use this file except in compliance with the License. You may obtain
// a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
// WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
// License for the specific language governing permissions and limitations
// under the License.

package beegopro

import (
	"database/sql"
	"errors"
	"path"
	"strings"

	beeLogger "github.com/beego/bee/logger"
	"github.com/beego/bee/utils"
)

func (c *Container) renderController(modelName string, content ModelsContent) (err error) {
	switch content.SourceGen {
	case "text":
		c.textRenderController(modelName, content)
		return
	case "database":
		c.databaseRenderController(modelName, content)
		return
	}
	err = errors.New("not support source gen, source gen is " + content.SourceGen)
	return
}

func (c *Container) textRenderController(cname string, content ModelsContent) {
	render := NewGenRender("controllers", cname, c.Option)
	modelPath := path.Join(c.Option.BeegoPath, "models", strings.ToLower(render.Name)+".go")
	if utils.IsExist(modelPath) {
		beeLogger.Log.Infof("Using matching model '%s'", render.Name)
		render.Exec("controllerModel.go.tmpl")
	} else {
		render.Exec("controller.go.tmpl")
	}
	return
}

func (c *Container) databaseRenderController(cname string, content ModelsContent) {
	// todo uniform sql open
	db, err := sql.Open(c.Option.Driver, c.Option.Dsn)
	if err != nil {
		beeLogger.Log.Fatalf("Could not connect to '%s' database using '%s': %s", c.Option.Driver, c.Option.Dsn, err)
		return
	}

	defer db.Close()

	trans, ok := dbDriver[c.Option.Driver]
	if !ok {
		beeLogger.Log.Fatalf("Generating app code from '%s' database is not supported yet.", c.Option.Driver)
		return
	}

	tb := getTableObject(cname, db, trans)
	if tb.Pk == "" {
		return
	}
	render := NewGenRender("controllers", tb.Name, c.Option)

	// 判断是否有models
	modelPath := path.Join(c.Option.BeegoPath, "models", strings.ToLower(render.Name)+".go")
	if utils.IsExist(modelPath) {
		beeLogger.Log.Infof("Using matching model '%s'", render.Name)
		render.Exec("controllerModel.go.tmpl")
	} else {
		render.Exec("controller.go.tmpl")
	}
	return
}
