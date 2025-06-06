/*
 * Teleport
 * Copyright (C) 2023  Gravitational, Inc.
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Affero General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Affero General Public License for more details.
 *
 * You should have received a copy of the GNU Affero General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.
 */

package azure

import (
	"context"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/arm"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/runtime"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute/v6"
	"github.com/gravitational/trace"

	"github.com/gravitational/teleport/api/types"
)

// virtualScaleSetUniformVMResourceType represents the resource type of uniform
// virtual scale set VMs.
const virtualScaleSetUniformVMResourceType = "virtualMachineScaleSets/virtualMachines"

// armCompute provides an interface for an Azure virtual machine client.
type armCompute interface {
	// Get retrieves information about an Azure virtual machine.
	Get(ctx context.Context, resourceGroupName string, vmName string, options *armcompute.VirtualMachinesClientGetOptions) (armcompute.VirtualMachinesClientGetResponse, error)
	// NewListPagers lists Azure virtual Machines.
	NewListPager(resourceGroup string, opts *armcompute.VirtualMachinesClientListOptions) *runtime.Pager[armcompute.VirtualMachinesClientListResponse]
	// NewListAllPager lists Azure virtual machines in any resource group.
	NewListAllPager(opts *armcompute.VirtualMachinesClientListAllOptions) *runtime.Pager[armcompute.VirtualMachinesClientListAllResponse]
}

// scaleSet provides an interfaces for an Azure VM scale set client.
type scaleSet interface {
	// Get retrieves a virtual machine from a VM scale set.
	Get(ctx context.Context, resourceGroupName string, vmScaleSetName string, instanceID string, options *armcompute.VirtualMachineScaleSetVMsClientGetOptions) (armcompute.VirtualMachineScaleSetVMsClientGetResponse, error)
}

// VirtualMachinesClient is a client for Azure virtual machines.
type VirtualMachinesClient interface {
	// Get returns the virtual machine (including scale set VMs) for the given
	// resource ID.
	Get(ctx context.Context, resourceID string) (*VirtualMachine, error)
	// GetByVMID returns the virtual machine for a given VM ID.
	GetByVMID(ctx context.Context, vmID string) (*VirtualMachine, error)
	// ListVirtualMachines gets all of the virtual machines in the given resource group.
	ListVirtualMachines(ctx context.Context, resourceGroup string) ([]*armcompute.VirtualMachine, error)
}

// VirtualMachine represents an Azure virtual machine.
type VirtualMachine struct {
	// ID resource ID.
	ID string `json:"id,omitempty"`
	// Name resource name.
	Name string `json:"name,omitempty"`
	// Subscription is the Azure subscription the VM is in.
	Subscription string
	// ResourceGroup is the resource group the VM is in.
	ResourceGroup string
	// VMID is the VM's ID.
	VMID string
	// Identities are the identities associated with the resource.
	Identities []Identity
}

// Identitiy represents an Azure virtual machine identity.
type Identity struct {
	// ResourceID the identity resource ID.
	ResourceID string
}

type vmClient struct {
	// api is the Azure virtual machine client.
	api armCompute
	// scaleSetAPI is the Azure VM scale set client.
	scaleSetAPI scaleSet
}

// NewVirtualMachinesClient creates a new Azure virtual machines client by
// subscription and credentials.
func NewVirtualMachinesClient(subscription string, cred azcore.TokenCredential, options *arm.ClientOptions) (VirtualMachinesClient, error) {
	computeAPI, err := armcompute.NewVirtualMachinesClient(subscription, cred, options)
	if err != nil {
		return nil, trace.Wrap(err)
	}
	scaleSetAPI, err := armcompute.NewVirtualMachineScaleSetVMsClient(subscription, cred, options)
	if err != nil {
		return nil, trace.Wrap(err)
	}

	return NewVirtualMachinesClientByAPI(computeAPI, scaleSetAPI), nil
}

// NewVirtualMachinesClientByAPI creates a new Azure virtual machines client by
// ARM API client.
func NewVirtualMachinesClientByAPI(api armCompute, scaleSetAPI scaleSet) VirtualMachinesClient {
	return &vmClient{
		api:         api,
		scaleSetAPI: scaleSetAPI,
	}
}

