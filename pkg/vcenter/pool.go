/*
 * Copyright 2023 VMware, Inc.
 * SPDX-License-Identifier: Apache-2.0
 */

package vcenter

import "context"

type Pool struct {
	sourceClientsByAZ map[string]*Client
	targetClientsByAZ map[string]*Client
	clients           []*Client
}

func NewPool() *Pool {
	return &Pool{
		sourceClientsByAZ: make(map[string]*Client, 1),
		targetClientsByAZ: make(map[string]*Client, 1),
	}
}

func NewPoolWithExternalClients(sourceClientsByAZ, targetClientsByAZ map[string]*Client) *Pool {
	p := NewPool()
	p.sourceClientsByAZ = sourceClientsByAZ
	p.targetClientsByAZ = targetClientsByAZ
	return p
}

func (p *Pool) AddSource(az, host, username, password, datacenter string, insecure bool) {
	if p.GetSourceClientByAZ(az) == nil {
		// reuse clients between AZs
		c := p.getClient(host, username, password, insecure)
		if c == nil {
			c = New(host, username, password, datacenter, insecure)
			p.clients = append(p.clients, c)
		}
		p.sourceClientsByAZ[az] = c
	}
}

func (p *Pool) AddTarget(az, host, username, password, datacenter string, insecure bool) {
	if p.GetTargetClientByAZ(az) == nil {
		// reuse clients between AZs
		c := p.getClient(host, username, password, insecure)
		if c == nil {
			c = New(host, username, password, datacenter, insecure)
			p.clients = append(p.clients, c)
		}
		p.targetClientsByAZ[az] = c
	}
}

func (p *Pool) GetSourceClientByAZ(az string) *Client {
	c, _ := p.sourceClientsByAZ[az]
	return c
}

func (p *Pool) GetTargetClientByAZ(az string) *Client {
	c, _ := p.targetClientsByAZ[az]
	return c
}

func (p *Pool) GetSourceClients() []*Client {
	var clients []*Client
	for _, c := range p.sourceClientsByAZ {
		clients = append(clients, c)
	}
	return clients
}

func (p *Pool) GetTargetClients() []*Client {
	var clients []*Client
	for _, c := range p.targetClientsByAZ {
		clients = append(clients, c)
	}
	return clients
}

func (p *Pool) GetClients() []*Client {
	return p.clients
}

func (p *Pool) SourceAZs() []string {
	var azs []string
	for az := range p.sourceClientsByAZ {
		azs = append(azs, az)
	}
	return azs
}

func (p *Pool) TargetAZs() []string {
	var azs []string
	for az := range p.targetClientsByAZ {
		azs = append(azs, az)
	}
	return azs
}

func (p *Pool) Close(ctx context.Context) {
	for _, c := range p.clients {
		c.Logout(ctx)
	}
}

func (p *Pool) getClient(host, username, password string, insecure bool) *Client {
	for _, c := range p.clients {
		if isSameVCenter(c, host, username, password, insecure) {
			return c
		}
	}
	return nil
}

func isSameVCenter(c *Client, host, username, password string, insecure bool) bool {
	return c.Host == host && c.Username == username && c.Password == password && c.Insecure == insecure
}
