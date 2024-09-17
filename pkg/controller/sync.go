package controller

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/Instabug/netbird-gitops/pkg/data"
	"github.com/Instabug/netbird-gitops/pkg/util"
	"github.com/nikoksr/notify"
)

func (c *Controller) doSync(ctx context.Context, cfg *data.CombinedConfig, dryRun bool) error {
	c.netbirdClient.DryRun = dryRun

	groupNameID, groupIDName, err := c.syncGroups(ctx, *cfg)
	if err != nil {
		return fmt.Errorf("Failed syncGroups: %w", err)
	}

	users, err := c.syncUsers(ctx, cfg, groupNameID, groupIDName)
	if err != nil {
		return fmt.Errorf("Failed syncUsers: %w", err)
	}

	peers, err := c.syncPeers(ctx, cfg)
	if err != nil {
		return fmt.Errorf("Failed syncPeers: %w", err)
	}

	err = c.syncPeerGroups(ctx, cfg, users, peers, groupNameID)
	if err != nil {
		return fmt.Errorf("Failed syncPeerGroups: %w", err)
	}

	err = c.syncNetworkRoutes(ctx, cfg, groupNameID)
	if err != nil {
		return fmt.Errorf("Failed syncNetworkRoutes: %w", err)
	}

	pcNameID, err := c.syncPostureChecks(ctx, cfg)
	if err != nil {
		return fmt.Errorf("Failed syncPostureChecks: %w", err)
	}

	err = c.syncPolicies(ctx, cfg, pcNameID, groupNameID)
	if err != nil {
		return fmt.Errorf("Failed syncPolicies: %w", err)
	}

	_, err = c.prunePostureChecks(ctx, cfg)
	if err != nil {
		return fmt.Errorf("Failed prunePostureChecks: %w", err)
	}

	err = c.syncDNSSettings(ctx, cfg, groupNameID)
	if err != nil {
		return fmt.Errorf("Failed syncDNSSettings: %w", err)
	}

	err = c.syncNameservers(ctx, cfg, groupNameID)
	if err != nil {
		return fmt.Errorf("Failed syncNameservers: %w", err)
	}

	err = c.pruneGroups(ctx, groupNameID)
	if err != nil {
		return fmt.Errorf("Failed pruneGroups: %w", err)
	}

	return nil
}