type vmTypes interface {
	*armcompute.VirtualMachine | *armcompute.VirtualMachineScaleSetVM
}

func parseVirtualMachine[T vmTypes](vm T) (*VirtualMachine, error) {
	var (
		id       string
		name     string
		identity *armcompute.VirtualMachineIdentity
		vmID     *string
	)

	switch v := any(vm).(type) {
	case *armcompute.VirtualMachine:
		id = *v.ID
		name = *v.Name
		identity = v.Identity
		if v.Properties != nil {
			vmID = v.Properties.VMID
		}

	case *armcompute.VirtualMachineScaleSetVM:
		id = *v.ID
		name = *v.Name
		identity = v.Identity
		if v.Properties != nil {
			vmID = v.Properties.VMID
		}
	}

	resourceID, err := arm.ParseResourceID(id)
	if err != nil {
		return nil, trace.Wrap(err)
	}

	var identities []Identity
	if identity != nil {
		if systemAssigned := StringVal(identity.PrincipalID); systemAssigned != "" {
			identities = append(identities, Identity{ResourceID: systemAssigned})
		}

		for identityID := range identity.UserAssignedIdentities {
			identities = append(identities, Identity{ResourceID: identityID})
		}
	}

	return &VirtualMachine{
		ID:            id,
		Name:          name,
		Subscription:  resourceID.SubscriptionID,
		ResourceGroup: resourceID.ResourceGroupName,
		VMID:          StringVal(vmID),
		Identities:    identities,
	}, nil
}

// Get returns the virtual machine (including scale set VMs) for the given
// resource ID.
//
// The virtual machine scale set (VMSS) supports two types of orchestration
// modes: uniform and flexible. Both have different resource ID format from the
// instance metadata API. A VM from a uniform VMSS has a different resource ID
// and requires a different API to retrieve its information. Flexible VMSS VMs
// use the same resource ID format as regular VMs and don't require special
// handling.
func (c *vmClient) Get(ctx context.Context, resourceID string) (*VirtualMachine, error) {
	parsedResourceID, err := arm.ParseResourceID(resourceID)
	if err != nil {
		return nil, trace.Wrap(err)
	}

	if parsedResourceID.ResourceType.Type == virtualScaleSetUniformVMResourceType {
		return c.getScaleSetVM(ctx, parsedResourceID)
	}

	resp, err := c.api.Get(ctx, parsedResourceID.ResourceGroupName, parsedResourceID.Name, nil)
	if err != nil {
		return nil, trace.Wrap(err)
	}

	vm, err := parseVirtualMachine(&resp.VirtualMachine)
	return vm, trace.Wrap(err)
}

// GetByVMID returns the virtual machine for a given VM ID.
func (c *vmClient) GetByVMID(ctx context.Context, vmID string) (*VirtualMachine, error) {
	pager := newListAllPager(c.api.NewListAllPager(&armcompute.VirtualMachinesClientListAllOptions{}))
	for pager.more() {
		res, err := pager.nextPage(ctx)
		if err != nil {
			return nil, trace.Wrap(ConvertResponseError(err))
		}

		for _, vm := range res {
			if vm.Properties != nil && *vm.Properties.VMID == vmID {
				result, err := parseVirtualMachine(vm)
				return result, trace.Wrap(err)
			}
		}
	}
	return nil, trace.NotFound("no VM with ID %q", vmID)
}

func (c *vmClient) getScaleSetVM(ctx context.Context, resourceID *arm.ResourceID) (*VirtualMachine, error) {
	if resourceID.Parent == nil {
		return nil, trace.BadParameter("expected resource ID to include scale set as parent resource")
	}

	resp, err := c.scaleSetAPI.Get(ctx, resourceID.ResourceGroupName, resourceID.Parent.Name, resourceID.Name, nil)
	if err != nil {
		return nil, trace.Wrap(err)
	}

	result, err := parseVirtualMachine(&resp.VirtualMachineScaleSetVM)
	return result, trace.Wrap(err)
}

