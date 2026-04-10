package rate_limiter

import "fmt"

var NoIpInShardError = fmt.Errorf("no ip in shard")
var UnableToGetIp = fmt.Errorf("unable to get ip")
var WrongServerIpAddress = fmt.Errorf("wrong server ip")
