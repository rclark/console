package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/spf13/cobra"
)

type Credentials struct {
	SessionID    string `json:"sessionId"`
	SessionKey   string `json:"sessionKey"`
	SessionToken string `json:"sessionToken"`
}

type Response struct {
	SigninToken string `json:"SigninToken"`
}

var root = &cobra.Command{
	Use:   "console",
	Short: "Log in to the AWS console",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()

		cfg, err := config.LoadDefaultConfig(ctx, config.WithSharedConfigProfile(profile))
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		creds, err := cfg.Credentials.Retrieve(ctx)
		if err != nil {
			return fmt.Errorf("failed to retrieve credentials: %w", err)
		}

		if creds.SessionToken == "" {
			return fmt.Errorf("you must select a profile that uses temporary credentials")
		}

		data, err := json.Marshal(Credentials{
			SessionID:    creds.AccessKeyID,
			SessionKey:   creds.SecretAccessKey,
			SessionToken: creds.SessionToken,
		})
		if err != nil {
			return err
		}

		region := cfg.Region
		duration := int64((12 * time.Hour).Seconds())

		query := url.Values{}
		query.Set("Action", "getSigninToken")
		query.Set("Session", string(data))
		query.Set("SessionDuration", strconv.FormatInt(duration, 10))

		federationURL := fmt.Sprintf("https://%s.signin.aws.amazon.com/federation", region)
		federationURL = federationURL + "?" + query.Encode()

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, federationURL, nil)
		if err != nil {
			return fmt.Errorf("federation request failed: %w", err)
		}

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return fmt.Errorf("federation response failed: %w", err)
		}

		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				return fmt.Errorf("failed to read error response: %v", err)
			}
			return fmt.Errorf("failed to get signin token: [%d] %s", resp.StatusCode, string(body))
		}

		var body Response
		if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
			return fmt.Errorf("failed to decode federation response body: %w", err)
		}

		query = url.Values{}
		query.Set("Action", "login")
		query.Set("SigninToken", body.SigninToken)
		query.Set("Destination", "https://console.aws.amazon.com/")
		query.Set("Issuer", "https://felt.com")

		signinURL := fmt.Sprintf("https://%s.signin.aws.amazon.com/federation", region)
		signinURL = signinURL + "?" + query.Encode()

		if err := open(signinURL); err != nil {
			return fmt.Errorf("failed to open browser: %w", err)
		}

		return nil
	},
}

var profile string

func init() {
	root.PersistentFlags().StringVarP(&profile, "profile", "p", "default", "profile to use")
}

func main() {
	if err := root.Execute(); err != nil {
		os.Exit(1)
	}
}

func open(url string) error {
	var cmd string
	var args []string

	switch runtime.GOOS {
	case "windows":
		cmd = "cmd"
		args = []string{"/c", "start"}
	case "darwin":
		cmd = "open"
	default: // "linux", "freebsd", "openbsd", "netbsd"
		cmd = "xdg-open"
	}
	args = append(args, url)
	return exec.Command(cmd, args...).Start()
}
