package bosh

import (
	"context"
	"fmt"
	"net/http"

	"github.com/cloudfoundry-community/gogobosh"
	"github.com/vmware-tanzu/vmotion-migration-tool-for-bosh-deployments/pkg/log"
)

type Client struct {
	ClientID     string
	ClientSecret string
	Environment  string
	client       *gogobosh.Client
}

func New(environment, clientID, clientSecret string) *Client {
	return &Client{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Environment:  environment,
	}
}

func (c *Client) VMsAndStemcells(ctx context.Context) ([]string, error) {
	l := log.FromContext(ctx)

	client, err := c.getOrCreateUnderlyingClient()
	if err != nil {
		return []string{}, err
	}

	l.Debug("Getting all BOSH managed stemcells")
	stemcells, err := client.GetStemcells()
	if err != nil {
		return []string{}, fmt.Errorf(
			"failed to get bosh stemcells, this can happen because of incorrect login details: %w", err)
	}

	var result []string
	for _, s := range stemcells {
		l.Debugf("  %s - %s", s.Name, s.CID)
		result = append(result, s.CID)
	}

	deployments, err := client.GetDeployments()
	if err != nil {
		return []string{}, fmt.Errorf("failed to get bosh deployments: %w", err)
	}

	for _, d := range deployments {
		l.Infof("Found deployment %s", d.Name)
		vms, err := client.GetDeploymentVMs(d.Name)
		if err != nil {
			return []string{}, fmt.Errorf("failed to get deployment %s VMs: %w", d.Name, err)
		}
		l.Infof("With %d BOSH managed VMs", len(vms))

		for _, vm := range vms {
			instanceName := vm.JobName + "/" + vm.ID
			l.Debugf("  %s - %s", vm.VMCID, instanceName)
			result = append(result, vm.VMCID)
		}
	}

	return result, nil
}

func (c *Client) getOrCreateUnderlyingClient() (*gogobosh.Client, error) {
	if c.client != nil {
		return c.client, nil
	}

	config := &gogobosh.Config{
		BOSHAddress:       fmt.Sprintf("https://%s:25555", c.Environment),
		ClientID:          c.ClientID,
		ClientSecret:      c.ClientSecret,
		HttpClient:        http.DefaultClient,
		SkipSslValidation: true,
	}

	log.WithoutContext().Debugf("Creating bosh client to connect to %s", config.BOSHAddress)
	client, err := gogobosh.NewClient(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create bosh client: %w", err)
	}

	c.client = client
	return client, nil
}
