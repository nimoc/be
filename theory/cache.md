---
permalink: /theory/cache/
---
# cache

## mysql和redis缓存实践<a id="mysql-redis"></a>

```
app #read/write# -> database(disk)
app #write# -> database(disk) -> sync redis(memory)
app #read# -> redis(memory)
```
