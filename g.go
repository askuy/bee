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

package main

import "os"

var cmdGenerate = &Command{
	UsageLine: "generate [Command]",
	Short:     "generate code based on application",
	Long: `
bee generate model [-driver=mysql] [-conn=root:@tcp(127.0.0.1:3306)/test] [-level=1]
    generate model based on an existing database
    -driver: [mysql | postgresql | sqlite], the default is mysql
    -conn:   the connection string used by the driver, the default is root:@tcp(127.0.0.1:3306)/test
    -level:  [1 | 2 | 3], 1 = model; 2 = models,controller; 3 = models,controllers,router

bee generate controller [modelfile]
    generate RESTFul controllers based on modelfile             

bee generate router [controllerfile]
    generate router based on controllerfile

bee generate docs
    generate swagger doc file

bee generate test [routerfile]
    generate testcase
`,
}

var driver docValue
var conn docValue
var level docValue

func init() {
	cmdGenerate.Run = generateCode
	cmdGenerate.Flag.Var(&driver, "driver", "database driver: mysql, postgresql, etc.")
	cmdGenerate.Flag.Var(&conn, "conn", "connection string used by the driver to connect to a database instance")
	cmdGenerate.Flag.Var(&level, "level", "1 = models only; 2 = models and controllers; 3 = models, controllers and routers")
}

func generateCode(cmd *Command, args []string) {
	curpath, _ := os.Getwd()
	if len(args) < 1 {
		ColorLog("[ERRO] command is missing\n")
		os.Exit(2)
	}

	gopath := os.Getenv("GOPATH")
	Debugf("gopath:%s", gopath)
	if gopath == "" {
		ColorLog("[ERRO] $GOPATH not found\n")
		ColorLog("[HINT] Set $GOPATH in your environment vairables\n")
		os.Exit(2)
	}

	gcmd := args[0]
	switch gcmd {
	case "docs":
		generateDocs(curpath)
	case "model":
		cmd.Flag.Parse(args[1:])
		if driver == "" {
			driver = "mysql"
		}
		if conn == "" {
			conn = "root:@tcp(127.0.0.1:3306)/test"
		}
		if level == "" {
			level = "1"
		}
		ColorLog("[INFO] Using '%s' as 'driver'\n", driver)
		ColorLog("[INFO] Using '%s' as 'conn'\n", conn)
		ColorLog("[INFO] Using '%s' as 'level'\n", level)
		generateModel(string(driver), string(conn), string(level), curpath)
	default:
		ColorLog("[ERRO] command is missing\n")
	}
	ColorLog("[SUCC] generate successfully created!\n")
}