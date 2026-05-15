# Scrape layer

Publishes raw `scrapev1` envelopes to NATS `scrape.>`.

- **Worker:** [scrape_worker/](scrape_worker/)
- **Sources:** [sources/](sources/) (ti, vuln, lola, ds, sbom, coderules, nuclei)
- **Contract:** [contract/scrapev1/](contract/scrapev1/)
- **Build:** `cd scrape && go build ./scrape_worker/...`
- **Deploy:** [../deploy/scrape/compose.yml](../deploy/scrape/compose.yml)
