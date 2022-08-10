package bosh

import (
	"context"
	"fmt"
	"github.com/sirupsen/logrus"
	"net/http"
	"time"

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

func (c *Client) Start(ctx context.Context, i Instance) error {
	log.FromContext(ctx).Debugf("Starting instance %s/%s (%s)", i.Job, i.ID, i.VMName)

	client, err := c.getOrCreateUnderlyingClient()
	if err != nil {
		return err
	}

	task, err := client.Start(i.Deployment, i.Job, i.VMName)
	if err != nil {
		return fmt.Errorf("error trying to start instance %s/%s in deployment %s: %w",
			i.Job, i.ID, i.Deployment, err)
	}
	_, err = client.WaitUntilDone(task, time.Minute*20)
	if err != nil {
		return fmt.Errorf("error waiting for instance %s/%s in deployment %s to start: %w",
			i.Job, i.ID, i.Deployment, err)
	}
	return nil
}

func (c *Client) Stop(ctx context.Context, i Instance) error {
	log.FromContext(ctx).Debugf("Stopping instance %s/%s (%s)", i.Job, i.ID, i.VMName)

	client, err := c.getOrCreateUnderlyingClient()
	if err != nil {
		return err
	}

	task, err := client.Stop(i.Deployment, i.Job, i.VMName)
	if err != nil {
		return fmt.Errorf("error trying to stop instance %s/%s in deployment %s: %w",
			i.Job, i.ID, i.Deployment, err)
	}
	_, err = client.WaitUntilDone(task, time.Minute*20)
	if err != nil {
		return fmt.Errorf("error waiting for instance %s/%s in deployment %s to stop: %w",
			i.Job, i.ID, i.Deployment, err)
	}
	return nil
}

func (c *Client) Instances(ctx context.Context) ([]Instance, error) {
	l := log.FromContext(ctx)

	client, err := c.getOrCreateUnderlyingClient()
	if err != nil {
		return []Instance{}, err
	}

	return c.instances(client, l)
}

func (c *Client) VMsAndStemcells(ctx context.Context) ([]string, error) {
	l := log.FromContext(ctx)

	client, err := c.getOrCreateUnderlyingClient()
	if err != nil {
		return []string{}, err
	}

	results, err := c.stemcells(client, l)
	if err != nil {
		return []string{}, err
	}
	vmResults, err := c.instances(client, l)
	if err != nil {
		return []string{}, err
	}

	for _, vm := range vmResults {
		results = append(results, vm.VMName)
	}

	return results, nil
}

func (c *Client) stemcells(client *gogobosh.Client, l *logrus.Entry) ([]string, error) {
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

	return result, nil
}

func (c *Client) instances(client *gogobosh.Client, l *logrus.Entry) ([]Instance, error) {
	l.Debug("Getting all BOSH managed VMs")
	deployments, err := client.GetDeployments()
	if err != nil {
		return []Instance{}, fmt.Errorf("failed to get bosh deployments: %w", err)
	}

	var result []Instance
	for _, d := range deployments {
		l.Infof("Found deployment %s", d.Name)
		vms, err := client.GetDeploymentVMs(d.Name)
		if err != nil {
			return []Instance{}, fmt.Errorf("failed to get deployment %s VMs: %w", d.Name, err)
		}
		l.Infof("With %d BOSH managed VMs", len(vms))

		for _, vm := range vms {
			instanceName := vm.JobName + "/" + vm.ID
			l.Debugf("  %s - %s", vm.VMCID, instanceName)
			result = append(result, Instance{
				ID:         vm.ID,
				VMName:     vm.VMCID,
				Name:       instanceName,
				Deployment: d.Name,
				Job:        vm.JobName,
			})
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
