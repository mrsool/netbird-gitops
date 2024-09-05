# NetBird GitOps

[![Go Reference](https://pkg.go.dev/badge/github.com/Instabug/netbird-gitops.svg)](https://pkg.go.dev/github.com/Instabug/netbird-gitops)
![CodeQL](https://github.com/instabug/netbird-gitops/actions/workflows/github-code-scanning/codeql/badge.svg)
![Build](https://github.com/instabug/netbird-gitops/actions/workflows/docker-publish.yml/badge.svg)

This program is made to synchronize [Netbird](https://netbird.io) configuration 
with a source-controller git repository.

## Installation

You can deploy this as a container alongside NetBird management service or as a 
standalone docker container

### Docker Compose

```yaml
services:
  gitops:
    image: instabug/netbird-gitops:latest
    restart: unless-stopped
    commands:
      - --notify-services-path=/notify.yaml
    volumes:
      - ./notify.yaml:/
      # SSH key used in case of SSH auth method
      # - ./key.pem:/key.pem
    environment:
      # Repository Clone URL
      - GIT_AUTH_METHOD=basic # Valid options (none, basic, ssh)
      - GIT_REPO_URL=https://github.com/Instabug/netbird-gitops.git
      # Path within repository for configurations, leave empty for root
      - GIT_RELATIVE_PATH=netbird-configs
      # HTTPS Username (Set to anything in case of access token)
      - GIT_USERNAME=someone
      # HTTPS Password/Access Token
      - GIT_PASSWORD=password
      # Uncomment in case of SSH key
      # - GIT_PRIVATE_KEY_PATH=/key.pem
      # - GIT_PRIVATE_KEY_PASSWORD=somepassword
      - NETBIRD_TOKEN=abcdef
      - NETBIRD_MANAGEMENT_API=https://api.netbird.io
      - LOG_LEVEL=info

```

## Configuration

Configuration files are written in YAML and can be written in 1 or more files 
within the directory specified

### Schema

You can check the [example](./example) for a configuration example.

> Note: All Group, and PostureCheck names are names, not IDs, NetBird GitOps does the translation

#### NetBird GitOps Config

Configuration for NetBird GitOps itself

```yaml
config:
  # autoSync behavior
  # - manual: only sync if --sync-and-exit is set
  # - update: only sync if Git is updated
  # - enforce: always sync
  autoSync: ("manual", "update", "enforce")
  # Set peer groups individually
  # When set to false, peers that belong to users are given the user's autogroups
  individualPeerGroups: false
```

#### DNS Settings

Configuration for NetBird DNS 

```yaml
dns:
  disable_for:
  - group1
  - group2

nameservers:
- name: Google DNS
  description: Google DNS servers
  nameservers:
    - ip: 8.8.8.8
      ns_type: udp
      port: 53
  enabled: true
  groups:
    - group1
    - group2
  primary: true
  domains:
    - example.com
  search_domains_enabled: true
```

#### Network Routes

```yaml
network_routes:
- network_type: ("IPv4"|"IPv6"|"Domain")
  description: Route Description # Optional
  network_id: Route 1 # Required
  enabled: true # Optional, defaults to false
  # peer_groups and peer are mutually exclusive
  peer_groups: # Optional, must be set if peer is not set
    - g2
  peer: c2312515613213 # Optional, must be set if peer_groups not set
  # domains and network are mutually exclusive
  domains:
    - example.com
  network: 0.0.0.0/0 
  metric: 9999 # Required
  masquerade: true # Optional, defaults to false
  groups: # Required
    - g1
  keep_route: true # Optional, deafults to false
```

#### Peers

Since peers cannot be added from API, this is used to manage Peer Groups and settings

```yaml
peers:
- id: cr6ibk8pcsa9d3fncct0 # Required
  name: "Test" # Required
  groups: # Optional, All is implicitly included
  - g2
  ssh_enabled: true # Optional, defaults to false
  expiration_disabled: true # Optional, defaults to false
```

#### Policies

```yaml
policies:
- name: Production # Required
  description: Production machines access # Required
  enabled: false # Optional, defaults to false
  source_posture_checks: # Optional
  - pc1
  action: accept # Required
  bidirectional: false # Optional, defaults to false
  protocol: all # Required (all|tcp|udp|icmp)
  sources: # Required
  - g1
  destinations: # Required
  - g3
```

#### Posture Checks

```yaml
posture_checks:
- name: pc1 # Required
  description: Something # Required
  checks:
    nb_version_check: # Optional
      min_version: "14.3" # Required
    os_version_check: # Optional
      android: # Optional
        min_version: "13" # Required
      ios: # Optional
        min_version: 17.3.1 # Required
      darwin: # Optional
        min_version: 14.2.1 # Required
      linux: # Optional
        min_kernel_version: 5.3.3 # Required
      windows: # Optional
        min_kernel_version: 10.0.1234 # Required
    geo_location_check: # Optional
      locations: # Required
        - country_code: DE # Required
          city_name: Berlin # Optional
      action: allow # Required (allow|block)
    peer_network_range_check: # Optional
      ranges: # Required
          - 192.168.1.0/24
          - 10.0.0.0/8
          - 2001:db8:1234:1a00::/56
      action: allow # Required (allow|block)
    process_check: # Optional
      processes: # Required
        - linux_path: /usr/local/bin/netbird # Optional
          mac_path: /Applications/NetBird.app/Contents/MacOS/netbird # Optional
          windows_path: "C:\ProgramData\\NetBird\\netbird.exe" # Optional
```

#### Users

```yaml
users:
- email: someone@somewhere.com # Required
  groups: # Required
  - g1
  - g2
  role: admin # Optional, defaults to user (user|admin|owner)
```

### Notification Services

This projects supports sending notifications to any services supported by [nikoksr/notify](https://github.com/nikoksr/notify), however only Slack is implemented currently.

#### Configuration schema

Configuring notification services exists in `notify.yaml` by default and can be overridden with `--notify-services-path`

```yaml
slack:
  token: xoxb-....
  channels:
  - channel-a
  - channel-b
```

## Usage

netbird-gitops can run in enforce mode where only Git configuration is the source of truth, it also supports manual syncing through the `--sync-and-exit` flag, which will pull the configuration, apply them and exit.

```bash
  -git-auth-method string
    	basic (username-password/access token), or ssh (private key), or none (default "none")
  -git-branch string
    	Name of branch to pull changes from (default "main")
  -git-password string
    	git basic auth password, must be defined if --git-auth-method is basic
  -git-private-key-password string
    	git SSH private key password (if any)
  -git-private-key-path string
    	git SSH private key path, must be defined if --git-auth-method is ssh
  -git-relative-path string
    	Relative path of NetBird configuration within the git repo
  -git-repo-url string
    	Git Repo URL (ssh/https) (Required)
  -git-username string
    	git basic auth username, must be defined if --git-auth-method is basic
  -log-level string
    	Log level (debug, info, warn, error)
  -netbird-mgmt-api string
    	NetBird Management API URL
  -netbird-token string
    	NetBird Management API token (default "nbp_woIGracLxicjqDafocrFpKPZYO4KCN3HOcE5")
  -notify-services-path string
    	Path to notification services configuration yaml (default "notify.yaml")
  -sync-and-exit
    	Force sync once and exit
```

## Legal

NetBird is a [registered trademark](https://netbird.io/terms) of [Wiretrustee UG (haftungsbeschränkt)](https://netbird.io/) & [AUTHORS](https://github.com/netbirdio/netbird/blob/main/AUTHORS)