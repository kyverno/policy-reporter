package config

import "github.com/kyverno/policy-reporter/pkg/target"

type ValueFilter struct {
	Include  []string       `mapstructure:"include"`
	Exclude  []string       `mapstructure:"exclude"`
	Selector map[string]any `mapstructure:"selector"`
}

type EmailReportFilter struct {
	DisableClusterReports bool        `mapstructure:"disableClusterReports"`
	Namespaces            ValueFilter `mapstructure:"namespaces"`
	Sources               ValueFilter `mapstructure:"sources"`
}

type TargetFilter struct {
	Namespaces   ValueFilter `mapstructure:"namespaces"`
	Priorities   ValueFilter `mapstructure:"priorities"`
	Policies     ValueFilter `mapstructure:"policies"`
	ReportLabels ValueFilter `mapstructure:"reportLabels"`
}

type MetricsFilter struct {
	Namespaces ValueFilter `mapstructure:"namespaces"`
	Policies   ValueFilter `mapstructure:"policies"`
	Severities ValueFilter `mapstructure:"severities"`
	Status     ValueFilter `mapstructure:"status"`
	Sources    ValueFilter `mapstructure:"sources"`
	Kinds      ValueFilter `mapstructure:"kinds"`
}

// SMTP configuration
type SMTP struct {
	Host        string `mapstructure:"host"`
	Port        int    `mapstructure:"port"`
	Username    string `mapstructure:"username"`
	Password    string `mapstructure:"password"`
	From        string `mapstructure:"from"`
	Encryption  string `mapstructure:"encryption"`
	SkipTLS     bool   `mapstructure:"skipTLS"`
	Certificate string `mapstructure:"certificate"`
}

// EmailReport configuration
type EmailReport struct {
	To       []string          `mapstructure:"to"`
	Format   string            `mapstructure:"format"`
	Filter   EmailReportFilter `mapstructure:"filter"`
	Channels []EmailReport     `mapstructure:"channels"`
}

// EmailReport configuration
type Templates struct {
	Dir string `mapstructure:"dir"`
}

// EmailReports configuration
type EmailReports struct {
	SMTP        SMTP        `mapstructure:"smtp"`
	Summary     EmailReport `mapstructure:"summary"`
	Violations  EmailReport `mapstructure:"violations"`
	ClusterName string      `mapstructure:"clusterName"`
	TitlePrefix string      `mapstructure:"titlePrefix"`
}

// BasicAuth configuration
type BasicAuth struct {
	Username  string `mapstructure:"username"`
	Password  string `mapstructure:"password"`
	SecretRef string `mapstructure:"secretRef"`
}

// API configuration
type API struct {
	Port      int       `mapstructure:"port"`
	Logging   bool      `mapstructure:"logging"`
	BasicAuth BasicAuth `mapstructure:"basicAuth"`
	DebugMode bool      `mapstructure:"debug"`
}

// REST configuration
type REST struct {
	Enabled bool `mapstructure:"enabled"`
}

// Metrics configuration
type Metrics struct {
	Filter       MetricsFilter `mapstructure:"filter"`
	CustomLabels []string      `mapstructure:"customLabels"`
	Mode         string        `mapstructure:"mode"`
	Enabled      bool          `mapstructure:"enabled"`
}

// Profiling configuration
type Profiling struct {
	Enabled bool `mapstructure:"enabled"`
}

// ReportFilter configuration
type ReportFilter struct {
	Namespaces            ValueFilter `mapstructure:"namespaces"`
	Sources               ValueFilter `mapstructure:"sources"`
	Kinds                 ValueFilter `mapstructure:"kinds"`
	DisableClusterReports bool        `mapstructure:"disableClusterReports"`
}

// Redis configuration
type Redis struct {
	Enabled  bool   `mapstructure:"enabled"`
	Address  string `mapstructure:"address"`
	Prefix   string `mapstructure:"prefix"`
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
	Database int    `mapstructure:"database"`
}

// LeaderElection configuration
type LeaderElection struct {
	LockName        string `mapstructure:"lockName"`
	PodName         string `mapstructure:"podName"`
	Namespace       string `mapstructure:"namespace"`
	LeaseDuration   int    `mapstructure:"leaseDuration"`
	RenewDeadline   int    `mapstructure:"renewDeadline"`
	RetryPeriod     int    `mapstructure:"retryPeriod"`
	ReleaseOnCancel bool   `mapstructure:"releaseOnCancel"`
	Enabled         bool   `mapstructure:"enabled"`
}

// K8sClient config struct
type K8sClient struct {
	QPS        float32 `mapstructure:"qps"`
	Burst      int     `mapstructure:"burst"`
	Kubeconfig string  `mapstructure:"kubeconfig"`
}

type Logging struct {
	LogLevel    int8   `mapstructure:"logLevel"`
	Encoding    string `mapstructure:"encoding"`
	Development bool   `mapstructure:"development"`
}

type Database struct {
	Type          string `mapstructure:"type"`
	DSN           string `mapstructure:"dsn"`
	Username      string `mapstructure:"username"`
	Password      string `mapstructure:"password"`
	Database      string `mapstructure:"database"`
	Host          string `mapstructure:"host"`
	EnableSSL     bool   `mapstructure:"enableSSL"`
	SecretRef     string `mapstructure:"secretRef"`
	MountedSecret string `mapstructure:"mountedSecret"`
}

type SourceSelector struct {
	Source string `mapstructure:"source"`
}

type SourceFilter struct {
	Selector              SourceSelector `mapstructure:"selector"`
	Kinds                 ValueFilter    `mapstructure:"kinds"`
	Sources               ValueFilter    `mapstructure:"sources"`
	Namespaces            ValueFilter    `mapstructure:"namespaces"`
	UncontrolledOnly      bool           `mapstructure:"uncontrolledOnly"`
	DisableClusterReports bool           `mapstructure:"disableClusterReports"`
}

type CustomID struct {
	Enabled bool     `mapstructure:"enabled"`
	Fields  []string `mapstructure:"fields"`
}

type SourceConfig struct {
	Selector SourceSelector `mapstructure:"selector"`
	CustomID `mapstructure:"customId"`
}

// Config of the PolicyReporter
type Config struct {
	Version        string
	Namespace      string         `mapstructure:"namespace"`
	API            API            `mapstructure:"api"`
	WorkerCount    int            `mapstructure:"worker"`
	DBFile         string         `mapstructure:"dbfile"`
	Metrics        Metrics        `mapstructure:"metrics"`
	REST           REST           `mapstructure:"rest"`
	ReportFilter   ReportFilter   `mapstructure:"reportFilter"`
	SourceFilters  []SourceFilter `mapstructure:"sourceFilters"`
	Redis          Redis          `mapstructure:"redis"`
	Profiling      Profiling      `mapstructure:"profiling"`
	EmailReports   EmailReports   `mapstructure:"emailReports"`
	LeaderElection LeaderElection `mapstructure:"leaderElection"`
	K8sClient      K8sClient      `mapstructure:"k8sClient"`
	Logging        Logging        `mapstructure:"logging"`
	Database       Database       `mapstructure:"database"`
	Targets        target.Targets `mapstructure:"target"`
	SourceConfig   []SourceConfig `mapstructure:"sourceConfig"`
	Templates      Templates      `mapstructure:"templates"`
}
