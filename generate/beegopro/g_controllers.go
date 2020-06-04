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
	"github.com/flosch/pongo2"
	"github.com/smartwalle/pongo2render"
	"path"
	"strings"

	beeLogger "github.com/beego/bee/logger"
	"github.com/beego/bee/utils"
)

func (c *Container) renderController(cname, fields string) (err error) {
	var render = pongo2render.NewRender(c.Option.GitPath + "/" + c.Option.ProType)

	p, f := path.Split(cname)
	controllerName := strings.Title(f)
	packageName := "controllers"

	if p != "" {
		i := strings.LastIndex(p[:len(p)-1], "/")
		packageName = p[i+1 : len(p)-1]
	}

	beeLogger.Log.Infof("Using '%s' as controller name", controllerName)
	beeLogger.Log.Infof("Using '%s' as package name", packageName)

	fp := path.Join(c.Option.BeegoPath, "controllers", p)
	err = createPath(fp)
	if err != nil {
		beeLogger.Log.Fatalf("Could not create the controllers directory: %s", err)
		return
	}

	fpath := path.Join(fp, strings.ToLower(controllerName)+".go")
	pkgPath := getPackagePath(c.Option.BeegoPath)

	ctx := pongo2.Context{
		"packageName":    packageName,
		"controllerName": controllerName,
		"pkgPath":        pkgPath,
		"apiPrefix":      c.Option.ApiPrefix,
	}

	var (
		buf string
	)
	// 判断是否有models
	modelPath := path.Join(c.Option.BeegoPath, "models", strings.ToLower(controllerName)+".go")
	if utils.IsExist(modelPath) {
		beeLogger.Log.Infof("Using matching model '%s'", controllerName)
		buf, err = render.Template("controllers/controllerModel.go.tmpl").Execute(ctx)
		if err != nil {
			beeLogger.Log.Fatalf("Could not create the model render tmpl: %s", err)
			return
		}
	} else {
		buf, err = render.Template("controllers/controller.go.tmpl").Execute(ctx)
		if err != nil {
			beeLogger.Log.Fatalf("Could not create the model render tmpl: %s", err)
			return
		}
	}
	err = c.write(fpath, buf)
	if err != nil {
		beeLogger.Log.Fatalf("Could not create model file: %s", err)
		return
	}
	beeLogger.Log.Infof("create file '%s'", fpath)
	return
}
