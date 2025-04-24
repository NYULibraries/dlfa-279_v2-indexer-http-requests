# dlfa-279_v2-indexer-http-requests

Jira: [v2 (EDIP) indexer: set up GitHub repo containing all Solr HTTP requests for a full indexing job](https://jira.nyu.edu/browse/DLFA-279)

Setup:

```bash
git clone git@github.com:NYULibraries/dlfa-279_set-up-github-repo-containing-all-v2-indexer-http-requests-xml.git
cd dlfa-279_set-up-github-repo-containing-all-v2-indexer-http-requests-xml/
go run main.go [PATH TO EAD FILES REPO]
```

Outputs:

* _http-requests-xml/prettified/_: the raw XML processed by `ead/eadutil.PrettifySolrAddMessageXML()`
* _http-requests-xml/raw/_: the raw XML

-----

# Versions of repos used for current diffs 

* go-ead-indexer: [see go.mod file]
* findingaids_eads_v2: [8d1b8fb6bd45327e90857c77bff8afa66358f4e7](https://github.com/NYULibraries/findingaids_eads_v2/tree/8d1b8fb6bd45327e90857c77bff8afa66358f4e7)