type vmPager struct {
	more     func() bool
	nextPage func(context.Context) ([]*armcompute.VirtualMachine, error)
}

func newListPager(azurePager *runtime.Pager[armcompute.VirtualMachinesClientListResponse]) vmPager {
	return vmPager{
		more: azurePager.More,
		nextPage: func(ctx context.Context) ([]*armcompute.VirtualMachine, error) {
			res, err := azurePager.NextPage(ctx)
			return res.Value, trace.Wrap(err)
		},
	}
}

func newListAllPager(azurePager *runtime.Pager[armcompute.VirtualMachinesClientListAllResponse]) vmPager {
	return vmPager{
		more: azurePager.More,
		nextPage: func(ctx context.Context) ([]*armcompute.VirtualMachine, error) {
			res, err := azurePager.NextPage(ctx)
			return res.Value, trace.Wrap(err)
		},
	}
}

// ListVirtualMachines lists all virtual machines in a given resource group
// using the Azure virtual machines API. If resourceGroup is "*", it lists
// all virtual machines in any resource group.
func (c *vmClient) ListVirtualMachines(ctx context.Context, resourceGroup string) ([]*armcompute.VirtualMachine, error) {
	var pager vmPager
	if resourceGroup == types.Wildcard {
		pager = newListAllPager(c.api.NewListAllPager(&armcompute.VirtualMachinesClientListAllOptions{}))
	} else {
		pager = newListPager(c.api.NewListPager(resourceGroup, &armcompute.VirtualMachinesClientListOptions{}))
	}
	var virtualMachines []*armcompute.VirtualMachine
	for pager.more() {
		res, err := pager.nextPage(ctx)
		if err != nil {
			return nil, trace.Wrap(ConvertResponseError(err))
		}
		virtualMachines = append(virtualMachines, res...)
	}

	return virtualMachines, nil
}

// RunCommandRequest combines parameters for running a command on an Azure virtual machine.
type RunCommandRequest struct {
	// Region is the region of the VM.
	Region string
	// ResourceGroup is the resource group for the VM.
	ResourceGroup string
	// VMName is the name of the VM.
	VMName string
	// Script is the URI of the script for the virtual machine to execute.
	Script string
	// Parameters is a list of parameters for the script.
	Parameters []string
}

// RunCommandClient is a client for Azure Run Commands.
type RunCommandClient interface {
	Run(ctx context.Context, req RunCommandRequest) error
}

type runCommandClient struct {
	api *armcompute.VirtualMachineRunCommandsClient
}

// NewRunCommandClient creates a new Azure Run Command client by subscription
// and credentials.
func NewRunCommandClient(subscription string, cred azcore.TokenCredential, options *arm.ClientOptions) (RunCommandClient, error) {
	runCommandAPI, err := armcompute.NewVirtualMachineRunCommandsClient(subscription, cred, options)
	if err != nil {
		return nil, trace.Wrap(err)
	}
	return &runCommandClient{
		api: runCommandAPI,
	}, nil
}

// Run runs a command on a virtual machine.
func (c *runCommandClient) Run(ctx context.Context, req RunCommandRequest) error {
	var params []*armcompute.RunCommandInputParameter
	for _, value := range req.Parameters {
		params = append(params, &armcompute.RunCommandInputParameter{
			Value: to.Ptr(value),
		})
	}
	poller, err := c.api.BeginCreateOrUpdate(ctx, req.ResourceGroup, req.VMName, "RunShellScript", armcompute.VirtualMachineRunCommand{
		Location: to.Ptr(req.Region),
		Properties: &armcompute.VirtualMachineRunCommandProperties{
			AsyncExecution: to.Ptr(false),
			Parameters:     params,
			Source: &armcompute.VirtualMachineRunCommandScriptSource{
				Script: to.Ptr(req.Script),
			},
		},
	}, nil)
	if err != nil {
		return trace.Wrap(err)
	}
	_, err = poller.PollUntilDone(ctx, nil /* options */)
	return trace.Wrap(err)
}
