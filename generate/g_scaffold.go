package generate

import (
	"github.com/beego/bee/generate/beegopro"
	"strings"

	"github.com/beego/bee/cmd/commands/migrate"
	beeLogger "github.com/beego/bee/logger"
	"github.com/beego/bee/utils"
)

func GenerateScaffold(sname, fields, currpath, driver, conn string) {
	beeLogger.Log.Infof("Do you want to create a '%s' model? [Yes|No] ", sname)

	// Generate the model
	if utils.AskForConfirmation() {
		schemas := make([]beegopro.Schema, 0)
		fds := strings.Split(fields, ",")
		for _, v := range fds {
			kv := strings.SplitN(v, ":", 2)
			if len(kv) != 2 {
				beeLogger.Log.Error("Fields format is wrong. Should be: key:type,key:type " + v)
				return
			}
			schemas = append(schemas, beegopro.Schema{
				Name: kv[0],
				Type: kv[1],
			})
		}

		beegopro.DefaultBeegoPro.TextRenderModel(sname, beegopro.ModelsContent{
			Schema:    schemas,
			SourceGen: "text",
			ApiPrefix: "/",
		})
		//GenerateModel(sname, fields, currpath)
	}

	// Generate the controller
	beeLogger.Log.Infof("Do you want to create a '%s' controller? [Yes|No] ", sname)
	if utils.AskForConfirmation() {
		GenerateController(sname, currpath)
	}

	// Generate the views
	beeLogger.Log.Infof("Do you want to create views for this '%s' resource? [Yes|No] ", sname)
	if utils.AskForConfirmation() {
		GenerateView(sname, currpath)
	}

	// Generate a migration
	beeLogger.Log.Infof("Do you want to create a '%s' migration and schema for this resource? [Yes|No] ", sname)
	if utils.AskForConfirmation() {
		upsql := ""
		downsql := ""
		if fields != "" {
			dbMigrator := NewDBDriver()
			upsql = dbMigrator.GenerateCreateUp(sname)
			downsql = dbMigrator.GenerateCreateDown(sname)
		}
		GenerateMigration(sname, upsql, downsql, currpath)
	}

	// Run the migration
	beeLogger.Log.Infof("Do you want to migrate the database? [Yes|No] ")
	if utils.AskForConfirmation() {
		migrate.MigrateUpdate(currpath, driver, conn, "")
	}
	beeLogger.Log.Successf("All done! Don't forget to add  beego.Router(\"/%s\" ,&controllers.%sController{}) to routers/route.go\n", sname, strings.Title(sname))
}
