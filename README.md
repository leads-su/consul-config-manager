# CCM - Consul Config Manager

CCM is a tool, which can help you automate configuration delivery for your applications.  
It uses Consul as a backend service, to store configuration, and watches for any updates which need to be delivered.

# Features

## Service Registration
Application automatically registeres itself as a service in Consul, so you can monitor its health through Consul UI.  
![Application Service](https://raw.githubusercontent.com/leads-su/consul-config-manager/main/docs/images/service.png)

Application will also provide information on its startup with all information, which will be available in the "Meta" section of Consul UI.  
![Application Meta](https://raw.githubusercontent.com/leads-su/consul-config-manager/main/docs/images/meta.png)

## KeyValue Watcher

This watcher is able to detect changes made to your Consul installation, and then act accordingly.  
Whenever change is detected, CCM will pull this information to local environment files and new configuration value will be available in a matter of seconds for you to use.

### Example of supported data structures
CCM requires you to provide KeyValue values in the following format.  
This is required, so the CCM itself, as well as GUI could know what they are working with.

**1. Number**  
Simply tells CCM to convert this value to a number (instead of string)
```json
{"type":"number","value":123456}
```
**2. String**  
Most basic type, all of the values by default are treated as strings
```json
{"type":"string","value":"database.localhost"}
```
**3. Array**  
Allows CCM to build an array of values
```json
{"type":"array","value":["123", 456, "789"]}
```
**4. Reference**  
This is a special type, it allows you to reference existing value, which will then be converted by the CCM to the real value upon receiving changed key
```json
{"type":"reference","value":"shared/database/mysql/username"}
```

## Task Runner
CCM could also act as a task runner on the host it is installed on.  

**For example:**
1. You have 10 API servers
2. You need to change the database address on all of them
3. You need to write an Ansible role, or a pipeline to do so
4. You also need to modify configuration values (in some cases)
5. You need to keep 10 terminals open at a time (or just SSH to 10 servers one after another)

With use of CCM and the GUI provided by us, you are able to create a single pipeline, which will watch for key changes and apply any necessary commands whenever you decide to change the database address.

CCM can execute commands as itself (ccm), root (root), or any other user you specify.

## Event Streaming Server
In order to provide realtime output for the Task Runner, CCM utilizes SSE (Server Sent Events).   
This allows to avoid hustle with WebSockets, as well as provides ability to store logs locally (and access them later through HTTP), as well as stream them to any other service.

![Event Streaming Server](https://raw.githubusercontent.com/leads-su/consul-config-manager/main/docs/images/realtime_log.png)

## Automatic Consul Server switching
CCM is able to be configured with multiple servers in mind.  
That means that in case there is a problem with one of the servers, CCM will switch to another one.  
Also, upon initial connection, all servers will be pinged and server with lowest latency will be used.

# Initila Setup

By default, CCM will use `/etc/ccm.d` as its configuration folder.  
It will also try to load `config.yml` from this directory as its default configuration source.  

**IMPORTANT** - file extension must be exactly `yml` and not `yaml`, otherwise it will start with the default configuration (which you might not want).

## Starting application

If you wish to supply different configuration folder or configuration file name, you can do so with usage of flags.

**Change configuration directory:**
```bash
--config-path=/etc/ccm.d
```

**Change configuration file name:**
```bash
--config-file=config.yml
```

**Start application:**
```bash
ccm start --config-path=/etc/ccm.d --config-file=config.yml
```


# Example Configuration
```yaml
agent:                                 # Agent Configuration
  network:                             # Agent Network Configuration
    interface: ""                      # Interfaces which will be used to obtain IP address (if "address" is empty)
    address: "127.0.0.1"               # Manually set address which will be visible in Consul for this CCM instance
    port: 32175                        # Port which will be used to serve metrics + SSE events
  health_check:                        # Agent Health Checks configuration
    ttl: true                          # Enable TTL healthcheck
    http: true                         # Enable HTTP healthcheck
consul:                                # Consul Configuration
  enabled: true                        # Enable / Disable Consul service
  datacenter: "dc0"                    # Datacenter Name
  addresses:                           # List of Consul Servers (can be many)
    - scheme: "http"                   # Scheme to be used to access Consul API
      host: "consul1.local"            # Hostname of the Consul server
      port: 8500                       # Port of the Consul server
  token: "consul-acl-access-token"     # Access Token used to access Consul server
  write_to: "/etc/ccm.d"               # Path, where configuration files will be written
environment: "production"              # Application Environment
log:                                   # Log Configuration
  level: INFO                          # Default log level
  write_to: "/var/log/ccm"             # Where to write logs to
sse:                                   # Server Sent Events configuration
  write_to: "/var/log/ccm/events"      # Where to write execution logs
notifier:                              # Notifier configuration
  enabled: True                        # Enable / Disable notifier
  notify_on:                           # Enable / Disable notifications by type
    success: false                     # Enable / Disable "success" notifications
    error: true                        # Enable / Disable "error" notifications
  notifiers:                           # Notifiers configuration
    telegram:                          # Telegram notifier configuration
      enabled: True                    # Enable / Disable Telegram notifier
      token: "telegram-token"          # Telegram access token
      recipients:                      # Telegram recipients (who will receive notifications)
        - 123456789
        - 987654321
        - 123459876
```