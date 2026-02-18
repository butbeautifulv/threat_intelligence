# Threat Intelligence

## Stage 1 :
Make tracing project with beta level parsing from one source of each category.

```
Vulns - Metasploit, Eploit-DB, NVD (CVSS, CPE, CWE),Vulners, API

LOLA - Lolbins/LOLScripts: LOLBAS, GTFOBins, LOFTS, TTP MITRE ATT&CK

DS (Detection & Simulation) - Redhat: Sigma rules, YARA rules, Atomic Red Team Tests, Caldera Profiles

TI - ТІ-Artifacts: ІОС (JP, URL, hash), IOA, Campaings / Clustrers
```

Save to Mongo for now

## Stage 2 :

Add kafka to increase fault tolerance. Add workers that listen to topic raw_data and store it to some kind of repository with replication.

Cue over go for schema validation?

 Make index layer database over raw freshly updated threats. Harden the code (SAST, SCA/SBOM) and maybe make Harbor repo to deliver safe packages for agent. 

## Stage 3 :

Create OpenClaw? or Picobot instance in the isolated field as an autoadjusting brain over it to later correlate it with SIEM and DLP systems.