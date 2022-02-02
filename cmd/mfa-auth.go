package cmd

import (
	"context"
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	iamTypes "github.com/aws/aws-sdk-go-v2/service/iam/types"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/aws/aws-sdk-go-v2/service/sts/types"
	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
	"gopkg.in/ini.v1"
)

var (
	mfaDeviceSerialNum string
	tokenCode          string
	mfaAuthCmd         = &cobra.Command{
		Use:   "mfa-auth",
		Short: "Authenticate with MFA device",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithSharedConfigProfile(profile))
			if err != nil {
				return err
			}

			return perform2faLogin(cfg, tokenCode)
		},
	}
)

func perform2faLogin(cfg aws.Config, mfaCode string) error {
	mfaArn, err := getEffectiveMfaArn(cfg)
	if err != nil {
		return err
	}

	token, err := getProfileSessionToken(cfg, mfaArn, mfaCode)
	if err != nil {
		return err
	}

	if err := saveStsCredentialsInProfile(token); err != nil {
		return err
	}

	fmt.Println("Retrieved sts credentials and saved to profile:", stsProfileName(), "for", durationSeconds, "seconds")

	return nil
}

func getEffectiveMfaArn(cfg aws.Config) (string, error) {
	if len(mfaDeviceSerialNum) < 1 {
		return defaultMfaDeviceArn(cfg, profile)
	}

	return mfaDeviceSerialNum, nil
}

func stsProfileName() string {
	if len(profile) < 1 {
		return "default_sts"
	}

	return fmt.Sprintf("%s_sts", profile)
}

func saveStsCredentialsInProfile(creds *types.Credentials) error {
	cf, err := readAwsCredentials()
	if err != nil {
		return err
	}

	s := cf.Section(stsProfileName())
	s.Key("aws_access_key_id").SetValue(*creds.AccessKeyId)
	s.Key("aws_secret_access_key").SetValue(*creds.SecretAccessKey)
	s.Key("aws_session_token").SetValue(*creds.SessionToken)

	credsPath, err := getAwsConfigFilePath("credentials")
	if err != nil {
		return err
	}

	return cf.SaveTo(credsPath)
}

func getMfaDevices(cfg aws.Config) ([]iamTypes.MFADevice, error) {
	c := iam.NewFromConfig(cfg)

	userResp, err := c.GetUser(context.TODO(), &iam.GetUserInput{})
	if err != nil {
		return nil, err
	}

	resp, err := c.ListMFADevices(context.TODO(), &iam.ListMFADevicesInput{
		UserName: userResp.User.UserName,
	})
	if err != nil {
		return nil, err
	}

	return resp.MFADevices, nil
}

func promptDeviceSelection(devices []iamTypes.MFADevice) (string, error) {
	var deviceArns []string
	for i := 0; i < len(devices); i++ {
		deviceArns[i] = *devices[i].SerialNumber
	}
	prompt := promptui.Select{
		Label: "Choose MFA Device",
		Items: deviceArns,
	}
	_, result, err := prompt.Run()
	if err != nil {
		return "", err
	}

	return result, nil
}

func discoverMfaDevice(cfg aws.Config, cf *ini.File, ps *ini.Section) (string, error) {
	devices, err := getMfaDevices(cfg)
	if err != nil {
		return "", err
	}

	if len(devices) < 1 {
		return "", errors.New("unable to discover MFA devices for current user")
	}

	configPath, err := getAwsConfigFilePath("config")
	if err != nil {
		return "", err
	}

	var deviceArn string
	if len(devices) == 1 {
		deviceArn = *devices[0].SerialNumber
	} else {
		deviceArn, err = promptDeviceSelection(devices)
		if err != nil {
			return "", err
		}
	}

	ps.Key("mfa").SetValue(deviceArn)
	if err = cf.SaveTo(configPath); err != nil {
		return "", err
	}

	return deviceArn, nil
}

func defaultMfaDeviceArn(cfg aws.Config, profile string) (string, error) {
	cf, err := readAwsConfig()
	if err != nil {
		return "", err
	}

	profileSection := cf.Section(fmt.Sprint("profile ", profile))
	deviceArn := profileSection.Key("mfa").String()
	if len(deviceArn) < 1 {
		return discoverMfaDevice(cfg, cf, profileSection)
	}

	return deviceArn, nil
}

func getProfileSessionToken(cfg aws.Config, mfaArn string, mfaCode string) (*types.Credentials, error) {
	c := sts.NewFromConfig(cfg)
	resp, err := c.GetSessionToken(context.TODO(), &sts.GetSessionTokenInput{
		DurationSeconds: &durationSeconds,
		SerialNumber:    &mfaArn,
		TokenCode:       &mfaCode,
	})

	if err != nil {
		return nil, err
	}

	return resp.Credentials, nil
}

func init() {
	mfaAuthCmd.Flags().StringVarP(&mfaDeviceSerialNum, "mfa-device-arn", "M", "", "MFA Device ARN")
	mfaAuthCmd.Flags().StringVarP(&tokenCode, "mfa-token", "T", "", "MFA Token")
}
