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

package generate

import (
	"errors"
	"fmt"
	"go/format"
	"io/ioutil"
	"os"
	"path"
	"strings"

	beeLogger "github.com/beego/bee/logger"
	"github.com/beego/bee/logger/colors"
	"github.com/beego/bee/utils"
	"github.com/flosch/pongo2"
	"github.com/smartwalle/pongo2render"
)

func GenerateModel(mname, fields, currpath string) {
	var render = pongo2render.NewRender("/home/www/server/opensource/bee-mod/default/v1/go/models")

	w := colors.NewColorWriter(os.Stdout)

	p, f := path.Split(mname)
	modelName := strings.Title(f)
	packageName := "models"
	if p != "" {
		i := strings.LastIndex(p[:len(p)-1], "/")
		packageName = p[i+1 : len(p)-1]
	}

	modelStruct, hastime, err := getStruct(modelName, fields)
	if err != nil {
		beeLogger.Log.Fatalf("Could not generate the model struct: %s", err)
	}

	beeLogger.Log.Infof("Using '%s' as model name", modelName)
	beeLogger.Log.Infof("Using '%s' as package name", packageName)

	fp := path.Join(currpath, "models", p)
	if _, err := os.Stat(fp); os.IsNotExist(err) {
		// Create the model's directory
		if err := os.MkdirAll(fp, 0777); err != nil {
			beeLogger.Log.Fatalf("Could not create the model directory: %s", err)
		}
	}

	ctx := pongo2.Context{
		"packageName": packageName,
		"name":        modelName,
		"modelStruct": modelStruct,
	}
	fpath := path.Join(fp, strings.ToLower(modelName)+".go")

	if hastime {
		ctx["timePkg"] = `"time"`
	} else {
		ctx["timePkg"] = ""
	}

	buf, err := render.Template("model.go.tmpl").Execute(ctx)
	if err != nil {
		beeLogger.Log.Fatalf("Could not create the model render tmpl: %s", err)
	}
	err = write(fpath, buf)
	if err != nil {
		beeLogger.Log.Fatalf("Could not create model file: %s", err)
		return
	}
	_, _ = fmt.Fprintf(w, "\t%s%screate%s\t %s%s\n", "\x1b[32m", "\x1b[1m", "\x1b[21m", fpath, "\x1b[0m")
}

func getStruct(structname, fields string) (string, bool, error) {
	if fields == "" {
		return "", false, errors.New("fields cannot be empty")
	}

	hastime := false
	structStr := "type " + structname + " struct{\n"
	fds := strings.Split(fields, ",")
	for i, v := range fds {
		kv := strings.SplitN(v, ":", 2)
		if len(kv) != 2 {
			return "", false, errors.New("the fields format is wrong. Should be key:type,key:type " + v)
		}

		typ, tag, hastimeinner := getType(kv[1])
		if typ == "" {
			return "", false, errors.New("the fields format is wrong. Should be key:type,key:type " + v)
		}

		if i == 0 && strings.ToLower(kv[0]) != "id" {
			structStr = structStr + "Id     int64     `orm:\"auto\"`\n"
		}

		if hastimeinner {
			hastime = true
		}
		structStr = structStr + utils.CamelString(kv[0]) + "       " + typ + "     " + tag + "\n"
	}
	structStr += "}\n"
	return structStr, hastime, nil
}

// fields support type
// http://beego.me/docs/mvc/model/models.md#mysql
func getType(ktype string) (kt, tag string, hasTime bool) {
	kv := strings.SplitN(ktype, ":", 2)
	switch kv[0] {
	case "string":
		if len(kv) == 2 {
			return "string", "`orm:\"size(" + kv[1] + ")\"`", false
		}
		return "string", "`orm:\"size(128)\"`", false
	case "text":
		return "string", "`orm:\"type(longtext)\"`", false
	case "auto":
		return "int64", "`orm:\"auto\"`", false
	case "pk":
		return "int64", "`orm:\"pk\"`", false
	case "datetime":
		return "time.Time", "`orm:\"type(datetime)\"`", true
	case "int", "int8", "int16", "int32", "int64":
		fallthrough
	case "uint", "uint8", "uint16", "uint32", "uint64":
		fallthrough
	case "bool":
		fallthrough
	case "float32", "float64":
		return kv[0], "", false
	case "float":
		return "float64", "", false
	}
	return "", "", false
}

// write 写bytes到文件
func write(filename string, buf string) (err error) {
	if utils.IsExist(filename) {
		err = errors.New("file is exist, path is " + filename)
		return
	}

	filePath := path.Dir(filename)
	err = createPath(filePath)
	if err != nil {
		err = errors.New("write create path " + err.Error())
		return
	}

	file, err := os.Create(filename)
	defer file.Close()
	if err != nil {
		err = errors.New("write create file " + err.Error())
		return
	}

	// 格式化代码
	bts, err := format.Source([]byte(buf))
	if err != nil {
		err = errors.New("format buf error " + err.Error())
		return
	}

	err = ioutil.WriteFile(filename, bts, 0644)
	if err != nil {
		err = errors.New("write write file " + err.Error())
		return
	}
	return
}

// createPath 调用os.MkdirAll递归创建文件夹
func createPath(filePath string) error {
	if !utils.IsExist(filePath) {
		err := os.MkdirAll(filePath, os.ModePerm)
		return err
	}
	return nil
}
