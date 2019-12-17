//go:generate bash generate/generate.sh
package client

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"runtime"

	"github.com/aws-cloudformation/rain/config"
	"github.com/aws-cloudformation/rain/console"
	"github.com/aws-cloudformation/rain/console/spinner"
	"github.com/aws-cloudformation/rain/version"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/awserr"
	"github.com/aws/aws-sdk-go-v2/aws/defaults"
	"github.com/aws/aws-sdk-go-v2/aws/external"
	"github.com/spf13/viper"
)

func MFAProvider() (string, error) {
	spinner.Pause()
	defer func() {
		fmt.Println()
		spinner.Resume()
	}()

	return console.Ask("MFA Token:"), nil
}

var awsCfg *aws.Config

func tryConfig(configs external.Configs, resolvers []external.AWSConfigResolver) (aws.Config, bool) {
	cfg, err := configs.ResolveAWSConfig(resolvers)
	if err != nil {
		config.Debugf("Credentials failed: %s", err)
		return cfg, false
	} else if _, err = cfg.Credentials.Retrieve(); err != nil {
		config.Debugf("Invalid credentials: %s", err)
		return cfg, false
	}

	return cfg, true
}

func loadConfig() aws.Config {
	var cfg aws.Config
	var err error
	var ok bool

	// Minimal resolvers
	var resolvers = []external.AWSConfigResolver{
		external.ResolveDefaultAWSConfig,
		external.ResolveCustomCABundle,
		external.ResolveRegion,
		external.ResolveFallbackEC2Credentials, // Initial defauilt credentails provider.
		external.ResolveEndpointCredentials,
		external.ResolveContainerEndpointPathCredentials, // TODO is this order right?
	}

	// Minimal configs
	var configs external.Configs = []external.Config{}
	if config.Profile != "" {
		configs = append(configs, external.WithSharedConfigProfile(config.Profile))
	} else if os.Getenv("AWS_PROFILE") != "" {
		config.Profile = os.Getenv("AWS_PROFILE")
	}
	if config.Region != "" {
		configs = append(configs, external.WithRegion(config.Region))
	}
	configs, err = configs.AppendFromLoaders(external.DefaultConfigLoaders)
	if err != nil {
		panic(err)
	}

	// Try loading cached credentials
	if viper.IsSet("credentials") {
		config.Debugf("Found cached credentials...")

		if viper.GetString("profile") != config.Profile {
			config.Debugf("...but they don't match the requested profile")
		} else {
			var creds aws.Credentials

			if json.Unmarshal([]byte(viper.GetString("credentials")), &creds) != nil {
				config.Debugf("...but I couldn't load them: %s", err)
			} else {
				if creds.Expired() {
					config.Debugf("...but they've expired")
				} else {
					configs = append(configs, external.WithCredentialsValue(creds))

					resolvers = append(resolvers, external.ResolveCredentialsValue)

					if cfg, ok = tryConfig(configs, resolvers); ok {
						config.Debugf("...and they've valid :)")
						return cfg
					}
				}
			}
		}
	}

	configs = append(configs, external.WithMFATokenFunc(MFAProvider))
	resolvers = append(resolvers, external.ResolveAssumeRoleCredentials)

	config.Debugf("Trying default configs...")
	if cfg, ok = tryConfig(configs, resolvers); ok {
		config.Debugf("...and they're valid")
		return cfg
	}

	panic("Unable to find valid credentials")
}

func Config() aws.Config {
	if awsCfg == nil {
		spinner.Status("Loading AWS config...")

		cfg := loadConfig()

		// Save the creds
		creds, _ := cfg.Credentials.Retrieve()
		j, err := json.Marshal(creds)
		if err != nil {
			config.Debugf("Unable to save credentials: %s", err)
		} else {
			viper.Set("credentials", string(j))
			viper.Set("profile", config.Profile)
			viper.WriteConfig()
		}

		// Set the user agent
		cfg.Handlers.Build.Remove(defaults.SDKVersionUserAgentHandler)
		cfg.Handlers.Build.PushFront(aws.MakeAddToUserAgentHandler(
			version.NAME,
			version.VERSION,
			runtime.Version(),
			fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH),
			fmt.Sprintf("%s/%s", aws.SDKName, aws.SDKVersion),
		))

		// For debugging
		// cfg.EndpointResolver = aws.ResolveWithEndpointURL("http://localhost:8000")

		awsCfg = &cfg

		spinner.Stop()
	}

	return *awsCfg
}

func SetRegion(region string) {
	awsCfg.Region = region
}

type Error error

func NewError(err error) Error {
	if err == nil {
		return nil
	}

	if err, ok := err.(awserr.Error); ok {
		return Error(errors.New(err.Message()))
	}

	return Error(err)
}
