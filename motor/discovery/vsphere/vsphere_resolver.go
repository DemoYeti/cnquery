package vsphere

import (
	"errors"
	"fmt"
	"strings"

	"go.mondoo.io/mondoo/motor/motorid"
	"go.mondoo.io/mondoo/motor/transports/resolver"

	"go.mondoo.io/mondoo/motor/discovery/common"

	"github.com/rs/zerolog/log"
	"go.mondoo.io/mondoo/motor/asset"
	"go.mondoo.io/mondoo/motor/platform"
	"go.mondoo.io/mondoo/motor/transports"
	"go.mondoo.io/mondoo/motor/transports/vsphere"
)

const (
	DiscoveryAll          = "all"
	DiscoveryInstances    = "instances"
	DiscoveryHostMachines = "host-machines"
)

type Resolver struct{}

func (r *Resolver) Name() string {
	return "VMware vSphere Resolver"
}

func (r *Resolver) AvailableDiscoveryTargets() []string {
	return []string{DiscoveryAll, DiscoveryInstances, DiscoveryHostMachines}
}

func (r *Resolver) Resolve(tc *transports.TransportConfig, cfn common.CredentialFn, sfn common.QuerySecretFn, userIdDetectors ...transports.PlatformIdDetector) ([]*asset.Asset, error) {
	resolved := []*asset.Asset{}

	// we leverage the vpshere transport to establish a connection
	m, err := resolver.NewMotorConnection(tc, cfn)
	if err != nil {
		return nil, err
	}
	defer m.Close()

	trans, ok := m.Transport.(*vsphere.Transport)
	if !ok {
		return nil, errors.New("could not initialize vsphere transport")
	}

	// detect platform info for the asset
	pf, err := m.Platform()
	if err != nil {
		return nil, err
	}

	// add asset for the api itself
	info := trans.Info()
	assetInfo := &asset.Asset{
		Name:        fmt.Sprintf("%s (%s)", tc.Host, info.Name),
		Platform:    pf,
		Connections: []*transports.TransportConfig{tc}, // pass-in the current config
		Labels: map[string]string{
			"vsphere.vmware.com/name": info.Name,
			"vsphere.vmware.com/uuid": info.InstanceUuid,
		},
	}
	platformIds, err := motorid.GatherIDs(m.Transport, pf, nil)
	if err != nil {
		return nil, err
	}
	assetInfo.PlatformIds = platformIds
	log.Debug().Strs("identifier", assetInfo.PlatformIds).Msg("motor connection")

	resolved = append(resolved, assetInfo)

	client := trans.Client()
	discoveryClient := New(client)

	if tc.IncludesDiscoveryTarget(DiscoveryAll) || tc.IncludesDiscoveryTarget(DiscoveryHostMachines) {
		// resolve esxi hosts
		hosts, err := discoveryClient.ListEsxiHosts()
		if err != nil {
			return nil, err
		}

		// add transport config for each host
		for i := range hosts {
			host := hosts[i]
			ht := tc.Clone()
			// pass-through "vsphere.vmware.com/reference-type" and "vsphere.vmware.com/inventorypath"
			ht.Options = host.Annotations
			host.Connections = append(host.Connections, ht)

			pf, err := platform.VspherePlatform(trans, host.PlatformIds[0])
			if err == nil {
				host.Platform = pf
			} else {
				log.Error().Err(err).Msg("could not determine platform information for esxi host")
			}

			resolved = append(resolved, host)
		}
	}

	if tc.IncludesDiscoveryTarget(DiscoveryAll) || tc.IncludesDiscoveryTarget(DiscoveryInstances) {
		// resolve vms
		vms, err := discoveryClient.ListVirtualMachines(tc)
		if err != nil {
			return nil, err
		}

		// add transport config for each vm
		for i := range vms {
			vm := vms[i]

			pf, err := platform.VspherePlatform(trans, vm.PlatformIds[0])
			if err == nil {
				vm.Platform = pf
			} else {
				log.Error().Err(err).Msg("could not determine platform information for esxi vm")
			}

			// find the secret reference for the asset
			EnrichVsphereToolsConnWithSecrets(vm, cfn, sfn)

			resolved = append(resolved, vm)
		}
	}

	// filter assets
	discoverFilter := map[string]string{}
	if tc.Discover != nil {
		discoverFilter = tc.Discover.Filter
	}

	if namesFilter, ok := discoverFilter["names"]; ok {
		names := strings.Split(namesFilter, ",")
		resolved = filter(resolved, func(a *asset.Asset) bool {
			return contains(names, a.Name)
		})
	}

	if moidsFilter, ok := discoverFilter["moids"]; ok {
		moids := strings.Split(moidsFilter, ",")
		resolved = filter(resolved, func(a *asset.Asset) bool {
			label, ok := a.Labels["vsphere.vmware.com/moid"]
			log.Debug().Strs("moids", moids).Str("search", label).Msg("check if moid is included")
			if !ok {
				return false
			}
			return contains(moids, label)
		})
	}

	return resolved, nil
}

func filter(a []*asset.Asset, keep func(asset *asset.Asset) bool) []*asset.Asset {
	n := 0
	for _, x := range a {
		if keep(x) {
			a[n] = x
			n++
		}
	}
	a = a[:n]
	return a
}

func contains(slice []string, entry string) bool {
	sanitizedEntry := strings.ToLower(strings.TrimSpace(entry))

	for i := range slice {
		if strings.ToLower(strings.TrimSpace(slice[i])) == sanitizedEntry {
			return true
		}
	}
	return false
}
