package vault

type Addresses = []*Address

type Address struct {
	Scheme string `mapstructure:"scheme"`
	Host   string `mapstructure:"host"`
	Port   uint   `mapstructure:"port"`
}
