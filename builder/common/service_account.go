package common

//go:generate packer-sdc mapstructure-to-hcl2 -type ServiceAccountConfig

type ServiceAccountConfig struct {
	PrivateKeyFileEnv string `mapstructure:"private_key_file_env" required:"true"`
	PublicKeyIDEnv    string `mapstructure:"public_key_id_env" required:"true"`
	AccountIDEnv      string `mapstructure:"account_id_env" required:"true"`
}
