# Engage Tool Install Coverage

Auto-generated for install coverage wave.

```bash
python3 scripts/engage/generate-tool-install-coverage.py
```

| Tool | Binary | Install Required | Ubuntu/Debian repo | Kali fallback | Upstream fallback | Runtime on this host |
|------|--------|------------------|---------------------|---------------|-------------------|----------------------|
| `advanced_payload_generation` | `advanced` | no (bridge/workflow) | n/a | n/a | n/a | n/a |
| `ai_generate_attack_suite` | `ai` | no (bridge/workflow) | n/a | n/a | n/a | n/a |
| `ai_generate_payload` | `ai` | no (bridge/workflow) | n/a | n/a | n/a | n/a |
| `ai_reconnaissance_workflow` | `ai` | no (bridge/workflow) | n/a | n/a | n/a | n/a |
| `ai_test_payload` | `ai` | no (bridge/workflow) | n/a | n/a | n/a | n/a |
| `ai_vulnerability_assessment` | `ai` | no (bridge/workflow) | n/a | n/a | n/a | n/a |
| `amass_scan` | `amass` | yes | missing | yes | yes | missing |
| `analyze_target_intelligence` | `analyze` | no (bridge/workflow) | n/a | n/a | n/a | n/a |
| `anew_data_processing` | `anew` | yes | missing | yes | no | missing |
| `angr_symbolic_execution` | `angr` | yes | missing | yes | no | missing |
| `api_fuzzer` | `api` | no (bridge/workflow) | n/a | n/a | n/a | n/a |
| `api_schema_analyzer` | `api` | no (bridge/workflow) | n/a | n/a | n/a | n/a |
| `arjun_parameter_discovery` | `arjun` | yes | missing | yes | no | missing |
| `arjun_scan` | `arjun` | yes | missing | yes | no | missing |
| `arp_scan_discovery` | `arp` | yes | missing | yes | no | missing |
| `autorecon_comprehensive` | `autorecon` | no (bridge/workflow) | n/a | n/a | n/a | n/a |
| `autorecon_scan` | `autorecon` | no (bridge/workflow) | n/a | n/a | n/a | n/a |
| `binwalk_analyze` | `binwalk` | yes | missing | yes | no | missing |
| `browser_agent_inspect` | `browser` | no (bridge/workflow) | n/a | n/a | n/a | n/a |
| `bugbounty_authentication_bypass_testing` | `bugbounty` | no (bridge/workflow) | n/a | n/a | n/a | n/a |
| `bugbounty_business_logic_testing` | `bugbounty` | no (bridge/workflow) | n/a | n/a | n/a | n/a |
| `bugbounty_comprehensive_assessment` | `bugbounty` | no (bridge/workflow) | n/a | n/a | n/a | n/a |
| `bugbounty_file_upload_testing` | `bugbounty` | no (bridge/workflow) | n/a | n/a | n/a | n/a |
| `bugbounty_osint_gathering` | `bugbounty` | no (bridge/workflow) | n/a | n/a | n/a | n/a |
| `bugbounty_reconnaissance_workflow` | `bugbounty` | no (bridge/workflow) | n/a | n/a | n/a | n/a |
| `bugbounty_vulnerability_hunting` | `bugbounty` | no (bridge/workflow) | n/a | n/a | n/a | n/a |
| `burpsuite_alternative_scan` | `burpsuite` | yes | missing | yes | no | missing |
| `burpsuite_scan` | `burpsuite` | yes | missing | yes | no | missing |
| `checkov_iac_scan` | `checkov` | no (bridge/workflow) | n/a | n/a | n/a | n/a |
| `checksec_analyze` | `checksec` | no (bridge/workflow) | n/a | n/a | n/a | n/a |
| `clair_vulnerability_scan` | `clair` | no (bridge/workflow) | n/a | n/a | n/a | n/a |
| `clear_cache` | `clear` | no (bridge/workflow) | n/a | n/a | n/a | ok |
| `cloudmapper_analysis` | `cloudmapper` | no (bridge/workflow) | n/a | n/a | n/a | n/a |
| `comprehensive_api_audit` | `comprehensive` | no (bridge/workflow) | n/a | n/a | n/a | n/a |
| `correlate_threat_intelligence` | `correlate` | yes | missing | yes | no | missing |
| `create_attack_chain_ai` | `create` | no (bridge/workflow) | n/a | n/a | n/a | n/a |
| `create_file` | `create` | no (bridge/workflow) | n/a | n/a | n/a | n/a |
| `create_scan_summary` | `create` | no (bridge/workflow) | n/a | n/a | n/a | n/a |
| `create_vulnerability_report` | `create` | no (bridge/workflow) | n/a | n/a | n/a | n/a |
| `ctf_auto_solve_challenge` | `api` | no (bridge/workflow) | n/a | n/a | n/a | n/a |
| `ctf_binary_analyzer` | `api` | no (bridge/workflow) | n/a | n/a | n/a | n/a |
| `ctf_create_challenge_workflow` | `api` | no (bridge/workflow) | n/a | n/a | n/a | n/a |
| `ctf_cryptography_solver` | `api` | no (bridge/workflow) | n/a | n/a | n/a | n/a |
| `ctf_forensics_analyzer` | `api` | no (bridge/workflow) | n/a | n/a | n/a | n/a |
| `ctf_suggest_tools` | `api` | no (bridge/workflow) | n/a | n/a | n/a | n/a |
| `ctf_team_strategy` | `api` | no (bridge/workflow) | n/a | n/a | n/a | n/a |
| `dalfox_xss_scan` | `dalfox` | yes | missing | yes | no | missing |
| `delete_file` | `delete` | yes | missing | yes | no | missing |
| `detect_technologies_ai` | `detect` | yes | missing | yes | no | missing |
| `dirb_scan` | `dirb` | yes | missing | yes | no | missing |
| `dirsearch_scan` | `dirsearch` | yes | missing | yes | no | missing |
| `discover_attack_chains` | `discover` | yes | missing | yes | no | missing |
| `display_system_metrics` | `display` | yes | missing | yes | no | missing |
| `dnsenum_scan` | `dnsenum` | yes | missing | yes | no | missing |
| `docker_bench_security_scan` | `docker` | yes | missing | yes | no | ok |
| `dotdotpwn_scan` | `dotdotpwn` | yes | missing | yes | no | missing |
| `enum4linux_ng_advanced` | `enum4linux` | yes | missing | yes | no | missing |
| `enum4linux_scan` | `enum4linux` | yes | missing | yes | no | missing |
| `error_handling_statistics` | `error` | yes | missing | yes | no | missing |
| `execute_command` | `execute` | no (bridge/workflow) | n/a | n/a | n/a | n/a |
| `execute_python_script` | `engage-python-exec` | yes | missing | yes | no | missing |
| `exiftool_extract` | `exiftool` | yes | missing | yes | no | missing |
| `falco_runtime_monitoring` | `falco` | yes | missing | yes | no | missing |
| `feroxbuster_scan` | `feroxbuster` | yes | missing | yes | yes | missing |
| `ffuf_scan` | `ffuf` | yes | ok | yes | yes | missing |
| `fierce_scan` | `fierce` | yes | missing | yes | no | missing |
| `foremost_carving` | `foremost` | yes | missing | yes | no | missing |
| `format_tool_output_visual` | `format` | yes | missing | yes | no | missing |
| `gau_discovery` | `gau` | yes | missing | yes | no | missing |
| `gdb_analyze` | `gdb` | yes | missing | yes | no | ok |
| `gdb_peda_debug` | `gdb` | yes | missing | yes | no | ok |
| `generate_exploit_from_cve` | `generate` | no (bridge/workflow) | n/a | n/a | n/a | n/a |
| `generate_payload` | `generate` | no (bridge/workflow) | n/a | n/a | n/a | n/a |
| `get_cache_stats` | `get` | no (bridge/workflow) | n/a | n/a | n/a | n/a |
| `get_live_dashboard` | `get` | no (bridge/workflow) | n/a | n/a | n/a | n/a |
| `get_process_dashboard` | `get` | no (bridge/workflow) | n/a | n/a | n/a | n/a |
| `get_process_status` | `get` | no (bridge/workflow) | n/a | n/a | n/a | n/a |
| `get_telemetry` | `get` | no (bridge/workflow) | n/a | n/a | n/a | n/a |
| `ghidra_analysis` | `ghidra` | yes | missing | yes | no | missing |
| `gobuster_scan` | `gobuster` | yes | ok | yes | yes | missing |
| `graphql_scanner` | `graphql` | yes | missing | yes | no | missing |
| `hakrawler_crawl` | `hakrawler` | yes | missing | yes | no | missing |
| `hashcat_crack` | `hashcat` | yes | missing | yes | no | missing |
| `hashpump_attack` | `hashpump` | yes | missing | yes | no | missing |
| `http_framework_test` | `http` | no (bridge/workflow) | n/a | n/a | n/a | n/a |
| `http_intruder` | `http` | no (bridge/workflow) | n/a | n/a | n/a | n/a |
| `http_repeater` | `http` | no (bridge/workflow) | n/a | n/a | n/a | n/a |
| `http_set_rules` | `http` | no (bridge/workflow) | n/a | n/a | n/a | n/a |
| `http_set_scope` | `http` | no (bridge/workflow) | n/a | n/a | n/a | n/a |
| `httpx_probe` | `httpx` | yes | missing | yes | yes | missing |
| `hydra_attack` | `hydra` | yes | ok | yes | no | missing |
| `install_python_package` | `engage-python-install` | yes | missing | yes | no | missing |
| `intelligent_smart_scan` | `intelligent` | yes | missing | yes | no | missing |
| `jaeles_vulnerability_scan` | `jaeles` | yes | missing | yes | no | missing |
| `john_crack` | `john` | yes | missing | yes | no | missing |
| `jwt_analyzer` | `jwt` | yes | missing | yes | no | missing |
| `katana_crawl` | `katana` | yes | missing | yes | no | missing |
| `kube_bench_cis` | `kube` | no (bridge/workflow) | n/a | n/a | n/a | n/a |
| `kube_hunter_scan` | `kube` | no (bridge/workflow) | n/a | n/a | n/a | n/a |
| `libc_database_lookup` | `libc` | yes | missing | yes | no | missing |
| `list_active_processes` | `list` | no (bridge/workflow) | n/a | n/a | n/a | n/a |
| `list_files` | `list` | no (bridge/workflow) | n/a | n/a | n/a | n/a |
| `masscan_high_speed` | `masscan` | yes | ok | yes | no | missing |
| `metasploit_run` | `metasploit` | yes | missing | yes | no | missing |
| `modify_file` | `modify` | yes | missing | yes | no | missing |
| `monitor_cve_feeds` | `monitor` | yes | missing | yes | no | missing |
| `msfvenom_generate` | `msfvenom` | yes | missing | yes | no | missing |
| `nbtscan_netbios` | `nbtscan` | yes | missing | yes | no | missing |
| `netexec_scan` | `netexec` | yes | missing | yes | no | missing |
| `nikto_scan` | `nikto` | yes | ok | yes | no | missing |
| `nmap_advanced_scan` | `nmap` | yes | ok | yes | no | ok |
| `nmap_scan` | `nmap` | yes | ok | yes | no | ok |
| `nuclei_scan` | `nuclei` | yes | missing | yes | yes | missing |
| `objdump_analyze` | `objdump` | yes | missing | yes | no | ok |
| `one_gadget_search` | `one` | yes | missing | yes | no | missing |
| `optimize_tool_parameters_ai` | `optimize` | yes | missing | yes | no | missing |
| `pacu_exploitation` | `pacu` | yes | missing | yes | no | missing |
| `paramspider_discovery` | `paramspider` | yes | missing | yes | no | missing |
| `paramspider_mining` | `paramspider` | yes | missing | yes | no | missing |
| `pause_process` | `pause` | yes | missing | yes | no | missing |
| `prowler_scan` | `prowler` | yes | missing | yes | no | missing |
| `pwninit_setup` | `pwninit` | yes | missing | yes | no | missing |
| `pwntools_exploit` | `pwntools` | yes | missing | yes | no | missing |
| `qsreplace_parameter_replacement` | `qsreplace` | yes | missing | yes | no | missing |
| `radare2_analyze` | `radare2` | yes | missing | yes | no | missing |
| `research_zero_day_opportunities` | `research` | yes | missing | yes | no | missing |
| `responder_credential_harvest` | `responder` | yes | missing | yes | no | missing |
| `resume_process` | `resume` | yes | missing | yes | no | missing |
| `ropgadget_search` | `ropgadget` | yes | missing | yes | no | missing |
| `ropper_gadget_search` | `ropper` | yes | missing | yes | no | missing |
| `rpcclient_enumeration` | `rpcclient` | yes | missing | yes | no | missing |
| `rustscan_fast_scan` | `rustscan` | yes | missing | yes | no | missing |
| `scout_suite_assessment` | `scout` | yes | missing | yes | no | missing |
| `select_optimal_tools_ai` | `select` | yes | missing | yes | no | missing |
| `server_health` | `server` | yes | missing | yes | no | missing |
| `smbmap_scan` | `smbmap` | yes | missing | yes | no | missing |
| `sqlmap_scan` | `sqlmap` | yes | ok | yes | no | missing |
| `steghide_analysis` | `steghide` | yes | missing | yes | no | missing |
| `strings_extract` | `strings` | yes | missing | yes | no | ok |
| `subfinder_scan` | `subfinder` | yes | missing | yes | yes | missing |
| `target_timeline_intelligence` | `api` | no (bridge/workflow) | n/a | n/a | n/a | n/a |
| `terminate_process` | `terminate` | yes | missing | yes | no | missing |
| `terrascan_iac_scan` | `terrascan` | yes | missing | yes | no | missing |
| `test_error_recovery` | `test` | yes | missing | yes | no | ok |
| `threat_hunting_assistant` | `threat` | yes | missing | yes | no | missing |
| `trivy_scan` | `trivy` | yes | missing | yes | no | missing |
| `uro_url_filtering` | `uro` | yes | missing | yes | no | missing |
| `volatility3_analyze` | `volatility3` | yes | missing | yes | no | missing |
| `volatility_analyze` | `volatility` | yes | missing | yes | no | missing |
| `vulnerability_intelligence_dashboard` | `vulnerability` | yes | missing | yes | no | missing |
| `wafw00f_scan` | `wafw00f` | yes | missing | yes | no | missing |
| `waybackurls_discovery` | `waybackurls` | yes | missing | yes | no | missing |
| `wfuzz_scan` | `wfuzz` | yes | missing | yes | no | missing |
| `wpscan_analyze` | `wpscan` | yes | missing | yes | no | missing |
| `x8_parameter_discovery` | `x8` | yes | missing | yes | no | missing |
| `xsser_scan` | `xsser` | yes | missing | yes | no | missing |
| `xxd_hexdump` | `xxd` | yes | missing | yes | no | ok |
| `zap_scan` | `zap` | yes | missing | yes | no | missing |

- Catalog tools: **158**
- Install-required tools: **104**
- Ubuntu/Debian repo-ok (install-required subset): **8**
- Runtime ready on this host (`ok` + `n/a`): **63/158**