func (c Controller) syncNameservers(ctx context.Context, cfg *data.CombinedConfig, groupNameID map[string]string) error {
	nameservers, err := c.netbirdClient.ListNameservers(ctx)
	if err != nil {
		return err
	}

	nsRevMap := util.SliceToMap(nameservers, func(ns data.Nameserver) string { return ns.Name })
	gitNSRevMap := util.SliceToMap(cfg.Nameservers, func(ns data.Nameserver) string { return ns.Name })

	for k, v := range gitNSRevMap {
		gitNS := data.Nameserver{
			Name:                 v.ID,
			Description:          v.Description,
			Nameservers:          v.Nameservers,
			Enabled:              v.Enabled,
			Groups:               util.Map(v.Groups, func(s string) string { return groupNameID[s] }),
			Primary:              v.Primary,
			Domains:              v.Domains,
			SearchDomainsEnabled: v.SearchDomainsEnabled,
		}
		if nbns, ok := nsRevMap[k]; ok {
			if nbns.Equals(gitNS) {
				slog.Debug("Nameserver matches", "name", v.Name)
				continue
			}
			slog.Warn("Updating Nameserver", "name", v.Name)
			notify.Send(ctx, "", fmt.Sprintf("Updating nameserver %s with config: %+v", v.Name, v))
			gitNS.ID = nbns.ID
			err = c.netbirdClient.UpdateNameserver(ctx, gitNS)
			if err != nil {
				return err
			}
		} else {
			slog.Warn("Creating Nameserver", "name", v.Name)
			notify.Send(ctx, "", fmt.Sprintf("Creating nameserver %s with config: %+v", v.Name, v))
			err = c.netbirdClient.CreateNameserver(ctx, gitNS)
			if err != nil {
				return err
			}
		}
	}

	for k, v := range nsRevMap {
		if _, ok := gitNSRevMap[k]; !ok {
			slog.Warn("Deleting Nameserver", "name", v.Name)
			notify.Send(ctx, "", fmt.Sprintf("Deleting Nameserver %s", v.Name))
			err = c.netbirdClient.DeleteNameserver(ctx, v)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (c Controller) syncDNSSettings(ctx context.Context, cfg *data.CombinedConfig, groupNameID map[string]string) error {
	settings, err := c.netbirdClient.GetDNSSettings(ctx)
	if err != nil {
		return err
	}

	if util.SortedEqual(settings.Items.DisableFor, util.Map(cfg.DNS.DisableFor, func(s string) string { return groupNameID[s] })) {
		slog.Debug("DNS Settings Matches")
		return nil
	}

	slog.Warn("Updating DNS Settings", "disabled_groups", cfg.DNS.DisableFor)
	notify.Send(ctx, "", fmt.Sprintf("Updating DNS Settings with config: %+v", cfg.DNS))
	c.netbirdClient.UpdateDNSSettings(ctx, cfg.DNS)
	return nil
}

func (c Controller) syncPolicies(ctx context.Context, cfg *data.CombinedConfig, pcNameID, groupNameID map[string]string) error {
	policies, err := c.netbirdClient.ListPolicies(ctx)
	if err != nil {
		return err
	}
	policyRevMap := util.SliceToMap(policies, func(p data.Policy) string { return p.Name })
	gitPolicyRevMap := util.SliceToMap(cfg.Policies, func(p data.Policy) string { return p.Name })

	for k, v := range gitPolicyRevMap {
		gitPolicy := data.Policy{
			Name:                v.Name,
			Description:         v.Description,
			Enabled:             v.Enabled,
			SourcePostureChecks: util.Map(v.SourcePostureChecks, func(p string) string { return pcNameID[p] }),
			Action:              v.Action,
			Bidirectional:       v.Bidirectional,
			Protocol:            v.Protocol,
			Ports:               v.Ports,
			Sources:             util.Map(v.Sources, func(p string) string { return groupNameID[p] }),
			Destinations:        util.Map(v.Destinations, func(p string) string { return groupNameID[p] }),
		}
		if nbp, ok := policyRevMap[k]; ok {
			if nbp.Equals(gitPolicy) {
				slog.Debug("Policies matching", "name", gitPolicy.Name)
				continue
			}
			slog.Warn("Updating Policy", "name", gitPolicy.Name)
			notify.Send(ctx, "", fmt.Sprintf("Updating Policy %s with config: %+v", gitPolicy.Name, gitPolicy))
			gitPolicy.ID = nbp.ID
			err = c.netbirdClient.UpdatePolicy(ctx, gitPolicy)
			if err != nil {
				return err
			}
		} else {
			slog.Warn("Creating Policy", "name", gitPolicy.Name)
			notify.Send(ctx, "", fmt.Sprintf("Creating Policy %s with config: %+v", gitPolicy.Name, gitPolicy))
			err = c.netbirdClient.CreatePolicy(ctx, gitPolicy)
			if err != nil {
				return err
			}
		}
	}

	for k, v := range policyRevMap {
		if _, ok := gitPolicyRevMap[k]; !ok {
			slog.Warn("Deleting Policy", "name", v.Name)
			notify.Send(ctx, "", fmt.Sprintf("Deleting Policy %s", v.Name))
			err = c.netbirdClient.DeletePolicy(ctx, v)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (c Controller) syncPostureChecks(ctx context.Context, cfg *data.CombinedConfig) (map[string]string, error) {
	pcNameID := make(map[string]string)
	pcs, err := c.netbirdClient.ListPostureChecks(ctx)
	if err != nil {
		return nil, err
	}

	pcRevMap := util.SliceToMap(pcs, func(p data.PostureCheck) string { return p.Name })
	gitPCRevMap := util.SliceToMap(cfg.PostureChecks, func(p data.PostureCheck) string { return p.Name })

	for k, v := range gitPCRevMap {
		if nbpc, ok := pcRevMap[k]; ok {
			v.ID = nbpc.ID
			slog.Warn("Updating postureCheck", "name", v.Name)
			notify.Send(ctx, "", fmt.Sprintf("Updating postureCheck %s with config: %+v", v.Name, v))
			err = c.netbirdClient.UpdatePostureCheck(ctx, v)
			if err != nil {
				return nil, err
			}
			pcNameID[v.Name] = nbpc.ID
		} else {
			slog.Warn("Creating postureCheck", "name", v.Name)
			notify.Send(ctx, "", fmt.Sprintf("Creating postureCheck %s with config: %+v", v.Name, v))
			pc, err := c.netbirdClient.CreatePostureCheck(ctx, v)
			if err != nil {
				return nil, err
			}
			pcNameID[v.Name] = pc.ID
		}
	}

	return pcNameID, nil
}

func (c Controller) prunePostureChecks(ctx context.Context, cfg *data.CombinedConfig) (map[string]string, error) {
	pcNameID := make(map[string]string)
	pcs, err := c.netbirdClient.ListPostureChecks(ctx)
	if err != nil {
		return nil, err
	}

	pcRevMap := util.SliceToMap(pcs, func(p data.PostureCheck) string { return p.Name })
	gitPCRevMap := util.SliceToMap(cfg.PostureChecks, func(p data.PostureCheck) string { return p.Name })

	for k, v := range pcRevMap {
		if _, ok := gitPCRevMap[k]; !ok {
			slog.Warn("Deleting postureCheck", "name", v.Name)
			notify.Send(ctx, "", fmt.Sprintf("Deleting postureCheck %s", v.Name))
			err = c.netbirdClient.DeletePostureCheck(ctx, v)
			if err != nil {
				return nil, err
			}
		}
	}
	return pcNameID, nil
}

func (c Controller) syncNetworkRoutes(ctx context.Context, cfg *data.CombinedConfig, groupNameID map[string]string) error {
	routes, err := c.netbirdClient.ListNetworkRoutes(ctx)
	if err != nil {
		return err
	}

	routesRevMap := util.SliceToMap(routes, func(r data.NetworkRoute) string { return r.NetworkID })
	gitRoutesRevMap := util.SliceToMap(cfg.NetworkRoutes, func(r data.NetworkRoute) string { return r.NetworkID })

	for k, v := range gitRoutesRevMap {
		gitRoute := data.NetworkRoute{
			NetworkType: v.NetworkType,
			Description: v.Description,
			NetworkID:   v.NetworkID,
			Enabled:     v.Enabled,
			Peer:        v.Peer,
			PeerGroups:  util.Map(v.PeerGroups, func(g string) string { return groupNameID[g] }),
			Network:     v.Network,
			Domains:     v.Domains,
			Metric:      v.Metric,
			Masquerade:  v.Masquerade,
			Groups:      util.Map(v.Groups, func(g string) string { return groupNameID[g] }),
			KeepRoute:   v.KeepRoute,
		}

		if _, ok := routesRevMap[k]; !ok {
			slog.Warn("Creating network route", "route", v)
			notify.Send(ctx, "", fmt.Sprintf("Creating network route %s with config: %+v", v.NetworkID, gitRoute))
			err = c.netbirdClient.CreateNetworkRoute(ctx, gitRoute)
			if err != nil {
				return err
			}
		} else {
			if routesRevMap[k].Equals(gitRoute) {
				slog.Debug("Route matches", "network_id", v.NetworkID)
				continue
			}
			gitRoute.ID = routesRevMap[k].ID
			slog.Warn("Updating network route", "old_route", routesRevMap[k], "route", gitRoute)
			notify.Send(ctx, "", fmt.Sprintf("Updating network route %s with config: %+v", v.NetworkID, gitRoute))
			err = c.netbirdClient.UpdateNetworkRoute(ctx, gitRoute)
			if err != nil {
				return err
			}
		}
	}

	for k, v := range routesRevMap {
		if _, ok := gitRoutesRevMap[k]; !ok {
			slog.Warn("Deleting network route", "network_id", v.NetworkID)
			notify.Send(ctx, "", fmt.Sprintf("Deleting network route %s", v.NetworkID))
			err = c.netbirdClient.DeleteNetworkRoute(ctx, v)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (c Controller) syncPeerGroups(ctx context.Context, cfg *data.CombinedConfig, users map[string]data.User, peers map[string]data.Peer, groupNameID map[string]string) error {
	groups, err := c.netbirdClient.ListGroups(ctx)
	if err != nil {
		return err
	}

	reverseGroupMapping := make(map[string][]string)
	for _, g := range groups {
		for _, p := range g.PeerData {
			reverseGroupMapping[g.ID] = append(reverseGroupMapping[g.ID], p.ID)
		}
	}

	reverseGroupMappingGit := make(map[string][]string)
	if cfg.Config.IndividualPeerGroups {
		for _, p := range cfg.Peers {
			for _, g := range p.GroupNames {
				reverseGroupMappingGit[groupNameID[g]] = append(reverseGroupMappingGit[groupNameID[g]], p.ID)
			}
		}
	} else {
		gitPeerRevMap := util.SliceToMap(cfg.Peers, func(p data.Peer) string { return p.ID })
		for k, p := range peers {
			if p.UserID == "" { // Setup Key, use individual peer Group as per usual
				slog.Debug("Using cfg.Peers", "peer", p.ID, "groups", gitPeerRevMap[k].GroupNames)
				for _, g := range gitPeerRevMap[k].GroupNames {
					reverseGroupMappingGit[groupNameID[g]] = append(reverseGroupMappingGit[groupNameID[g]], p.ID)
				}
			} else {
				slog.Debug("Using user autogroups", "peer", p.ID, "groups", users[p.UserID].Groups)
				for _, g := range users[p.UserID].Groups {
					reverseGroupMappingGit[g] = append(reverseGroupMappingGit[g], p.ID)
				}
			}
		}
	}

	for _, g := range groups {
		if g.Name == "All" {
			continue
		}
		toAdd, toRemove := util.Diff(reverseGroupMappingGit[g.ID], reverseGroupMapping[g.ID])
		toAdd = util.Select(toAdd, func(s string) bool { _, ok := peers[s]; return ok })
		toRemove = util.Select(toRemove, func(s string) bool { _, ok := peers[s]; return ok })
		if len(toAdd)+len(toRemove) == 0 {
			slog.Debug("Group peers matches", "name", g.Name)
			continue
		}
		if len(toAdd) > 0 {
			slog.Warn("Adding peers to group", "group_name", g.Name, "peers", toAdd)
			notify.Send(ctx, "", fmt.Sprintf("Adding peers %+v to group %s", toAdd, g.Name))
		}
		if len(toRemove) > 0 {
			slog.Warn("Removing peers from group", "group_name", g.Name, "peers", toRemove)
			notify.Send(ctx, "", fmt.Sprintf("Removing peers %+v from group %s", toRemove, g.Name))
		}

		g.Peers = reverseGroupMappingGit[g.ID]

		err = c.netbirdClient.UpdateGroup(ctx, g)
		if err != nil {
			return err
		}
	}

	return nil
}

func (c Controller) syncPeers(ctx context.Context, cfg *data.CombinedConfig) (map[string]data.Peer, error) {
	peers, err := c.netbirdClient.ListPeers(ctx)
	if err != nil {
		return nil, err
	}

	gitPeerRevMap := util.SliceToMap(cfg.Peers, func(v data.Peer) string { return v.ID })

	for _, p := range peers {
		gitPeer := gitPeerRevMap[p.ID]
		if _, ok := gitPeerRevMap[p.ID]; !ok {
			// TODO: Delete?
			if p.LoginExpirationEnabled && !p.SSHEnabled {
				continue
			}
			slog.Warn("Peer exists in NetBird but not in Git, disabling SSH and enabling login expiration", "id", p.ID)
			notify.Send(ctx, "", fmt.Sprintf("Peer %s doesn't exist in source control: disabling SSH and enabling login expiration", p.ID))
			p.LoginExpirationEnabled = true
			p.SSHEnabled = false
		} else if p.LoginExpirationEnabled == !gitPeer.ExpirationDisabled && p.Name == gitPeer.Name && p.SSHEnabled == gitPeer.SSHEnabled {
			slog.Debug("Peer matches git", "id", p.ID, "name", p.Name)
			continue
		} else {
			slog.Warn("Updating Peer", "id", p.ID,
				"old_name", p.Name, "new_name", gitPeer.Name,
				"old_expiration_enabled", p.LoginExpirationEnabled, "new_expiration_enabled", !gitPeer.ExpirationDisabled,
				"old_ssh_enabled", p.SSHEnabled, "new_ssh_enabled", gitPeer.SSHEnabled)
			notify.Send(ctx, "", fmt.Sprintf("Updating peer %s with config: %+v", p.ID, p))
			p.LoginExpirationEnabled = !gitPeer.ExpirationDisabled
			p.SSHEnabled = gitPeer.SSHEnabled
			p.Name = gitPeer.Name
			p.GroupNames = gitPeer.GroupNames
		}

		err = c.netbirdClient.UpdatePeer(ctx, p)
		if err != nil {
			return nil, err
		}
	}

	peerRevMap := util.SliceToMap(peers, func(p data.Peer) string { return p.ID })
	for k := range gitPeerRevMap {
		if _, ok := peerRevMap[k]; !ok {
			slog.Warn("Peer exists in Git but not NetBird, Deleted from upstream?", "id", k)
			notify.Send(ctx, "", fmt.Sprintf("Peer %s exists in Git but not in NetBird, deleted from upstream?", k))
		}
	}

	return peerRevMap, nil
}

func (c Controller) syncUsers(ctx context.Context, cfg *data.CombinedConfig, groupNameID, groupIDName map[string]string) (map[string]data.User, error) {
	users, err := c.netbirdClient.ListUsers(ctx)
	if err != nil {
		return nil, err
	}

	userRevMap := make(map[string]data.User)

	emailMappingGit := util.SliceToMap(cfg.Users, func(v data.User) string { return v.Email })
	for _, u := range users {
		userRevMap[u.ID] = u
		if u.ServiceUser {
			// TODO: Manage service users
			continue
		}
		if u.Email == "" {
			slog.Warn("User exists in NetBird with no email", "id", u.ID)
			notify.Send(ctx, "", fmt.Sprintf("User ID %s exists in NetBird with no email, most likely deleted from SSO", u.ID))
			// TODO: Handle deleted user
		}
		gitUser := emailMappingGit[u.Email]
		nbUserGroupNames := util.Map(u.Groups, func(a string) string { return groupIDName[a] })
		if gitUser.Email == "" {
			// User exists in NetBird but not git
			// TODO: Deletion or just blocking?
			slog.Warn("User exists in NetBird but not in Git", "email", u.Email)
			notify.Send(ctx, "", fmt.Sprintf("User %s exists in NetBird but not in Git, user blocked", u.Email))
			u.Blocked = true
			u.Groups = []string{}
			u.Role = ""
		} else if util.SortedEqual(gitUser.Groups, nbUserGroupNames) && u.Role == gitUser.GetRole() {
			slog.Debug("User matches in Netbird and Git", "email", u.Email)
			// User autogroups and role equal
			continue
		}

		// User autogroups not equal
		// Map group names to IDs
		slog.Warn("Updating user", "email", u.Email, "old_groups", nbUserGroupNames, "new_groups", gitUser.Groups, "old_role", u.Role, "new_role", gitUser.GetRole())
		u.Groups = util.Map(gitUser.Groups, func(a string) string { return groupNameID[a] })
		u.Role = gitUser.GetRole()
		notify.Send(ctx, "", fmt.Sprintf("Updating user %s with config: %+v", u.Email, u))

		err := c.netbirdClient.UpdateUser(ctx, u)
		if err != nil {
			return nil, err
		}
	}

	return userRevMap, nil
}

func (c Controller) syncGroups(ctx context.Context, cfg data.CombinedConfig) (groupNameToID map[string]string, groupIDToName map[string]string, err error) {
	// Get all groups from all configuration
	groupNameToID = make(map[string]string)
	for _, g := range cfg.DNS.DisableFor {
		groupNameToID[g] = ""
	}
	for _, route := range cfg.NetworkRoutes {
		for _, g := range route.Groups {
			groupNameToID[g] = ""
		}
		for _, g := range route.PeerGroups {
			groupNameToID[g] = ""
		}
	}
	for _, peer := range cfg.Peers {
		for _, g := range peer.Groups {
			groupNameToID[g.Name] = ""
		}
	}
	for _, policy := range cfg.Policies {
		for _, g := range policy.Sources {
			groupNameToID[g] = ""
		}
		for _, g := range policy.Destinations {
			groupNameToID[g] = ""
		}
	}
	for _, user := range cfg.Users {
		for _, g := range user.Groups {
			groupNameToID[g] = ""
		}
	}

	// Get NetBird groups
	groups, err := c.netbirdClient.ListGroups(ctx)
	if err != nil {
		return nil, nil, err
	}

	groupIDToName = make(map[string]string)

	for _, g := range groups {
		groupIDToName[g.ID] = g.Name
		if _, ok := groupNameToID[g.Name]; ok {
			groupNameToID[g.Name] = g.ID
		}
	}

	for k, v := range groupNameToID {
		if v != "" {
			continue
		}

		slog.Warn("Creating group", "name", k)
		notify.Send(ctx, "", fmt.Sprintf("Creating group %s", k))
		g, err := c.netbirdClient.CreateGroup(ctx, data.Group{
			Name: k,
		})

		if err != nil {
			return nil, nil, err
		}
		slog.Info("Created group", "name", k, "id", g.ID)

		groupNameToID[g.Name] = g.ID
		groupIDToName[g.ID] = g.Name
	}

	return
}

func (c Controller) pruneGroups(ctx context.Context, groupNameID map[string]string) error {
	groups, err := c.netbirdClient.ListGroups(ctx)
	if err != nil {
		return err
	}

	for _, g := range groups {
		if _, ok := groupNameID[g.Name]; !ok {
			if g.Name == "All" {
				continue
			}
			slog.Warn("Deleting Group", "name", g.Name)
			notify.Send(ctx, "", fmt.Sprintf("Deleting group %s as it's not used by any configuration", g.Name))
			c.netbirdClient.DeleteGroup(ctx, g)
		}
	}

	return nil
}
