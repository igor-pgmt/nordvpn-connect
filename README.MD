# nordvpn-connect

This script retrieves nordvpn servers and connects your pc through ipsec to the fastest server.

### Installation
```sh
git clone https://github.com/igor-pgmt/nordvpn-connect.git
cd nordvpn-connect/
go build
sudo chmod +x nordvpn-connect
sudo mv nordvpn-connect /usr/bin/
```

### Usage

##### You can run this program with next flags:
```sh
Usage of nordVPNconnect:

  -conn string
    	Please, specify connection command ("up", "down" or "nothing"). [string] (default: up (default "up")
  -reload
    	Please, specify reload flag to reload ipsec settings. [bool] (default: true) (default true)
  -restart
    	Please, specify restart flag to restart ipsec settings. [bool] (default: false)
  -retryTimeout uint
    	Please, specify timeout to retry to up connection in seconds. [uint64] (default: 3) (default 5)
  -saveToConf
    	Please, specify saveToConf flag to save the fastest server to ipsec configuration file. [bool] (default: true) (default true)
  -statsURL string
    	Please, specify API url with percentage statistics. [string] (default: https://nordvpn.com/api/server/stats) (default "https://nordvpn.com/api/server/stats")
  -v	Please, specify v flag to enable verbose mode. [bool] (default: false)
```

##### Examples:

Connect to the fastest server:
```sh
sudo nordvpn-connect
```

Verbose mode:
```sh
sudo nordvpn-connect
```

Restart ipsec and connect to the fastest server:
```sh
sudo nordvpn-connect -restart
```

Stop the connection:
```sh
sudo nordvpn-connect -conn down
```

Just get the fastest server without any actions (ipsec restart is disabled by default):
```sh
sudo nordvpn-connect -conn nothing -saveToConf=false -reload=false
```
