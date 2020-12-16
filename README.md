# ESKDump

Allows dumping for ElasticSearch servers behind Kibana > 5 instances.

## Deprecated

Check [ES Toolkit](https://github.com/LeakIX/estk) instead

## Usage

```shell script
$ eskdump "http://127.0.0.1:5601" "member" 5000 > members
2020/09/29 14:34:58 ESKDump starting...
2020/09/29 14:34:58 Kibana endpoint : http://127.0.0.1:5601
2020/09/29 14:34:58 Index : member
2020/09/29 14:35:00 Got scrollId : FGluY2x1ZGVfY29udGV4dF91dWlkDXF1ZXJ5QW5kRmV0Y2gBFHhnUGIyWFFCTGtXeG54YUEzQ1ZFAAAAAAAHa54WTElweWl0SHhRanFoODA5TUNMeFRfdw==
2020/09/29 14:35:00 Dumping 4689639 documents to stdout :
Docs   0% |                               | (8482/4689639, 2432 it/s) [2s:32m4s]
```
