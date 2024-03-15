package conf

type AppConfig struct {
	Redis     RedisConf     `yaml:"redis"`
	Jwt       JwtConf       `yaml:"jwt"`
	ComfyUI   ComfyUIConf   `yaml:"comfy-ui" mapstructure:"comfy-ui"`
	Bytedance BytedanceConf `yaml:"bytedance"`
	Baidu     BaiduConf     `yaml:"baidu"`
	OpenAI    OpenAIConf    `yaml:"openai"`
}

type JwtConf struct {
	PrivateKey string `yaml:"private-key" mapstructure:"private-key"`
	PublicKey  string `yaml:"public-key" mapstructure:"public-key"`
}

type LimitConf struct {
	Rate       int      `yaml:"rate"`
	Bucket     int      `yaml:"bucket"`
	VipRate    int      `yaml:"vip-rate" mapstructure:"vip-rate"`
	VipBucket  int      `yaml:"vip-bucket" mapstructure:"vip-bucket"`
	Predicates []string `yaml:"predicates"`
}

type RedisConf struct {
	Address  string `yaml:"address"`
	Database int    `yaml:"database"`
}

type ComfyUIConf struct {
	Location string    `yaml:"location"`
	Limit    LimitConf `yaml:"limit"`
	Address  []string  `yaml:"address"`
}

type BytedanceConf struct {
	Location      string    `yaml:"location"`
	Limit         LimitConf `yaml:"limit"`
	Address       string    `yaml:"address"`
	AppId         string    `yaml:"app-id" mapstructure:"app-id"`
	Authorization string    `yaml:"authorization"`
}

type BaiduConf struct {
	Location     string    `yaml:"location"`
	TokenFile    string    `yaml:"token-file" mapstructure:"token-file"`
	Limit        LimitConf `yaml:"limit"`
	Address      string    `yaml:"address"`
	ClientId     string    `yaml:"client-id" mapstructure:"client-id"`
	ClientSecret string    `yaml:"client-secret" mapstructure:"client-secret"`
}

type OpenAIConf struct {
	Location      string    `yaml:"location"`
	Limit         LimitConf `yaml:"limit"`
	Address       string    `yaml:"address"`
	Authorization string    `yaml:"authorization"`
}
