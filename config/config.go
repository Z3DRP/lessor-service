package config

import (
	"errors"
	"fmt"
	"log"

	"github.com/spf13/viper"
)

type ErrConfigRead struct {
	FileType     string
	Path         string
	ConfigObject string
	Err          error
}

func (ec *ErrConfigRead) Error() string {
	return fmt.Sprintf("An error occurred while reading %v config file: FileType: %v, Path: %v :: %v", ec.ConfigObject, ec.FileType, ec.Path, ec.Err)
}

func (ec *ErrConfigRead) Unwrap() error {
	return ec.Err
}

func NewConfigReadError(configObj string, e error) *ErrConfigRead {
	return &ErrConfigRead{
		FileType:     "yaml",
		Path:         "./config",
		ConfigObject: configObj,
		Err:          e,
	}
}

type Configurations struct {
	ZServer        ZServerConfig `mapstructure:"zserver"`
	DatabaseStore  DbConfig      `mapstructure:"database"`
	ZypherSettings ZypherConfig  `mapstructure:"zysettings"`
	ZEmailSettings ZEmailConfig  `mapstructure:"zemailsettings"`
	AuthKey        string        `mapstructure:"authkey"`
	AwsS3Key       string        `mapstructure:"awsS3Key"`
	Salty          string        `mapstructure:"salty"`
}

type ZServerConfig struct {
	Address      string `mapstructure:"address"`
	ReadTimeout  int    `mapstructure:"readTimeout"`
	WriteTimeout int    `mapstructure:"writeTimeout"`
	Fchain       string `mapstructure:"fchain"`
	Pkey         string `mapstructure:"pkey"`
}

type DbConfig struct {
	Provider     string `mapstructure:"provider"`
	DbName       string `mapstructure:"dbName"`
	DbUsr        string `mapstructure:"dbUsr"`
	DbPwd        string `mapstructure:"dbPwd"`
	Host         string `mapstructure:"host"`
	SslRoot      string `mapstructure:"sslRoot"`
	Port         string `mapstructure:"port"`
	DialTimeout  int    `mapstructure:"dialTimeout"`
	WriteTimeout int    `mapstructure:"writeTimeout"`
	ReadTimeout  int    `mapstructure:"readTimeout"`
	MaxOpenConns int    `mapstructure:"maxOpenConns"`
	MaxIdleConns int    `mapstructure:"maxIdleConns"`
	ConnTimeout  int    `mapstructure:"connTimeout"`
}

type ZypherConfig struct {
	Shift        int  `mapstructure:"shift"`
	ShiftCount   int  `mapstructure:"shiftCount"`
	HashCount    int  `mapstructure:"hashCount"`
	Alternate    bool `mapstructure:"alternate"`
	IgnSpace     bool `mapstructure:"ignSpace"`
	RestrictHash bool `mapstructure:"restrictHash"`
}

type ZEmailConfig struct {
	SenderAddress   string `mapstructure:"senderAddress"`
	SenderPwd       string `mapstructure:"senderPwd"`
	RecieverAddress string `mapstructure:"recieverAddress"`
	SmtpServer      string `mapstructure:"smtpServer"`
	SmtpPort        int    `mapstructure:"smtpPort"`
}

func ReadConfig(configPath string) (*Configurations, error) {
	viper.SetConfigType("yaml")
	viper.SetConfigName("config")
	//viper.AddConfigPath("./config")
	viper.AddConfigPath(configPath)
	viper.AutomaticEnv()
	var configs Configurations

	if err := viper.ReadInConfig(); err != nil {
		emsg := fmt.Sprintf("error reading config file, %v", err)
		return nil, errors.New(emsg)
	}

	err := viper.Unmarshal(&configs)
	if err != nil {
		emsg := fmt.Sprintf("unable to decode config to json, %v", err)
		return nil, errors.New(emsg)
	}
	return &configs, nil
}

func ReadZypherSettings() (ZypherConfig, error) {
	viper.SetConfigType("yaml")
	viper.SetConfigName("config")
	viper.AddConfigPath("./config")
	viper.AutomaticEnv()
	var configs Configurations

	if err := viper.ReadInConfig(); err != nil {
		emsg := fmt.Sprintf("error reading config file, %v", err)
		log.Printf("read config err: %v\n", err)
		return ZypherConfig{}, errors.New(emsg)
	}

	err := viper.Unmarshal(&configs)
	if err != nil {
		emsg := fmt.Sprintf("unable to decode config to json:: %v", err)
		log.Printf("err processing config: %v\n", err)
		return ZypherConfig{}, errors.New(emsg)
	}
	return configs.ZypherSettings, nil
}

func ReadEmailConfig() (*ZEmailConfig, error) {
	viper.SetConfigType("yaml")
	viper.SetConfigName("config")
	viper.AddConfigPath("./config")
	var configs Configurations

	if err := viper.ReadInConfig(); err != nil {
		emsg := fmt.Sprintf("error reading config file, %v", err)
		return nil, errors.New(emsg)
	}

	err := viper.Unmarshal(&configs)
	if err != nil {
		emsg := fmt.Sprintf("unable to decode config to json:: %v", err)
		return nil, errors.New(emsg)
	}
	return &configs.ZEmailSettings, nil
}

func GetAuthToken() ([]byte, error) {
	if err := setupConfig(); err != nil {
		log.Printf("failed auth setup: %v\n", err)
		return nil, err
	}

	var configs Configurations

	if err := viper.ReadInConfig(); err != nil {
		emsg := fmt.Sprintf("error reading config file, %v", err)
		log.Printf("viper read config err: %v\n", err)
		return nil, errors.New(emsg)
	}
	return []byte(configs.AuthKey), nil
}

func GetS3Key() (string, error) {
	if err := setupConfig(); err != nil {
		log.Printf("failed s3 setup %v\n", err)
		return "", err
	}

	var configs Configurations

	if err := viper.ReadInConfig(); err != nil {
		emsg := fmt.Sprintf("error reading config file, %v", err)
		log.Printf("viper read config err: %v\n", err)
		return "", errors.New(emsg)
	}

	return configs.AwsS3Key, nil
}

func GetSalty() (string, error) {
	if err := setupConfig(); err != nil {
		log.Printf("failed salty setup %v\n", err)
		return "", err
	}

	var configs Configurations

	if err := viper.ReadInConfig(); err != nil {
		emsg := fmt.Sprintf("error reading config file, %v", err)
		log.Printf("viper read config err: %v\n", err)
		return "", errors.New(emsg)
	}

	return configs.Salty, nil
}

func IsValidOrigin(origin string) bool {
	return true
	//	validOrigin := map[string]bool{
	//		"http://localhost:3000":      true,
	//		"https://localhost:3000":     true,
	//		"http://zrp3.dev":      true,
	//		"https://zrp3.dev":     true,
	//		"http://www.zrp3.dev":  true,
	//		"https://www.zrp3.dev": true,
	//		"www.zrp3.dev":         true,
	//	}
	//
	// return validOrigin[origin]
}

func setupConfig() error {
	viper.SetConfigType("yaml")
	viper.SetConfigName("config")
	viper.AddConfigPath("./config")
	if err := viper.ReadInConfig(); err != nil {
		emsg := fmt.Sprintf("error reading config file, %v", err)
		log.Printf("failed config setup: %v\n", err)
		return errors.New(emsg)
	}
	return nil

}
