/*
 * Copyright 2023 VMware, Inc.
 * SPDX-License-Identifier: Apache-2.0
 */

package vcenter

import "context"

// Pool is a pool of vcenter client instances
type Pool struct {
	sourceClientsByAZ map[string]*Client
	targetClientsByAZ map[string]*Client
	clients           []*Client
}

// NewPool creates a new empty pool
// Call AddSource and AddTarget to fully initialize the pool
func NewPool() *Pool {
	return &Pool{
		sourceClientsByAZ: make(map[string]*Client, 1),
		targetClientsByAZ: make(map[string]*Client, 1),
	}
}

// NewPoolWithExternalClients creates a new vcenter pool using externally managed vcenter clients
// This is used for testing purposes
func NewPoolWithExternalClients(sourceClientsByAZ, targetClientsByAZ map[string]*Client) *Pool {
	p := NewPool()
	p.sourceClientsByAZ = sourceClientsByAZ
	p.targetClientsByAZ = targetClientsByAZ
	return p
}

// AddSource adds a new source az/client pair
// If the AZ's vcenter matches another AZ's source vcenter then the client is re-used
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

// AddTarget adds a new target az/client pair
// If the AZ's vcenter matches another AZ's target vcenter then the client is re-used
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

// GetSourceClientByAZ returns the source vcenter client associated with the AZ, otherwise nil
func (p *Pool) GetSourceClientByAZ(az string) *Client {
	return p.sourceClientsByAZ[az]
}

// GetTargetClientByAZ returns the target vcenter client associated with the AZ, otherwise nil
func (p *Pool) GetTargetClientByAZ(az string) *Client {
	return p.targetClientsByAZ[az]
}

// GetSourceClients returns all source vcenter clients
func (p *Pool) GetSourceClients() []*Client {
	var clients []*Client
	for _, c := range p.sourceClientsByAZ {
		clients = append(clients, c)
	}
	return clients
}

// GetTargetClients returns all target vcenter clients
func (p *Pool) GetTargetClients() []*Client {
	var clients []*Client
	for _, c := range p.targetClientsByAZ {
		clients = append(clients, c)
	}
	return clients
}

// GetClients returns all target and source vcenter clients
func (p *Pool) GetClients() []*Client {
	return p.clients
}

// SourceAZs returns all source AZs
// This should always match TargetAZs except during initialization
func (p *Pool) SourceAZs() []string {
	var azs []string
	for az := range p.sourceClientsByAZ {
		azs = append(azs, az)
	}
	return azs
}

// TargetAZs returns all target AZs
// This should always match SourceAZs except during initialization
func (p *Pool) TargetAZs() []string {
	var azs []string
	for az := range p.targetClientsByAZ {
		azs = append(azs, az)
	}
	return azs
}

// Close calls logout on all managed vcenter clients
func (p *Pool) Close(ctx context.Context) {
	for _, c := range p.clients {
		c.Logout(ctx)
	}
}

func (p *Pool) getClient(host, username, password string, insecure bool) *Client {
	for _, c := range p.clients {
		if c.isSameVCenter(host, username, password, insecure) {
			return c
		}
	}
	return nil
}
