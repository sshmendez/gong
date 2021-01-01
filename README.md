# GONG

A Go Multiplex Server Framework

# What

Create files servers and reverse proxies from a straight forward json configuration. No more DSL's

# How

GONG config has two top level parameters: `port` and `hosts`. `port` tells GONG where to serve the mux and `hosts` is a list of *host configurations*

A host config has four parameters: `hostname`, `path`, `type`, `config`


## hostname, types

`hostname` tells GONG which hostname to watch and `type` defines how GONG will direct those requests

### types 

- Reverse Proxy


types have associated `config` sections


## config

**Reverse Proxy** has 2 primary config sections

- remote (hostname)
- port  

requests are then forwarded to a constructed uri





