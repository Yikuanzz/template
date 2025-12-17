package plugins

import (
	"backend/app/plugins/db"

	"go.uber.org/fx"
)

// ProvidePlugins 提供所有插件
var PluginsModule = fx.Module("plugins",
	fx.Provide(
		// Database
		db.ProvideDatabase,
	),
)
