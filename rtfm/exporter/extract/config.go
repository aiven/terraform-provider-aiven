package extract

type Config struct {
	TerraformBinaryPath string `envconfig:"TERRAFORM_BINARY_PATH" required:"true" default:"/opt/homebrew/bin/terraform"`
	ExportRootPath      string `envconfig:"EXPORT_ROOT_PATH"`
	PluginCacheDir      string `envconfig:"TF_PLUGIN_CACHE_DIR" default:"~/.terraform.d/plugin-cache"`
}
