package config

// CONF 配置信息，全局可访问的配置变量
var CONF AppConfig

// Version 当前应用版本号 my-version
const Version = "1.3.3"

type AppConfig struct {
	App struct {
		Host        string `mapstructure:"host" json:"host" yaml:"host"`                        // 应用主机地址，用于绑定服务
		Port        string `mapstructure:"port" json:"port" yaml:"port"`                        // HTTP服务端口
		SslPort     string `mapstructure:"ssl-port" json:"sslPort" yaml:"ssl-port"`             // HTTPS服务端口
		EnableUI    bool   `mapstructure:"enable-ui" json:"enableUI" yaml:"enable-ui"`          // 是否启用UI界面
		ApiKey      string `mapstructure:"api-key" json:"apiKey" yaml:"api-key"`                // API访问密钥
		ApiUser     uint   `mapstructure:"api-user" json:"apiUser" yaml:"api-user"`             // API用户ID
		Type        string `mapstructure:"type" json:"type" yaml:"type"`                        // 应用类型
		ContextPath string `mapstructure:"context-path" json:"contextPath" yaml:"context-path"` // 应用上下文路径
	} `mapstructure:"app" json:"app" yaml:"app"` // 应用基本配置

	Term struct {
		Command string `mapstructure:"command" json:"command" yaml:"command"` // 终端命令配置
	} // 终端相关配置

	Debug bool `mapstructure:"debug" json:"debug" yaml:"debug"` // 是否启用调试模式

	Db struct {
		Type         string `mapstructure:"type" json:"type" yaml:"type"`                             // 数据库类型：mysql、sqlite、postgres
		Host         string `mapstructure:"host" json:"host" yaml:"host"`                             // 数据库服务器地址
		Port         string `mapstructure:"port" json:"port" yaml:"port"`                             // 数据库服务端口
		Config       string `mapstructure:"config" json:"config" yaml:"config"`                       // 数据库高级配置参数
		DbName       string `mapstructure:"db-name" json:"dbname" yaml:"db-name"`                     // 数据库名称
		Username     string `mapstructure:"username" json:"username" yaml:"username"`                 // 数据库用户名
		Password     string `mapstructure:"password" json:"password" yaml:"password"`                 // 数据库密码
		MaxIdleConns int    `mapstructure:"max-idle-conns" json:"maxIdleConns" yaml:"max-idle-conns"` // 空闲中的最大连接数
		MaxOpenConns int    `mapstructure:"max-open-conns" json:"maxOpenConns" yaml:"max-open-conns"` // 打开到数据库的最大连接数
		MaxLifeTime  int    `mapstructure:"max-life-time" json:"maxLifeTime" yaml:"max-life-time"`    // 连接的最大生命周期
		LogMode      string `mapstructure:"log-mode" json:"logMode" yaml:"log-mode"`                  // 是否开启Gorm全局日志
		LogZap       bool   `mapstructure:"log-zap" json:"logZap" yaml:"log-zap"`                     // 是否通过zap写入日志文件
		DataKey      string `mapstructure:"data-key" json:"dataKey" yaml:"data-key"`                  // 数据加密秘钥，使用后不要修改否则数据无法解密
	} `mapstructure:"db" json:"db" yaml:"db"` // 数据库相关配置

	Redis struct {
		Enable   bool   `mapstructure:"enable" json:"enable" yaml:"enable"`       // 是否启用Redis
		Host     string `mapstructure:"host" json:"host" yaml:"host"`             // Redis服务器地址
		Port     string `mapstructure:"port" json:"port" yaml:"port"`             // Redis服务端口
		Password string `mapstructure:"password" json:"password" yaml:"password"` // Redis密码
		Db       int    `mapstructure:"db" json:"db" yaml:"db"`                   // Redis数据库编号
	} `mapstructure:"redis" json:"redis" yaml:"redis"` // Redis相关配置

	Jwt struct {
		AppKey  string `mapstructure:"app-key" json:"appkey" yaml:"app-key"`  // JWT应用密钥
		Expire  int64  `mapstructure:"expire" json:"expire" yaml:"expire"`    // JWT过期时间
		Subject string `mapstructure:"subject" json:"subject" yaml:"subject"` // JWT主题
		Issuer  string `mapstructure:"issuer" json:"issuer" yaml:"issuer"`    // JWT签发者
	} `mapstructure:"jwt" json:"jwt" yaml:"jwt"` // JWT相关配置

	FileUpload struct {
		MaxSize int64 `mapstructure:"max-size" json:"maxSize" yaml:"max-size"` // 文件上传最大大小
	} `mapstructure:"file-upload" json:"fileUpload" yaml:"file-upload"` // 文件上传相关配置

	RClone struct {
		BinPath    string `mapstructure:"bin-path" json:"binPath" yaml:"bin-path"`          // RClone二进制文件路径
		ConfigPath string `mapstructure:"config-path" json:"configPath" yaml:"config-path"` // RClone配置文件路径
		Port       uint   `mapstructure:"port" json:"port" yaml:"port"`                     // RClone服务端口
		CmdTimeOut uint   `mapstructure:"cmd-timeout" json:"cmdTimeout" yaml:"cmdTimeout"`  // RClone命令超时时间
	} `mapstructure:"rclone" json:"rclone" yaml:"rclone"` // RClone相关配置

	Md5 struct {
		Hash string `mapstructure:"hash" json:"hash" yaml:"hash"` // MD5哈希值
	} `mapstructure:"md5" json:"md5" yaml:"md5"` // MD5相关配置

	LogConfig `mapstructure:"log" json:"log" yaml:"log"` // 日志相关配置
}
