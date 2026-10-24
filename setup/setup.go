package main

import (
	"os"

	"github.com/goravel/framework/packages"
	"github.com/goravel/framework/packages/match"
	"github.com/goravel/framework/packages/modify"
	"github.com/goravel/framework/support/env"
	"github.com/goravel/framework/support/path"
)

func main() {
	setup := packages.Setup(os.Args)
	moduleImport := setup.Paths().Module().Import()
	provider := "&goralens.ServiceProvider{}"
	migration := "&goralens.Migration{}"
	appConfigPath := path.Config("app.go")
	configPath := path.Config("goralens.go")
	loggingConfigPath := path.Config("logging.go")
	logChannels := match.Config("logging.channels")
	logChannel := `map[string]any{
            "driver": "custom",
            "via":    goralens.LogDriver(),
        }`
	envLines := "\nGORALENS_ENABLED=true\n"

	setup.Install(

		modify.When(func(_ map[string]any) bool {
			return env.IsBootstrapSetup()
		}, modify.RegisterProvider(moduleImport, provider)),
		modify.When(func(_ map[string]any) bool {
			return !env.IsBootstrapSetup()
		}, modify.GoFile(appConfigPath).
			Find(match.Imports()).Modify(modify.AddImport(moduleImport)).
			Find(match.Providers()).Modify(modify.Register(provider))),

		modify.When(func(_ map[string]any) bool {
			return env.IsBootstrapSetup()
		}, modify.RegisterMigration(moduleImport, migration)),

		modify.WhenFileNotExists(configPath,
			modify.File(configPath).Overwrite(configStub(setup.Paths().Config().Package(), setup.Paths().Facades().Import(), setup.Paths().Facades().Package()))),

		modify.WhenFileExists(loggingConfigPath, modify.GoFile(loggingConfigPath).
			Find(match.Imports()).Modify(modify.AddImport(moduleImport)).
			Find(logChannels).Modify(modify.AddConfig("goralens", logChannel))),

		modify.WhenFileNotContains(path.Base(".env"), "GORALENS_ENABLED", modify.File(path.Base(".env")).Append(envLines)),
		modify.WhenFileNotContains(path.Base(".env.example"), "GORALENS_ENABLED", modify.File(path.Base(".env.example")).Append(envLines)),
	).Uninstall(
		modify.WhenFileExists(loggingConfigPath, modify.GoFile(loggingConfigPath).
			Find(logChannels).Modify(modify.RemoveConfig("goralens")).
			Find(match.Imports()).Modify(modify.RemoveImport(moduleImport))),
		modify.File(configPath).Remove(),
		modify.When(func(_ map[string]any) bool {
			return env.IsBootstrapSetup()
		}, modify.UnregisterMigration(moduleImport, migration)),
		modify.When(func(_ map[string]any) bool {
			return env.IsBootstrapSetup()
		}, modify.UnregisterProvider(moduleImport, provider)),
		modify.When(func(_ map[string]any) bool {
			return !env.IsBootstrapSetup()
		}, modify.GoFile(appConfigPath).
			Find(match.Providers()).Modify(modify.Unregister(provider)).
			Find(match.Imports()).Modify(modify.RemoveImport(moduleImport))),
	).Execute()
}
